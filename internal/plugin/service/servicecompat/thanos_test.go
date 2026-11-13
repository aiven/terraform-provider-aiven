package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/service"
)

func TestThanos(t *testing.T) {
	t.Parallel()
	runCompat(t, thanosFixture, []compatCase{
		{
			name: "config matches the state",
		},
		{
			name:     "config has no user config",
			config:   map[string]any{"thanos_user_config": nil},
			wantPlan: []string{"thanos_user_config", "state"},
		},
		{
			name: "config has no nested blocks",
			config: map[string]any{"thanos_user_config": []any{map[string]any{
				"ip_filter": []any{"0.0.0.0/0", "::/0"},
			}}},
			wantPlan: []string{
				"state",
				"thanos_user_config[0].public_access",
				"thanos_user_config[0].query_frontend",
			},

			// The API keeps the removed blocks
			wantApply: []string{
				"thanos_user_config[0].public_access",
				"thanos_user_config[0].query_frontend",
			},
		},
		{
			// SDKv2 stored tech_emails as a list block; config uses the same shape
			name: "tech_emails in state",
			state: map[string]any{
				"tech_emails": []any{map[string]any{"email": "me@example.com"}},
			},
			config: map[string]any{
				"tech_emails": []any{map[string]any{"email": "me@example.com"}},
			},
			api: func(s *service.ServiceGetOut) {
				s.TechEmails = []service.TechEmailOut{{Email: "me@example.com"}}
			},
		},
	})
}

func thanosFixture() fixture {
	const (
		host   = "my-thanos.example.com"
		uri    = "https://avnadmin:secret@my-thanos.example.com:12345"
		query  = "https://my-thanos.example.com:12345/query"
		qfront = "https://my-thanos.example.com:12345/query-frontend"
		rw     = "https://my-thanos.example.com:12345/receiver/write"
	)

	f := newFixture("thanos", 0, "my-thanos", host, 12348)
	maps.Copy(f.state, map[string]any{
		"disk_space_cap": "1600GiB", "disk_space_default": "160GiB",
		"disk_space_step": "50GiB", "disk_space_used": "160GiB",
		"components": []any{
			componentState("query", host, 12345, "primary"),
			componentState("receiver", host, 12348, "primary"),
		},
		"thanos": []any{map[string]any{
			"query_frontend_uri": qfront, "query_uri": query,
			"receiver_remote_write_uri": rw, "uris": []any{uri},
		}},
		"thanos_user_config": []any{thanosUserConfigState(nil)},
		"service_uri":        uri,
	})
	f.config["thanos_user_config"] = []any{thanosUserConfig(nil)}
	f.api.Components = []service.ComponentOut{
		componentAPI("query", host, 12345, service.UsageTypePrimary),
		componentAPI("receiver", host, 12348, service.UsageTypePrimary),
	}
	f.api.ConnectionInfo = &service.ConnectionInfoOut{
		Thanos: []string{uri}, QueryUri: new(query), QueryFrontendUri: new(qfront), ReceiverRemoteWriteUri: new(rw),
	}
	f.api.DiskSpaceMb = new(163840)
	f.api.ServiceUri = uri
	f.api.ServiceUriParams = map[string]string{"host": host, "port": "12348", "user": "avnadmin", "password": "secret"}
	f.api.UserConfig = map[string]any{
		"ip_filter": []any{"0.0.0.0/0", "::/0"},
		"public_access": map[string]any{
			"query": true, "compactor": false, "query_frontend": false,
			"receiver_ingesting": false, "receiver_routing": false, "ruler": false, "store": false,
		},
		"query-frontend": map[string]any{"query-range.align-range-with-step": true},
	}
	f.specs.DiskSpaceMb = 163840
	f.specs.DiskSpaceCapMb = new(1638400)
	f.specs.DiskSpaceStepMb = new(51200)
	return f
}

func thanosUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"query_frontend": []any{map[string]any{"query_range_align_range_with_step": true}},
		"ip_filter":      []any{"0.0.0.0/0", "::/0"},
		"public_access": []any{map[string]any{
			"query": true, "compactor": false, "query_frontend": false,
			"receiver_ingesting": false, "receiver_routing": false, "ruler": false, "store": false,
		}},
	}
	maps.Copy(result, options)
	return result
}

func thanosUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"compactor":        []any{},
		"env":              nil,
		"ip_filter":        []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object": []any{},
		"ip_filter_string": nil,
		"object_storage_usage_alert_threshold_gb": 0,
		"private_access":     []any{},
		"privatelink_access": []any{},
		"public_access": []any{map[string]any{
			"compactor": false, "query": true, "query_frontend": false,
			"receiver_ingesting": false, "receiver_routing": false, "ruler": false, "store": false,
		}},
		"query": []any{},
		"query_frontend": []any{map[string]any{
			"query_range_align_range_with_step": true,
		}},
		"receiver_ingesting": nil,
		"receiver_routing":   nil,
		"ruler":              nil,
		"service_log":        false,
		"static_ips":         false,
		"store":              nil,
	}
	maps.Copy(result, options)
	return result
}
