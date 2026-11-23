package kafkaquota_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aiven/go-client-codegen/handler/kafka"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const kafkaQuotaResource = "aiven_kafka_quota"

func TestAccAivenKafkaQuota(t *testing.T) {
	projectName := acc.ProjectName()
	randName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	serviceName := fmt.Sprintf("test-acc-sr-%s", randName)
	user := fmt.Sprintf("acc_test_user_%s", randName)
	clientID := fmt.Sprintf("acc_test_client_%s", randName)

	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		serviceName,
		acc.WithServiceType("kafka"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
	)

	t.Run("backward_compat", func(t *testing.T) {
		u1 := fmt.Sprintf("%s_1", user)
		c1 := fmt.Sprintf("%s_1", clientID)
		u2 := fmt.Sprintf("%s_2", user)
		c3 := fmt.Sprintf("%s_3", clientID)

		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() {
				acc.TestAccPreCheck(t)
				require.NoError(t, <-serviceIsReady)
			},
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				PreConfig: func() { require.NoError(t, <-serviceIsReady) },
				TFConfig:  testAccKafkaQuotaBackwardCompatConfig(projectName, serviceName, u1, c1, u2, c3),
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "id", schemautil.BuildResourceID(projectName, serviceName, c1, u1)),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.user_only", kafkaQuotaResource), "id", schemautil.BuildResourceID(projectName, serviceName, "", u2)),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.client_only", kafkaQuotaResource), "id", schemautil.BuildResourceID(projectName, serviceName, c3, "")),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "consumer_byte_rate", "4000"),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.user_only", kafkaQuotaResource), "request_percentage", "20.22"),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.client_only", kafkaQuotaResource), "producer_byte_rate", "2000"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.user_only", kafkaQuotaResource), "client_id"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.client_only", kafkaQuotaResource), "user"),
				),
			}),
		})
	})

	t.Run("basic", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck: func() {
				acc.TestAccPreCheck(t)
				require.NoError(t, <-serviceIsReady)
			},
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaQuotaDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccKafkaQuotaFullConfig(
						projectName, serviceName, user, clientID,
						1000, 1000, 101,
					),
					ExpectError: regexp.MustCompile(`101\s+is\s+greater\s+than\s+the\s+maximum\s+of\s+100`),
				},
				{
					Config: testAccKafkaQuotaMissingIdentityConfig(projectName, serviceName),
					ExpectError: regexp.MustCompile(
						`At\s+least\s+one\s+attribute\s+out\s+of\s+\[\S*client_id\S*user\S*\]\s+must\s+be\s+specified`,
					),
				},
				{
					Config: testAccKafkaQuotaFullConfig(projectName, serviceName, user, clientID, 1000, 1000, 10),
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
					Config:             testAccKafkaQuotaFullConfig(projectName, serviceName, user, clientID, 1000, 1000, 10),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
				{
					Config:             testAccKafkaQuotaFullConfig(projectName, serviceName, user, clientID, 2000, 2000, 100),
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
				},
				{
					Config: testAccKafkaQuotaFullConfig(projectName, serviceName, user, clientID, 3000, 3000, 10),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(fmt.Sprintf("%s.full", kafkaQuotaResource), plancheck.ResourceActionUpdate),
							acc.ExpectOnlyAttributesChanged(fmt.Sprintf("%s.full", kafkaQuotaResource), "consumer_byte_rate", "producer_byte_rate"),
						},
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "consumer_byte_rate", "3000"),
						resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "producer_byte_rate", "3000"),
						resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "request_percentage", "10"),
					),
				},
				{
					Config: testAccKafkaQuotaFullConfig(
						projectName, serviceName, fmt.Sprintf("%s_updated", user), clientID, 3000, 3000, 10,
					),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectNonEmptyPlan(),
							plancheck.ExpectResourceAction(fmt.Sprintf("%s.full", kafkaQuotaResource), plancheck.ResourceActionReplace),
						},
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(fmt.Sprintf("%s.full", kafkaQuotaResource), "user", fmt.Sprintf("%s_updated", user)),
					),
				},
				{
					Config: testAccKafkaQuotaConsumerOnlyConfig(projectName, serviceName, user, clientID, 100),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "consumer_byte_rate", "100"),
						resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "producer_byte_rate"),
						resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "request_percentage"),
					),
				},
				{
					Config: testAccKafkaQuotaConsumerOnlyConfig(projectName, serviceName, user, clientID, 3000),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), plancheck.ResourceActionUpdate),
							acc.ExpectOnlyAttributesChanged(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "consumer_byte_rate"),
						},
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "consumer_byte_rate", "3000"),
						resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "producer_byte_rate"),
						resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.byte_rate", kafkaQuotaResource), "request_percentage"),
					),
				},
				{
					Config: testAccKafkaQuotaMultiConfig(projectName, serviceName, user, clientID),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(fmt.Sprintf("%s.new_full", kafkaQuotaResource), "request_percentage", "40.5"),
						resource.TestCheckResourceAttr(fmt.Sprintf("%s.user", kafkaQuotaResource), "request_percentage", "20.22"),
						resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.user", kafkaQuotaResource), "client_id"),
						resource.TestCheckResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "producer_byte_rate", "2000"),
						resource.TestCheckNoResourceAttr(fmt.Sprintf("%s.client", kafkaQuotaResource), "user"),
					),
				},
			},
		})
	})
}

func testAccKafkaQuotaFullConfig(project, service, user, clientID string, consumer, producer int, pct float64) string {
	return fmt.Sprintf(`
resource "%s" "full" {
  project            = %q
  service_name       = %q
  user               = %q
  client_id          = %q
  consumer_byte_rate = %d
  producer_byte_rate = %d
  request_percentage = %g
}
`, kafkaQuotaResource, project, service, user, clientID, consumer, producer, pct)
}

func testAccKafkaQuotaMissingIdentityConfig(project, service string) string {
	return fmt.Sprintf(`
resource "%s" "full" {
  project            = %q
  service_name       = %q
  consumer_byte_rate = 1000
  producer_byte_rate = 1000
  request_percentage = 10
}
`, kafkaQuotaResource, project, service)
}

func testAccKafkaQuotaConsumerOnlyConfig(project, service, user, clientID string, consumer int) string {
	return fmt.Sprintf(`
resource "%s" "byte_rate" {
  project            = %q
  service_name       = %q
  user               = %q
  client_id          = %q
  consumer_byte_rate = %d
}
`, kafkaQuotaResource, project, service, user, clientID, consumer)
}

func testAccKafkaQuotaMultiConfig(project, service, user, clientID string) string {
	return fmt.Sprintf(`
resource "%s" "new_full" {
  project            = %q
  service_name       = %q
  user               = "%s_1"
  client_id          = "%s_1"
  consumer_byte_rate = 4000
  producer_byte_rate = 4000
  request_percentage = 40.5
}

resource "%s" "user" {
  project            = %q
  service_name       = %q
  user               = "%s_2"
  request_percentage = 20.22
}

resource "%s" "client" {
  project            = %q
  service_name       = %q
  client_id          = "%s_3"
  producer_byte_rate = 2000
}
`, kafkaQuotaResource, project, service, user, clientID,
		kafkaQuotaResource, project, service, user,
		kafkaQuotaResource, project, service, clientID)
}

func testAccKafkaQuotaBackwardCompatConfig(project, service, u1, c1, u2, c3 string) string {
	return fmt.Sprintf(`
resource "%s" "full" {
  project            = %q
  service_name       = %q
  user               = %q
  client_id          = %q
  consumer_byte_rate = 4000
  producer_byte_rate = 4000
  request_percentage = 40.5
}

resource "%s" "user_only" {
  project            = %q
  service_name       = %q
  user               = %q
  request_percentage = 20.22
}

resource "%s" "client_only" {
  project            = %q
  service_name       = %q
  client_id          = %q
  producer_byte_rate = 2000
}
`, kafkaQuotaResource, project, service, u1, c1,
		kafkaQuotaResource, project, service, u2,
		kafkaQuotaResource, project, service, c3)
}

func testAccCheckAivenKafkaQuotaDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("failed to instantiate GenAiven client: %w", err)
	}
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != kafkaQuotaResource {
			continue
		}

		p, sn, cID, u, err := schemautil.SplitResourceID4(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error splitting resource ID: %w", err)
		}

		var params [][2]string
		if u != "" {
			params = append(params, kafka.ServiceKafkaQuotaDescribeUser(u))
		}
		if cID != "" {
			params = append(params, kafka.ServiceKafkaQuotaDescribeClientId(cID))
		}

		_, err = c.ServiceKafkaQuotaDescribe(ctx, p, sn, params...)
		if err != nil {
			if !common.IsCritical(err) {
				continue
			}
			return err
		}

		return fmt.Errorf("kafka quota %q still exists", rs.Primary.ID)
	}

	return nil
}
