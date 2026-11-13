package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/service"
)

func TestFlink(t *testing.T) {
	t.Parallel()
	runCompat(t, flinkFixture, []compatCase{
		{name: "config matches the state"},
		{
			name:     "config has no user config",
			config:   map[string]any{"flink_user_config": nil},
			wantPlan: []string{"flink_user_config", "state"},
		},
		{
			name: "config has no ip_filter",
			config: map[string]any{"flink_user_config": []any{map[string]any{
				"flink_version":        "1.19",
				"number_of_task_slots": 10,
			}}},
		},
	})
}

func flinkFixture() fixture {
	const (
		host      = "my-flink.example.com"
		uri       = "https://avnadmin:secret@my-flink.example.com"
		hostPort1 = "my-flink.example.com:443"
		hostPort2 = "my-flink.example.com:8081"
	)

	f := newFixture("flink", 1, "my-flink", host, 0)
	zeroDisk(&f)
	maps.Copy(f.state, map[string]any{
		"components":        []any{componentState("flink", host, 443, "primary")},
		"flink":             []any{map[string]any{"host_ports": []any{hostPort1, hostPort2}}},
		"flink_user_config": []any{flinkUserConfigState(nil)},
		"service_uri":       uri,
	})
	f.config["flink_user_config"] = []any{flinkUserConfig(nil)}
	f.api.Components = []service.ComponentOut{componentAPI("flink", host, 443, service.UsageTypePrimary)}
	f.api.ConnectionInfo = &service.ConnectionInfoOut{Flink: []string{hostPort1, hostPort2}}
	f.api.ServiceUri = uri
	f.api.ServiceUriParams = map[string]string{"host": host, "port": "0", "user": "avnadmin", "password": "secret"}
	f.api.UserConfig = map[string]any{
		"flink_version": "1.19", "ip_filter": []any{"0.0.0.0/0", "::/0"}, "number_of_task_slots": 10,
	}
	return f
}

func flinkUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"flink_version":        "1.19",
		"ip_filter":            []any{"0.0.0.0/0", "::/0"},
		"number_of_task_slots": 10,
	}
	maps.Copy(result, options)
	return result
}

func flinkUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"additional_backup_regions": nil,
		"custom_code":               false,
		"flink_version":             "1.19",
		"ip_filter":                 []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object":          []any{},
		"ip_filter_string":          nil,
		"number_of_task_slots":      10,
		"pekko_ask_timeout_s":       0,
		"pekko_framesize_b":         0,
		"privatelink_access":        []any{},
		"public_access":             []any{},
		"service_log":               false,
		"static_ips":                false,
	}
	maps.Copy(result, options)
	return result
}
