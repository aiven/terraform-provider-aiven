package nativeacl_test

import (
	"context"
	"fmt"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenKafkaNativeACL(t *testing.T) {
	projectName := acc.ProjectName()
	kafkaName := acc.RandName("kafka")

	// All subtests share a single Kafka service to avoid provisioning one per case.
	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		kafkaName,
		acc.WithServiceType("kafka"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
	)
	require.NoError(t, <-serviceIsReady)

	t.Run("basic", func(t *testing.T) {
		resourceName := "aiven_kafka_native_acl.foo"
		resourceName2 := "aiven_kafka_native_acl.bar"

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaNativeACLResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccKafkaNativeACLResource(projectName, kafkaName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", kafkaName),
						resource.TestCheckResourceAttrSet(resourceName, "acl_id"),
						resource.TestCheckResourceAttr(resourceName, "resource_name", "name-test"),
						resource.TestCheckResourceAttr(resourceName, "resource_type", "Topic"),
						resource.TestCheckResourceAttr(resourceName, "pattern_type", "LITERAL"),
						resource.TestCheckResourceAttr(resourceName, "principal", "User:alice"),
						resource.TestCheckResourceAttr(resourceName, "host", "host-test"),
						resource.TestCheckResourceAttr(resourceName, "operation", "Create"),
						resource.TestCheckResourceAttr(resourceName, "permission_type", "ALLOW"),

						// The host defaults to "*" on the API side when omitted.
						resource.TestCheckResourceAttr(resourceName2, "host", "*"),
					),
				},
				{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	// Verifies that state created by the previous SDK-based provider version is
	// compatible with the Plugin Framework version. The old provider manages only
	// the ACL against the shared Kafka service, so we don't recreate the service
	// across the provider switch.
	t.Run("backward_compat", func(t *testing.T) {
		resourceName := "aiven_kafka_native_acl.foo"

		resource.Test(t, resource.TestCase{
			PreCheck:     func() { acc.TestAccPreCheck(t) },
			CheckDestroy: testAccCheckAivenKafkaNativeACLResourceDestroy,
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				TFConfig: testAccKafkaNativeACLResource(projectName, kafkaName),
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "resource_name", "name-test"),
					resource.TestCheckResourceAttr(resourceName, "resource_type", "Topic"),
					resource.TestCheckResourceAttr(resourceName, "pattern_type", "LITERAL"),
					resource.TestCheckResourceAttr(resourceName, "principal", "User:alice"),
					resource.TestCheckResourceAttr(resourceName, "operation", "Create"),
					resource.TestCheckResourceAttr(resourceName, "permission_type", "ALLOW"),
				),
			}),
		})
	})
}

func testAccKafkaNativeACLResource(projectName, kafkaName string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_native_acl" "foo" {
  project         = %[1]q
  service_name    = %[2]q
  resource_name   = "name-test"
  resource_type   = "Topic"
  pattern_type    = "LITERAL"
  principal       = "User:alice"
  host            = "host-test"
  operation       = "Create"
  permission_type = "ALLOW"
}

resource "aiven_kafka_native_acl" "bar" {
  project         = %[1]q
  service_name    = %[2]q
  resource_name   = "name-test"
  resource_type   = "Topic"
  pattern_type    = "LITERAL"
  principal       = "User:alice"
  operation       = "Create"
  permission_type = "ALLOW"
}
`, projectName, kafkaName)
}

func testAccCheckAivenKafkaNativeACLResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("failed to instantiate GenAiven client: %w", err)
	}

	ctx := context.Background()

	// loop through the resources in state, verifying each Kafka-native ACL is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_kafka_native_acl" {
			continue
		}

		projectName, serviceName, aclID, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = c.ServiceKafkaNativeAclGet(ctx, projectName, serviceName, aclID)
		if err != nil {
			if avngen.IsNotFound(err) {
				continue
			}

			return err
		}

		return fmt.Errorf("kafka native acl (%s) still exists", rs.Primary.ID)
	}

	return nil
}
