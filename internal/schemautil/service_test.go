package schemautil

import (
	"testing"

	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/stretchr/testify/assert"
)

func TestContainsRedactedCreds(t *testing.T) {
	cases := []struct {
		name     string
		hash     map[string]any
		expected error
	}{
		{
			name:     "contains redacted",
			hash:     map[string]any{"password": "<redacted>"},
			expected: errContainsRedactedCreds,
		},
		{
			name:     "contains invalid redacted",
			hash:     map[string]any{"password": "<REDACTED>"},
			expected: nil,
		},
		{
			name:     "does not contain redacted",
			hash:     map[string]any{"password": "123"},
			expected: nil,
		},
	}

	for _, opt := range cases {
		t.Run(opt.name, func(t *testing.T) {
			err := ContainsRedactedCreds(opt.hash)
			assert.Equal(t, err, opt.expected)
		})
	}
}

func TestFlattenServiceComponents(t *testing.T) {
	plConnID := "plc58e4782ab22"
	sslTrue := true

	tests := []struct {
		name     string
		input    *service.ServiceGetOut
		expected []map[string]interface{}
	}{
		{
			name: "component with privatelink_connection_id",
			input: &service.ServiceGetOut{
				Components: []service.ComponentOut{
					{
						Component:               "pg",
						Host:                    "privatelink-1-pg1-test.aivencloud.com",
						Port:                    12656,
						Route:                   service.RouteTypePrivatelink,
						Usage:                   service.UsageTypePrimary,
						Ssl:                     &sslTrue,
						PrivatelinkConnectionId: &plConnID,
					},
				},
			},
			expected: []map[string]interface{}{
				{
					"component":                   "pg",
					"host":                        "privatelink-1-pg1-test.aivencloud.com",
					"port":                        12656,
					"connection_uri":              "privatelink-1-pg1-test.aivencloud.com:12656",
					"route":                       service.RouteTypePrivatelink,
					"usage":                       service.UsageTypePrimary,
					"ssl":                         true,
					"kafka_authentication_method": service.KafkaAuthenticationMethodType(""),
					"privatelink_connection_id":   "plc58e4782ab22",
				},
			},
		},
		{
			name: "component without privatelink_connection_id",
			input: &service.ServiceGetOut{
				Components: []service.ComponentOut{
					{
						Component: "pg",
						Host:      "pg1-test.aivencloud.com",
						Port:      12656,
						Route:     service.RouteTypeDynamic,
						Usage:     service.UsageTypePrimary,
						Ssl:       &sslTrue,
					},
				},
			},
			expected: []map[string]interface{}{
				{
					"component":                   "pg",
					"host":                        "pg1-test.aivencloud.com",
					"port":                        12656,
					"connection_uri":              "pg1-test.aivencloud.com:12656",
					"route":                       service.RouteTypeDynamic,
					"usage":                       service.UsageTypePrimary,
					"ssl":                         true,
					"kafka_authentication_method": service.KafkaAuthenticationMethodType(""),
					"privatelink_connection_id":   "",
				},
			},
		},
		{
			name: "multiple components with mixed privatelink",
			input: &service.ServiceGetOut{
				Components: []service.ComponentOut{
					{
						Component: "pg",
						Host:      "pg1-test.aivencloud.com",
						Port:      12656,
						Route:     service.RouteTypeDynamic,
						Usage:     service.UsageTypePrimary,
						Ssl:       &sslTrue,
					},
					{
						Component:               "pg",
						Host:                    "privatelink-1-pg1-test.aivencloud.com",
						Port:                    12656,
						Route:                   service.RouteTypePrivatelink,
						Usage:                   service.UsageTypePrimary,
						Ssl:                     &sslTrue,
						PrivatelinkConnectionId: &plConnID,
					},
				},
			},
			expected: []map[string]interface{}{
				{
					"component":                   "pg",
					"host":                        "pg1-test.aivencloud.com",
					"port":                        12656,
					"connection_uri":              "pg1-test.aivencloud.com:12656",
					"route":                       service.RouteTypeDynamic,
					"usage":                       service.UsageTypePrimary,
					"ssl":                         true,
					"kafka_authentication_method": service.KafkaAuthenticationMethodType(""),
					"privatelink_connection_id":   "",
				},
				{
					"component":                   "pg",
					"host":                        "privatelink-1-pg1-test.aivencloud.com",
					"port":                        12656,
					"connection_uri":              "privatelink-1-pg1-test.aivencloud.com:12656",
					"route":                       service.RouteTypePrivatelink,
					"usage":                       service.UsageTypePrimary,
					"ssl":                         true,
					"kafka_authentication_method": service.KafkaAuthenticationMethodType(""),
					"privatelink_connection_id":   "plc58e4782ab22",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenServiceComponents(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
