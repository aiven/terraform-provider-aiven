package apiconvert

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// TestPropsReqs is a test for propsReqs.
func TestPropsReqs(t *testing.T) {
	type args struct {
		schemaType  userconfig.SchemaType
		serviceName string
	}

	tests := []struct {
		name string
		args args
		want struct {
			wantP map[string]any
			wantR map[string]struct{}
		}
	}{
		{
			name: "basic",
			args: args{
				schemaType:  userconfig.IntegrationEndpointTypes,
				serviceName: "rsyslog",
			},
			want: struct {
				wantP map[string]any
				wantR map[string]struct{}
			}{
				map[string]any{
					"ca": map[string]any{
						"example":    "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----\n",
						"max_length": 16384,
						"title":      "PEM encoded CA certificate",
						"type": []any{
							"string",
							"null",
						},
					},
					"cert": map[string]any{
						"example":    "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----\n",
						"max_length": 16384,
						"title":      "PEM encoded client certificate",
						"type": []any{
							"string",
							"null",
						},
					},
					"format": map[string]any{
						"default": "rfc5424",
						"enum": []any{
							map[string]any{"value": "rfc5424"},
							map[string]any{"value": "rfc3164"},
							map[string]any{"value": "custom"},
						},
						"example": "rfc5424",
						"title":   "message format",
						"type":    "string",
					},
					"key": map[string]any{
						"example":    "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
						"max_length": 16384,
						"title":      "PEM encoded client key",
						"type": []any{
							"string",
							"null",
						},
					},
					"logline": map[string]any{
						"example":    "<%pri%>%timestamp:::date-rfc3339% %HOSTNAME% %app-name% %msg%",
						"max_length": 512,
						"min_length": 1,
						"pattern":    "^[ -~\\t]+$",
						"title":      "custom syslog message format",
						"type":       "string",
					},
					"port": map[string]any{
						"default": "514",
						"example": "514",
						"maximum": 65535,
						"minimum": 1,
						"title":   "rsyslog server port",
						"type":    "integer",
					},
					"sd": map[string]any{
						"example":    "TOKEN tag=\"LiteralValue\"",
						"max_length": 1024,
						"title":      "Structured data block for log message",
						"type": []any{
							"string",
							"null",
						},
					},
					"server": map[string]any{
						"example":    "logs.example.com",
						"max_length": 255,
						"min_length": 4,
						"title":      "rsyslog server IP address or hostname",
						"type":       "string",
					},
					"tls": map[string]any{
						"default": true,
						"example": true,
						"title":   "Require TLS",
						"type":    "boolean",
					},
				},
				map[string]struct{}{
					"format": {},
					"port":   {},
					"server": {},
					"tls":    {},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotP, gotR, _ := propsReqs(tt.args.schemaType, tt.args.serviceName)

			if !cmp.Equal(gotP, tt.want.wantP) {
				t.Errorf(cmp.Diff(tt.want.wantP, gotP))
			}

			if !cmp.Equal(gotR, tt.want.wantR) {
				t.Errorf(cmp.Diff(tt.want.wantR, gotR))
			}
		})
	}
}
