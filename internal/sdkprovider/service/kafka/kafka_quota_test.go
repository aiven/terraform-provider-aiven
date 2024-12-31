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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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

	registry.MustAddTemplate(t, "kafka_quota", `
resource "aiven_kafka_quota" "{{ .resource_name }}" {
  project      = data.aiven_project.foo.project
  service_name = aiven_kafka.bar.service_name
  {{- if .user }}
  user         = "{{ .user }}"
  {{- end }}
  {{- if .client_id }}
  client_id    = "{{ .client_id }}"
  {{- end }}
  {{- if .consumer_byte_rate }}
  consumer_byte_rate = {{ .consumer_byte_rate }}
  {{- end }}
  {{- if .producer_byte_rate }}
  producer_byte_rate = {{ .producer_byte_rate }}
  {{- end }}
  {{- if .request_percentage }}
  request_percentage = {{ .request_percentage }}
  {{- end }}
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
					Add("kafka_quota", map[string]any{
						"resource_name":      "full",
						"service_name":       serviceName,
						"user":               user,
						"client_id":          clientID,
						"consumer_byte_rate": 1000,
						"producer_byte_rate": 1000,
						"request_percentage": 101, // invalid value
					}).
					MustRender(t),
				ExpectError: regexp.MustCompile(`expected .+ to be in the range \(\d+ - \d+\), got \d+`),
			},
			{
				// missing user and client_id
				Config: newComposition().
					Add("kafka_quota", map[string]any{
						"resource_name":      "full",
						"service_name":       serviceName,
						"consumer_byte_rate": 1000,
						"producer_byte_rate": 1000,
						"request_percentage": 10,
					}).
					MustRender(t),
				ExpectError: regexp.MustCompile(`"(?:user|client_id)": one of ` + "`" + `client_id,user` + "`" + ` must be specified`),
			},
			{
				// valid configuration
				Config: newComposition().
					Add("kafka_quota", map[string]any{
						"resource_name":      "full",
						"service_name":       serviceName,
						"user":               user,
						"client_id":          clientID,
						"consumer_byte_rate": 1000,
						"producer_byte_rate": 1000,
						"request_percentage": 10,
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
				),
			},
			{
				// check that the resource is not updated without changes
				Config: newComposition().
					Add("kafka_quota", map[string]any{
						"resource_name":      "full",
						"service_name":       serviceName,
						"user":               user,
						"client_id":          clientID,
						"consumer_byte_rate": 1000,
						"producer_byte_rate": 1000,
						"request_percentage": 10,
					}).
					MustRender(t),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				// check plan that resource should be updated
				Config: newComposition().
					Add("kafka_quota", map[string]any{
						"resource_name":      "full",
						"service_name":       serviceName,
						"user":               user,
						"client_id":          clientID,
						"consumer_byte_rate": 2000,
						"producer_byte_rate": 2000,
						"request_percentage": 100,
					}).
					MustRender(t),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// check that the update action is triggered, only changed attributes are updated
				Config: newComposition().
					Add("kafka_quota", map[string]any{
						"resource_name":      "full",
						"service_name":       serviceName,
						"user":               user,
						"client_id":          clientID,
						"consumer_byte_rate": 3000,
						"producer_byte_rate": 3000,
						"request_percentage": 10,
					}).
					MustRender(t),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(fmt.Sprintf("%s.full", kafkaQuotaResource), plancheck.ResourceActionUpdate),
						acc.ExpectOnlyAttributesChanged(fmt.Sprintf("%s.full", kafkaQuotaResource), "consumer_byte_rate", "producer_byte_rate"),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "project", projectName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "service_name", serviceName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "user", user),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "client_id", clientID),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "consumer_byte_rate", "3000"),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "producer_byte_rate", "3000"),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "request_percentage", "10"),
				),
			},
			{
				// check that resource is replaced when user is updated
				Config: newComposition().
					Add("kafka_quota", map[string]any{
						"resource_name":      "full",
						"service_name":       serviceName,
						"user":               fmt.Sprintf("%s_updated", user),
						"client_id":          clientID,
						"consumer_byte_rate": 3000,
						"producer_byte_rate": 3000,
						"request_percentage": 10,
					}).
					MustRender(t),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction(fmt.Sprintf("%s.full", kafkaQuotaResource), plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "project", projectName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "service_name", serviceName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "user", fmt.Sprintf("%s_updated", user)),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "client_id", clientID),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "consumer_byte_rate", "3000"),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "producer_byte_rate", "3000"),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "request_percentage", "10"),
				),
			},
			{
				// create new resource with only consumer_byte_rate set
				Config: newComposition().
					Add("kafka_quota", map[string]any{
						"resource_name":      "byte_rate",
						"service_name":       serviceName,
						"user":               user,
						"client_id":          clientID,
						"consumer_byte_rate": 100,
					}).
					MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "project", projectName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "service_name", serviceName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "user", user),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "client_id", clientID),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "consumer_byte_rate", "100"),

					resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "producer_byte_rate"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "request_percentage"),
				),
			},
			{
				// check that the resource is updated and only consumer_byte_rate was modified
				Config: newComposition().
					Add("kafka_quota", map[string]any{
						"resource_name":      "byte_rate",
						"service_name":       serviceName,
						"user":               user,
						"client_id":          clientID,
						"consumer_byte_rate": 3000,
					}).
					MustRender(t),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), plancheck.ResourceActionUpdate),
						acc.ExpectOnlyAttributesChanged(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "consumer_byte_rate"),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "project", projectName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "service_name", serviceName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "user", user),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "client_id", clientID),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "consumer_byte_rate", "3000"),

					resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "producer_byte_rate"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "request_percentage"),
				),
			},
			{
				// craete multiple resources with different configurations
				Config: newComposition().
					Add("kafka_quota", map[string]any{
						"resource_name":      "new_full",
						"service_name":       serviceName,
						"user":               user,
						"client_id":          clientID,
						"consumer_byte_rate": 4000,
						"producer_byte_rate": 4000,
						"request_percentage": 40,
					}).
					Add("kafka_quota", map[string]any{
						"resource_name":      "user",
						"service_name":       serviceName,
						"user":               user,
						"request_percentage": 20,
					}).
					Add("kafka_quota", map[string]any{
						"resource_name":      "client",
						"service_name":       serviceName,
						"client_id":          clientID,
						"producer_byte_rate": 2000,
					}).
					MustRender(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.new_full", kafkaQuotaResource), "project", projectName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.new_full", kafkaQuotaResource), "service_name", serviceName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.new_full", kafkaQuotaResource), "user", user),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.new_full", kafkaQuotaResource), "client_id", clientID),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.new_full", kafkaQuotaResource), "consumer_byte_rate", "4000"),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.new_full", kafkaQuotaResource), "producer_byte_rate", "4000"),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.new_full", kafkaQuotaResource), "request_percentage", "40"),

					resource.TestCheckResourceAttr(fmt.Sprintf("%s.user", kafkaQuotaResource), "project", projectName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.user", kafkaQuotaResource), "service_name", serviceName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.user", kafkaQuotaResource), "user", user),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.user", kafkaQuotaResource), "request_percentage", "20"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.user", kafkaQuotaResource), "client_id"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.user", kafkaQuotaResource), "consumer_byte_rate"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.user", kafkaQuotaResource), "producer_byte_rate"),

					resource.TestCheckResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "project", projectName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "service_name", serviceName),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "client_id", clientID),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "producer_byte_rate", "2000"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "user"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "consumer_byte_rate"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "request_percentage"),
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
