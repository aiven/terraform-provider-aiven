package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/aiven/go-client-codegen/handler/staticip"
)

// TestPG also covers the logic all services share, e.g., the disks, static IPs, tags.
func TestPG(t *testing.T) {
	t.Parallel()
	runCompat(t, pgFixture, []compatCase{
		{
			name: "config matches the state",
		},
		{
			// SDKv2 stored the user config the API returned
			name:     "config has no user config",
			config:   map[string]any{"pg_user_config": nil},
			wantPlan: []string{`pg_user_config: [1 items] -> [0 items]`, `state: "RUNNING" -> (known after apply)`},
		},
		{
			name: "config has no nested blocks",
			config: map[string]any{"pg_user_config": []any{map[string]any{
				"pg_version":  "16",
				"backup_hour": 3,
			}}},
			wantPlan: []string{`pg_user_config[0].pg: [1 items] -> [0 items]`, `pg_user_config[0].pglookout: [1 items] -> [0 items]`, `state: "RUNNING" -> (known after apply)`},

			// The API keeps the removed blocks
			wantApply: []string{`pg_user_config[0].pg: [1 items] -> [0 items]`, `pg_user_config[0].pglookout: [1 items] -> [0 items]`},
		},
		{
			// An SDKv2 update stored the IP filter in every option
			name: "ip_filter_object with nested blocks removed",
			state: map[string]any{"pg_user_config": []any{pgUserConfigState(map[string]any{
				"ip_filter":        []any{"10.0.0.0/8"},
				"ip_filter_object": []any{map[string]any{"description": "lab", "network": "10.0.0.0/8"}},
				"ip_filter_string": []any{"10.0.0.0/8"},
			})}},
			config: map[string]any{"pg_user_config": []any{map[string]any{
				"pg_version":       "16",
				"backup_hour":      3,
				"ip_filter_object": []any{map[string]any{"description": "lab", "network": "10.0.0.0/8"}},
			}}},
			api: func(s *service.ServiceGetOut) {
				s.UserConfig["ip_filter"] = []any{map[string]any{"description": "lab", "network": "10.0.0.0/8"}}
			},
			wantPlan:  []string{`pg_user_config[0].pg: [1 items] -> [0 items]`, `pg_user_config[0].pglookout: [1 items] -> [0 items]`, `state: "RUNNING" -> (known after apply)`},
			wantApply: []string{`pg_user_config[0].pg: [1 items] -> [0 items]`, `pg_user_config[0].pglookout: [1 items] -> [0 items]`},
		},
		{
			// A refresh stored it
			name:  "static_ips refreshed to empty",
			state: map[string]any{"static_ips": []any{}},
		},
		{
			// Older versions stored it for no additional disk
			name:  "additional_disk_space 0B",
			state: map[string]any{"additional_disk_space": "0B"},
		},
		{
			// SDKv2 stored it after disk_space was removed from the config
			name:  "disk_space empty",
			state: map[string]any{"disk_space": ""},
		},
		{
			name:   "additional_disk_space",
			state:  map[string]any{"additional_disk_space": "30GiB", "disk_space_used": "110GiB"},
			config: map[string]any{"additional_disk_space": "30GiB"},
			api:    func(s *service.ServiceGetOut) { s.DiskSpaceMb = new(112640) },
		},
		{
			name:   "project_vpc_id",
			state:  map[string]any{"project_vpc_id": testProject + "/" + testVPC},
			config: map[string]any{"project_vpc_id": testProject + "/" + testVPC},
			api:    func(s *service.ServiceGetOut) { s.ProjectVpcId = testVPC },
		},
		{
			name: "static_ips",
			state: map[string]any{
				"static_ips":     []any{"ip-1"},
				"pg_user_config": []any{pgUserConfigState(map[string]any{"static_ips": true})},
			},
			config: map[string]any{
				"static_ips":     []any{"ip-1"},
				"pg_user_config": []any{pgUserConfig(map[string]any{"static_ips": true})},
			},
			api: func(s *service.ServiceGetOut) { s.UserConfig["static_ips"] = true },
			ips: []staticip.StaticIPOut{{
				CloudName:         testCloud,
				ServiceName:       "my-pg",
				State:             staticip.StaticIPStateTypeAssigned,
				StaticIPAddressId: "ip-1",
			}},
		},
		{
			name:   "tag",
			state:  map[string]any{"tag": []any{map[string]any{"key": "env", "value": "dev"}}},
			config: map[string]any{"tag": []any{map[string]any{"key": "env", "value": "dev"}}},
			api:    func(s *service.ServiceGetOut) { s.Tags = map[string]string{"env": "dev"} },
		},
		{
			name:   "tech_emails",
			state:  map[string]any{"tech_emails": []any{map[string]any{"email": "me@example.com"}}},
			config: map[string]any{"tech_emails": []any{map[string]any{"email": "me@example.com"}}},
			api:    func(s *service.ServiceGetOut) { s.TechEmails = []service.TechEmailOut{{Email: "me@example.com"}} },
		},
		{
			name:   "termination_protection",
			state:  map[string]any{"termination_protection": true},
			config: map[string]any{"termination_protection": true},
			api:    func(s *service.ServiceGetOut) { s.TerminationProtection = true },
		},
		{
			name: "read replica",
			state: map[string]any{"service_integrations": []any{map[string]any{
				"integration_type":    "read_replica",
				"source_service_name": "my-source",
			}}},
			config: map[string]any{"service_integrations": []any{map[string]any{
				"integration_type":    "read_replica",
				"source_service_name": "my-source",
			}}},
			api: func(s *service.ServiceGetOut) {
				s.ServiceIntegrations = []service.ServiceIntegrationOut{{
					Active:          true,
					DestService:     new("my-pg"),
					Enabled:         true,
					IntegrationType: service.IntegrationTypeReadReplica,
					SourceService:   "my-source",
				}}
			},
		},
		{
			name:     "tag update",
			config:   map[string]any{"tag": []any{map[string]any{"key": "env", "value": "prod"}}},
			wantPlan: []string{`tag: [0 items] -> [1 items]`, `state: "RUNNING" -> (known after apply)`},
		},
		{
			// The hosts and the credentials stay known, as in SDKv2: other resources and providers use them
			name: "user config update",
			config: map[string]any{
				"pg_user_config": []any{pgUserConfig(map[string]any{"backup_hour": 4})},
			},
			wantPlan: []string{
				`components: [1 items] -> (known after apply)`,
				`pg: [1 items] -> (known after apply)`,
				`pg_user_config[0].backup_hour: 3 -> 4`,
				`state: "RUNNING" -> (known after apply)`,
			},
		},
		{
			// It adds a route, which can change the hosts
			name: "public access update",
			config: map[string]any{
				"pg_user_config": []any{pgUserConfig(map[string]any{
					"public_access": []any{map[string]any{"pg": true}},
				})},
			},
			wantPlan: []string{
				`additional_disk_space: "" -> (known after apply)`,
				`components: [1 items] -> (known after apply)`,
				`disk_space_cap: "400GiB" -> (known after apply)`,
				`disk_space_default: "80GiB" -> (known after apply)`,
				`disk_space_step: "10GiB" -> (known after apply)`,
				`disk_space_used: "80GiB" -> (known after apply)`,
				`maintenance_window_enabled: true -> (known after apply)`,
				`pg: [1 items] -> (known after apply)`,
				`state: "RUNNING" -> (known after apply)`,
				`pg_user_config[0].public_access: [0 items] -> [1 items]`,
				`service_host: "my-pg.example.com" -> (known after apply)`,
				`service_password: "secret" -> (known after apply)`,
				`service_port: 12345 -> (known after apply)`,
				`service_uri: "postgres://avnadmin:secret@my-pg.example.com:12345/defaultdb?sslmode=require" -> (known after apply)`,
				`service_username: "avnadmin" -> (known after apply)`,
			},
		},
		{
			// It shrinks the disk to the plan default, as SDKv2 did
			name:  "additional_disk_space removed",
			state: map[string]any{"additional_disk_space": "30GiB", "disk_space_used": "110GiB"},
			api:   func(s *service.ServiceGetOut) { s.DiskSpaceMb = new(112640) },
			wantPlan: []string{
				`additional_disk_space: "30GiB" -> "0B"`,
				`components: [1 items] -> (known after apply)`,
				`disk_space_cap: "400GiB" -> (known after apply)`,
				`disk_space_default: "80GiB" -> (known after apply)`,
				`disk_space_step: "10GiB" -> (known after apply)`,
				`disk_space_used: "110GiB" -> (known after apply)`,
				`maintenance_window_enabled: true -> (known after apply)`,
				`pg: [1 items] -> (known after apply)`,
				`state: "RUNNING" -> (known after apply)`,
			},
		},
		{
			// The read returns them in any order, which isn't a change
			name: "static_ips in another order with a plan update",
			state: map[string]any{
				"static_ips":     []any{"ip-1", "ip-2"},
				"pg_user_config": []any{pgUserConfigState(map[string]any{"static_ips": true})},
			},
			config: map[string]any{
				"plan":           "business-4",
				"static_ips":     []any{"ip-2", "ip-1"},
				"pg_user_config": []any{pgUserConfig(map[string]any{"static_ips": true})},
			},
			api: func(s *service.ServiceGetOut) { s.UserConfig["static_ips"] = true },
			ips: []staticip.StaticIPOut{
				{CloudName: testCloud, ServiceName: "my-pg", State: staticip.StaticIPStateTypeAssigned, StaticIPAddressId: "ip-1"},
				{CloudName: testCloud, ServiceName: "my-pg", State: staticip.StaticIPStateTypeAssigned, StaticIPAddressId: "ip-2"},
			},
			wantPlan: []string{
				`additional_disk_space: "" -> (known after apply)`,
				`plan: "startup-4" -> "business-4"`,
				`components: [1 items] -> (known after apply)`,
				`disk_space_cap: "400GiB" -> (known after apply)`,
				`disk_space_default: "80GiB" -> (known after apply)`,
				`disk_space_step: "10GiB" -> (known after apply)`,
				`disk_space_used: "80GiB" -> (known after apply)`,
				`maintenance_window_enabled: true -> (known after apply)`,
				`pg: [1 items] -> (known after apply)`,
				`state: "RUNNING" -> (known after apply)`,
			},
		},
		{
			name:        "tech_emails removed outside Terraform",
			state:       map[string]any{"tech_emails": []any{map[string]any{"email": "me@example.com"}}},
			config:      map[string]any{"tech_emails": []any{map[string]any{"email": "me@example.com"}}},
			wantRefresh: []string{`tech_emails: [1 items] -> [0 items]`},
			wantPlan:    []string{`state: "RUNNING" -> (known after apply)`, `tech_emails: [0 items] -> [1 items]`},
		},
		{
			// It changes the disks and the components
			name:   "plan update",
			config: map[string]any{"plan": "business-4"},
			wantPlan: []string{
				`additional_disk_space: "" -> (known after apply)`,
				`plan: "startup-4" -> "business-4"`,
				`components: [1 items] -> (known after apply)`,
				`disk_space_cap: "400GiB" -> (known after apply)`,
				`disk_space_default: "80GiB" -> (known after apply)`,
				`disk_space_step: "10GiB" -> (known after apply)`,
				`disk_space_used: "80GiB" -> (known after apply)`,
				`maintenance_window_enabled: true -> (known after apply)`,
				`pg: [1 items] -> (known after apply)`,
				`state: "RUNNING" -> (known after apply)`,
			},
		},
	})
}

const testVPC = "00000000-0000-0000-0000-000000000002"

func pgFixture() fixture {
	const (
		host = "my-pg.example.com"
		uri  = "postgres://avnadmin:secret@my-pg.example.com:12345/defaultdb?sslmode=require"
	)

	f := newFixture("pg", 1, "my-pg", host, 12345)
	maps.Copy(f.state, map[string]any{
		"components":     []any{componentState("pg", host, 12345, "primary")},
		"pg_user_config": []any{pgUserConfigState(nil)},
		"service_uri":    uri,
		"pg": []any{map[string]any{
			"bouncer": "", "dbname": "defaultdb", "host": host, "max_connections": 100,
			"params": []any{map[string]any{
				"database_name": "defaultdb", "host": host, "password": "secret",
				"port": 12345, "sslmode": "require", "user": "avnadmin",
			}},
			"password": "secret", "port": 12345, "replica_uri": "", "sslmode": "require",
			"standby_uris": []any{}, "syncing_uris": []any{}, "uri": uri, "uris": []any{uri}, "user": "avnadmin",
		}},
	})
	f.config["pg_user_config"] = []any{pgUserConfig(nil)}
	f.api.Components = []service.ComponentOut{componentAPI("pg", host, 12345, service.UsageTypePrimary)}
	f.api.ConnectionInfo = &service.ConnectionInfoOut{
		PG: []string{uri},
		PGParams: []service.PGParamOut{{
			Dbname: "defaultdb", Host: host, Password: "secret", Port: "12345", Sslmode: "require", User: "avnadmin",
		}},
	}
	f.api.Metadata = map[string]any{"max_connections": 100}
	f.api.ServiceUri = uri
	f.api.ServiceUriParams = map[string]string{
		"dbname": "defaultdb", "host": host, "password": "secret", "port": "12345", "sslmode": "require", "user": "avnadmin",
	}
	f.api.UserConfig = map[string]any{
		"backup_hour": 3, "backup_minute": 15,
		"ip_filter": []any{"0.0.0.0/0", "::/0"},
		"pg": map[string]any{
			"io_combine_limit": 16, "io_max_concurrency": -1, "io_method": "worker", "password_encryption": "scram-sha-256",
		},
		"pg_version": "16",
		"pglookout":  map[string]any{"max_failover_replication_time_lag": 60},
	}
	return f
}

// pgUserConfig is the config that has every block SDKv2 stored.
func pgUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"pg_version": "16",
		"pg":         []any{map[string]any{"password_encryption": "scram-sha-256"}},
		"pglookout":  []any{map[string]any{"max_failover_replication_time_lag": 60}},
	}
	maps.Copy(result, options)
	return result
}

// pgUserConfigState is the user config SDKv2 stored: the API values, and zero values for the rest.
func pgUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"admin_password":        "",
		"admin_username":        "",
		"backup_hour":           3,
		"backup_interval_hours": 0,
		"backup_minute":         15,
		"enable_ipv6":           false,
		"ip_filter":             []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object":      []any{},
		"ip_filter_string":      nil,
		"migration":             []any{},
		"pg": []any{map[string]any{
			"autovacuum_max_workers": 0,
			"io_combine_limit":       16,
			"io_max_concurrency":     -1,
			"io_method":              "worker",
			"jit":                    false,
			"log_line_prefix":        "",
			"password_encryption":    "scram-sha-256",
		}},
		"pg_read_replica": false,
		"pg_version":      "16",
		"pgbouncer":       []any{},
		"pglookout":       []any{map[string]any{"max_failover_replication_time_lag": 60}},
		"private_access":  []any{},
		"public_access":   []any{},
		"service_log":     false,
		"static_ips":      false,
		"timescaledb":     []any{},
		"variant":         "",
		"work_mem":        0,
	}
	maps.Copy(result, options)
	return result
}
