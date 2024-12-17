package kafka_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aiven/go-client-codegen/handler/kafka"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const kafkaQuotaResource = "aiven_kafka_quota"

func TestAccAivenKafkaQuota(t *testing.T) {
	var (
		registry    = acc.NewTemplateRegistry(kafkaQuotaResource)
		randName    = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		serviceName = fmt.Sprintf("test-acc-sr-%s", randName)
		projectName = os.Getenv("AIVEN_PROJECT_NAME")

		user     = fmt.Sprintf("acc_test_user_%s", randName)
		clientID = fmt.Sprintf("acc_test_client_%s", randName)
	)

	// Add templates
	registry.MustAddTemplate(t, "project_data", `
data "aiven_project" "foo" {
  project = "{{ .project }}"
}`)

	registry.MustAddTemplate(t, "aiven_kafka", `
resource "aiven_kafka" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "{{ .service_name }}"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}`)

	registry.MustAddTemplate(t, "kafka_quota_full", `
resource "aiven_kafka_quota" "{{ .resource_name }}" {
  project             = data.aiven_project.foo.project
  service_name        = aiven_kafka.bar.service_name
  user                = "{{ .user }}"
  client_id           = "{{ .client_id }}"
  consumer_byte_rate  = {{ .consumer_byte_rate }}
  producer_byte_rate  = {{ .producer_byte_rate }}
  request_percentage  = {{ .request_percentage }}
}`)

	registry.MustAddTemplate(t, "kafka_quota_user", `
resource "aiven_kafka_quota" "{{ .resource_name }}" {
  project             = data.aiven_project.foo.project
  service_name        = aiven_kafka.bar.service_name
  user                = "{{ .user }}"
  request_percentage  = {{ .request_percentage }}
}`)

	registry.MustAddTemplate(t, "kafka_quota_client_id", `
resource "aiven_kafka_quota" "{{ .resource_name }}" {
  project             = data.aiven_project.foo.project
  service_name        = aiven_kafka.bar.service_name
  client_id           = "{{ .client_id }}"
  producer_byte_rate  = {{ .producer_byte_rate }}
}`)

	registry.MustAddTemplate(t, "wrong_configuration", `
resource "aiven_kafka_quota" "{{ .resource_name }}" {
  project             = data.aiven_project.foo.project
  service_name        = aiven_kafka.bar.service_name
  consumer_byte_rate  = {{ .consumer_byte_rate }}
  producer_byte_rate  = {{ .producer_byte_rate }}
  request_percentage  = {{ .request_percentage }}
}`)

	var newComposition = func() *acc.CompositionBuilder {
		return registry.NewCompositionBuilder().
			Add("project_data", map[string]interface{}{
				"project": projectName}).
			Add("aiven_kafka", map[string]interface{}{
				"service_name": serviceName,
			})
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenKafkaQuotaDestroy,
		Steps: []resource.TestStep{
			{
				Config: newComposition().
					Add("kafka_quota_full", map[string]any{
						"resource_name":      "full",
						"service_name":       serviceName,
						"user":               user,
						"client_id":          clientID,
						"consumer_byte_rate": 1000,
						"producer_byte_rate": 1000,
						"request_percentage": 101,
					}).
					MustRender(t),
				ExpectError: regexp.MustCompile(`expected .+ to be in the range \(\d+ - \d+\), got \d+`),
			},
			{
				Config: newComposition().
					Add("wrong_configuration", map[string]any{
						"resource_name":      "full",
						"service_name":       serviceName,
						"consumer_byte_rate": 1000,
						"producer_byte_rate": 1000,
						"request_percentage": 10,
					}).
					MustRender(t),
				ExpectError: regexp.MustCompile(`at least one of user or client_id must be specified`),
			},
			{
				Config: newComposition().
					Add("kafka_quota_full", map[string]any{
						"resource_name":      "full",
						"service_name":       serviceName,
						"user":               user,
						"client_id":          clientID,
						"consumer_byte_rate": 1000,
						"producer_byte_rate": 1000,
						"request_percentage": 10,
					}).
					Add("kafka_quota_user", map[string]any{
						"resource_name":      "user",
						"service_name":       serviceName,
						"user":               user,
						"request_percentage": 20,
					}).
					MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "project", projectName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "service_name", serviceName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "user", user),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "client_id", clientID),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "consumer_byte_rate", "1000"),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "producer_byte_rate", "1000"),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "request_percentage", "10"),

					resource.TestCheckResourceAttr(fmt.Sprintf("%s.user", kafkaQuotaResource), "project", projectName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.user", kafkaQuotaResource), "service_name", serviceName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.user", kafkaQuotaResource), "user", user),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.user", kafkaQuotaResource), "request_percentage", "20"),
				),
			},
			{
				Config: newComposition().
					Add("kafka_quota_client_id", map[string]any{
						"resource_name":      "client",
						"service_name":       serviceName,
						"client_id":          clientID,
						"producer_byte_rate": 1000,
					}).
					MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "project", projectName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "service_name", serviceName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "client_id", clientID),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "producer_byte_rate", "1000"),
				),
			},
			{
				Taint: []string{fmt.Sprintf("%s.client", kafkaQuotaResource)},
				Config: newComposition().
					Add("kafka_quota_client_id", map[string]any{
						"resource_name":      "client",
						"service_name":       serviceName,
						"client_id":          clientID,
						"producer_byte_rate": 1000,
					}).
					MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "project", projectName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "service_name", serviceName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "client_id", clientID),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "producer_byte_rate", "1000"),
				),
			},
		},
	})
}

func testAccCheckAivenKafkaQuotaDestroy(s *terraform.State) error {
	var (
		c, err = acc.GetTestGenAivenClient()
		ctx    = context.Background()
	)

	if err != nil {
		return fmt.Errorf("failed to instantiate GenAiven client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_kafka_quota" {
			continue
		}

		p, sn, cID, u, err := schemautil.SplitResourceID4(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error splitting resource ID: %w", err)
		}

		var params [][2]string
		if u != "" {
			params = append(params, kafka.ServiceKafkaQuotaDeleteUser(u))
		}
		if cID != "" {
			params = append(params, kafka.ServiceKafkaQuotaDeleteClientId(cID))
		}

		_, err = c.ServiceKafkaQuotaDescribe(ctx, p, sn, params...)
		if err != nil {
			if !common.IsCritical(err) {
				return nil
			}

			return err
		}

		return fmt.Errorf("kafka quota still exists")
	}

	return nil
}
