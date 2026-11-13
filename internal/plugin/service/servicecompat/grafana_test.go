package servicecompat_test

import (
	"maps"
	"testing"

	"github.com/aiven/go-client-codegen/handler/service"
)

func TestGrafana(t *testing.T) {
	t.Parallel()
	runCompat(t, grafanaFixture, []compatCase{
		{name: "config matches the state"},
		{
			name:     "config has no user config",
			config:   map[string]any{"grafana_user_config": nil},
			wantPlan: []string{"grafana_user_config", "state"},
		},
	})
}

func grafanaFixture() fixture {
	const (
		host = "my-grafana.example.com"
		uri  = "https://avnadmin:secret@my-grafana.example.com:443"
	)

	f := newFixture("grafana", 1, "my-grafana", host, 443)
	zeroDisk(&f)
	maps.Copy(f.state, map[string]any{
		"components":          []any{componentState("grafana", host, 443, "primary")},
		"grafana":             []any{map[string]any{"uris": []any{uri}}},
		"grafana_user_config": []any{grafanaUserConfigState(nil)},
		"service_uri":         uri,
	})
	f.config["grafana_user_config"] = []any{grafanaUserConfig(nil)}
	f.api.Components = []service.ComponentOut{componentAPI("grafana", host, 443, service.UsageTypePrimary)}
	f.api.ConnectionInfo = &service.ConnectionInfoOut{Grafana: []string{uri}}
	f.api.ServiceUri = uri
	f.api.ServiceUriParams = map[string]string{"host": host, "port": "443", "user": "avnadmin", "password": "secret"}
	f.api.UserConfig = map[string]any{"ip_filter": []any{"0.0.0.0/0", "::/0"}}
	return f
}

func grafanaUserConfig(options map[string]any) map[string]any {
	result := map[string]any{
		"ip_filter": []any{"0.0.0.0/0", "::/0"},
	}
	maps.Copy(result, options)
	return result
}

func grafanaUserConfigState(options map[string]any) map[string]any {
	result := map[string]any{
		"additional_backup_regions":        nil,
		"alerting_enabled":                 false,
		"alerting_error_or_timeout":        "",
		"alerting_max_annotations_to_keep": 0,
		"alerting_nodata_or_nullvalues":    "",
		"allow_embedding":                  false,
		"auth_azuread":                     []any{},
		"auth_basic_enabled":               false,
		"auth_generic_oauth":               []any{},
		"auth_github":                      []any{},
		"auth_gitlab":                      []any{},
		"auth_google":                      []any{},
		"auth_jwt":                         []any{},
		"auth_okta":                        []any{},
		"auth_saml":                        []any{},
		"custom_domain":                    "",
		"enable_ipv6":                      false,
		"ip_filter":                        []any{"0.0.0.0/0", "::/0"},
		"ip_filter_object":                 []any{},
		"ip_filter_string":                 nil,
		"private_access":                   []any{},
		"privatelink_access":               []any{},
		"public_access":                    []any{},
		"service_log":                      false,
		"smtp":                             []any{},
		"static_ips":                       false,
		"unified_alerting":                 []any{},
	}
	maps.Copy(result, options)
	return result
}
