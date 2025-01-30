package template

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/kafka"
)

func TestComplexSchema(t *testing.T) {
	t.Parallel()

	pgResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
			},
			"plan": {
				Type:     schema.TypeString,
				Required: true,
			},
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"postgresql": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "Values provided by the PostgreSQL server.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uri": {
							Type:      schema.TypeString,
							Optional:  true,
							Computed:  true,
							Sensitive: true,
						},
						"uris": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"params": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host": {
										Type:      schema.TypeString,
										Sensitive: true,
									},
									"port": {
										Type:      schema.TypeInt,
										Sensitive: true,
									},
									"user": {
										Type:      schema.TypeString,
										Sensitive: true,
									},
									"password": {
										Type:      schema.TypeString,
										Sensitive: true,
									},
									"database_name": {
										Type:      schema.TypeString,
										Sensitive: true,
									},
									"sslmode": {
										Type:      schema.TypeString,
										Sensitive: true,
									},
								},
							},
						},
						"max_connections": {
							Type:      schema.TypeInt,
							Computed:  true, // is computed and will be ignored in the output
							Sensitive: true,
						},
					},
				},
			},
		},
	}

	const expectedHCL = `resource "aiven_pg" "complex_pg" {
  plan = "business-4"
  project = "test-project"
  service_name = "pg-complex"
  postgresql {
    uri = "postgresql://user:pass@host:5432/db"
    uris = [
      "postgresql://user:pass@host1:5432/db",
      "postgresql://user:pass@host2:5432/db",
    ]
    params {
      user = "admin"
      password = "secret"
      database_name = "mydb"
      sslmode = "verify-full"
      host = "pg-host"
      port = 5432
    }
  }
}`

	set := NewSDKStore(t)
	set.registerResource("aiven_pg", pgResource, ResourceKindResource)

	config := map[string]any{
		"resource_name": "complex_pg",
		"project":       "test-project",
		"service_name":  "pg-complex",
		"plan":          "business-4",
		"postgresql": []map[string]any{{
			"uri": "postgresql://user:pass@host:5432/db",
			"uris": []string{
				"postgresql://user:pass@host1:5432/db",
				"postgresql://user:pass@host2:5432/db",
			},
			"params": []map[string]any{{
				"host":          "pg-host",
				"port":          5432,
				"user":          "admin",
				"password":      "secret",
				"database_name": "mydb",
				"sslmode":       "verify-full",
			}},
			"max_connections": 100, // is computed and will be ignored in the output
		}},
	}

	builder := set.NewBuilder()
	builder.AddResource("aiven_pg", config)

	result, err := builder.Render(t)
	require.NoError(t, err)

	assert.Equal(t,
		normalizeHCL(expectedHCL),
		normalizeHCL(result),
		"Generated HCL does not match expected output",
	)
}

func TestKafkaQuotaResource(t *testing.T) {
	t.Parallel()

	ts := InitializeTemplateStore(t)
	ts.registerResource(
		"aiven_kafka_quota",
		kafka.ResourceKafkaQuota(),
		ResourceKindResource,
	)

	b := ts.NewBuilder()

	b.AddResource("aiven_kafka_quota", map[string]any{
		"resource_name":      "test-kafka",
		"project":            "test-project",
		"service_name":       "test-kafka",
		"client_id":          "test-client-id",
		"user":               "test-user",
		"consumer_byte_rate": 123,
		"producer_byte_rate": 456,
		"request_percentage": 78,
	})

	expected := `resource "aiven_kafka_quota" "test-kafka" {
	  project = "test-project"
	  service_name = "test-kafka"
	  client_id = "test-client-id"
	  user = "test-user"
	  consumer_byte_rate = 123
      producer_byte_rate = 456
	  request_percentage = 78
	}`

	res := b.MustRender(t)

	assert.Equal(t, normalizeHCL(expected), normalizeHCL(res))
}

func TestPostgresResource(t *testing.T) {
	t.Parallel()

	templateSet := InitializeTemplateStore(t) // get template with all registered resources

	builder := templateSet.NewBuilder()

	builder.AddResource("aiven_pg", map[string]any{
		"resource_name":           "pg",
		"project":                 Reference("data.aiven_project.pr1.project"),
		"cloud_name":              "google-europe-west1",
		"plan":                    "startup-4",
		"service_name":            "my-pg1",
		"maintenance_window_dow":  "monday",
		"maintenance_window_time": "10:00:00",
		"static_ips": []any{
			Reference("aiven_static_ip.ips[0].static_ip_address_id"),
			Reference("aiven_static_ip.ips[1].static_ip_address_id"),
			Reference("aiven_static_ip.ips[2].static_ip_address_id"),
			Reference("aiven_static_ip.ips[3].static_ip_address_id"),
		},
		"pg_user_config": []map[string]any{{
			"pg_version": 11,
			"static_ips": true,
			"public_access": []map[string]any{{
				"pg":         true,
				"prometheus": false,
			}},
			"pg": []map[string]any{{
				"idle_in_transaction_session_timeout": 900,
				"log_min_duration_statement":          -1,
			}},
		}},
		"timeouts": map[string]any{
			"create": "20m",
			"update": "15m",
		},
	})

	expected := `resource "aiven_pg" "pg" {
  project                 = data.aiven_project.pr1.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "my-pg1"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  static_ips = [
    aiven_static_ip.ips[0].static_ip_address_id,
    aiven_static_ip.ips[1].static_ip_address_id,
    aiven_static_ip.ips[2].static_ip_address_id,
    aiven_static_ip.ips[3].static_ip_address_id,
  ]

  pg_user_config {
    pg_version = 11
    static_ips = true

    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
      log_min_duration_statement          = -1
    }
  }

  timeouts {
    create = "20m"
    update = "15m"
  }
}`

	result := builder.MustRender(t)
	assert.Equal(t, normalizeHCL(expected), normalizeHCL(result))
}
