package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// This is the test to demonstrate what we would like to achieve
func TestExpectedBehaviour(t *testing.T) {
	templateSet := InitializeTemplateStore(t) // all existing resources are pre-registered

	// you can add new resources to the template if needed
	templateSet.AddTemplate("aiven_kafka", `
resource "aiven_kafka" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "{{ .service_name }}"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}`)

	// you can add new external resources to the template
	templateSet.AddTemplate("gcp_provider", `
provider "google" {
	project = "{{ .gcp_project }}"
	region  = "{{ .gcp_region }}"
}`)

	builder := templateSet.NewBuilder().
		AddDataSource("aiven_project", map[string]any{ // provide configuration for the data source
			"resource_name": "foo",
			"project":       "test_project",
		}).
		AddResource("aiven_kafka_quota", map[string]any{
			"resource_name":      "example_quota_1", // add a resource that exists in the provider list
			"project":            "test_project",
			"service_name":       "test-service",
			"user":               "test_user",
			"client_id":          "test_client_id",
			"consumer_byte_rate": 1000,
		}).
		AddResource("aiven_kafka_quota", map[string]any{ // add a second resource with different resource name
			"resource_name":      "example_quota_2",
			"project":            Reference("data.aiven_project.foo.project"), // reference to the existing resource
			"service_name":       "test_service_2",
			"user":               "test_user_2",
			"client_id":          "test_client_id_2",
			"consumer_byte_rate": 1000,
		}).
		AddResource("aiven_kafka_quota", map[string]any{ // add a second resource with different resource name
			"resource_name":      "example_quota_3",
			"project":            Literal("demo_project"), // value is a literal
			"service_name":       "test_service_2",
			"user":               "test_user_2",
			"client_id":          "test_client_id_2",
			"consumer_byte_rate": 1000,
		}).
		AddTemplate("aiven_kafka", map[string]any{"service_name": "test-service"}).
		AddTemplate("gcp_provider", map[string]any{"gcp_project": "test-project", "gcp_region": "us-central1"})

	str := builder.MustRender(t) // render the template
	assert.NotEmpty(t, str)

	println(str)

	builder.Replace("aiven_kafka_quota.example_quota_2", map[string]any{
		"resource_name":      "example_quota_2",
		"project":            Reference("data.aiven_project.foo.project"),
		"service_name":       "test_service_2",
		"user":               "test_user_2",
		"client_id":          "test_client_id_2",
		"consumer_byte_rate": 900,
	}).Remove("aiven_kafka_quota.example_quota_3")

	str = builder.MustRender(t) // render the template
	assert.NotEmpty(t, str)

	println(str)

	nextStep := builder.MustRender(t) // render the template again
	assert.NotEmpty(t, nextStep)
}
