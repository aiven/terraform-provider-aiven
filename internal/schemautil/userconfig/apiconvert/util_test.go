package apiconvert

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// TestPropsReqs is a test for propsReqs.
func TestPropsReqs(t *testing.T) {
	type args struct {
		st userconfig.SchemaType
		n  string
	}

	tests := []struct {
		name string
		args args
		want struct {
			wantP map[string]interface{}
			wantR map[string]struct{}
		}
	}{
		{
			name: "basic",
			args: args{
				st: userconfig.IntegrationEndpointTypes,
				n:  "rsyslog",
			},
			want: struct {
				wantP map[string]interface{}
				wantR map[string]struct{}
			}{
				map[string]interface{}{
					"ca": map[string]interface{}{
						"example":    "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----\n",
						"max_length": 16384,
						"title":      "PEM encoded CA certificate",
						"type": []interface{}{
							"string",
							"null",
						},
					},
					"cert": map[string]interface{}{
						"example":    "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----\n",
						"max_length": 16384,
						"title":      "PEM encoded client certificate",
						"type": []interface{}{
							"string",
							"null",
						},
					},
					"format": map[string]interface{}{
						"default": "rfc5424",
						"enum": []interface{}{
							map[string]interface{}{"value": "rfc5424"},
							map[string]interface{}{"value": "rfc3164"},
							map[string]interface{}{"value": "custom"},
						},
						"example": "rfc5424",
						"title":   "message format",
						"type":    "string",
					},
					"key": map[string]interface{}{
						"example":    "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
						"max_length": 16384,
						"title":      "PEM encoded client key",
						"type": []interface{}{
							"string",
							"null",
						},
					},
					"logline": map[string]interface{}{
						"example":    "<%pri%>%timestamp:::date-rfc3339% %HOSTNAME% %app-name% %msg%",
						"max_length": 512,
						"min_length": 1,
						"title":      "custom syslog message format",
						"type":       "string",
					},
					"port": map[string]interface{}{
						"default": "514",
						"example": "514",
						"maximum": 65535,
						"minimum": 1,
						"title":   "rsyslog server port",
						"type":    "integer",
					},
					"sd": map[string]interface{}{
						"example":    "TOKEN tag=\"LiteralValue\"",
						"max_length": 1024,
						"title":      "Structured data block for log message",
						"type": []interface{}{
							"string",
							"null",
						},
					},
					"server": map[string]interface{}{
						"example":    "logs.example.com",
						"max_length": 255,
						"min_length": 4,
						"title":      "rsyslog server IP address or hostname",
						"type":       "string",
					},
					"tls": map[string]interface{}{
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
			gotP, gotR, _ := propsReqs(tt.args.st, tt.args.n)

			if !cmp.Equal(gotP, tt.want.wantP) {
				t.Errorf(cmp.Diff(tt.want, gotP))
			}

			if !cmp.Equal(gotR, tt.want.wantR) {
				t.Errorf(cmp.Diff(tt.want, gotR))
			}
		})
	}
}
