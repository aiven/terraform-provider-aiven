package jardeployment_test

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

func TestAccAivenFlinkJarApplicationDeployment(t *testing.T) {
	acc.SkipIfNotBeta(t)

	const resourceName = "aiven_flink_jar_application_deployment.foo"

	projectName := acc.ProjectName()
	jarFile := acc.FlinkJarFile(t)

	serviceName := acc.RandName("flink")
	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		serviceName,
		acc.WithServiceType("flink"),
		acc.WithPlan("business-4"),
		acc.WithCloud("google-europe-west1"),
		// Jar applications are only available when the service accepts custom code.
		acc.WithUserConfig(map[string]any{"custom_code": true}),
	)

	// The new provider must read the SDKv2 state as is, which the second step's empty plan proves.
	// restart_enabled is the field to watch: the SDK schema had no default and the API never
	// returns the field, so its value only ever comes from the configuration. Both cases run, an
	// omitted field and an explicit one, and neither may end up replacing the deployment.
	compatCases := []struct {
		name           string
		deploymentAttr string
		checks         []resource.TestCheckFunc
	}{
		{
			name: "restart_enabled omitted",
			// Whatever the old provider stored is what the new one has to keep, so the
			// value itself is asserted in the base test below, on the new provider alone.
		},
		{
			name:           "restart_enabled set",
			deploymentAttr: "restart_enabled = false",
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "restart_enabled", "false"),
			},
		},
	}

	for _, tc := range compatCases {
		t.Run("backward compatibility test: "+tc.name, func(t *testing.T) {
			config := testAccFlinkJarApplicationDeployment(
				projectName, serviceName, acc.RandName("compat"), jarFile, tc.deploymentAttr,
			)
			checks := append([]resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "project", projectName),
				resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
				resource.TestCheckResourceAttrSet(resourceName, "application_id"),
				resource.TestCheckResourceAttrSet(resourceName, "version_id"),
				resource.TestCheckResourceAttrSet(resourceName, "deployment_id"),
				resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				resource.TestCheckResourceAttrSet(resourceName, "created_by"),
			}, tc.checks...)

			resource.ParallelTest(t, resource.TestCase{
				PreCheck: func() { acc.TestAccPreCheck(t) },
				Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
					PreConfig:          func() { require.NoError(t, <-serviceIsReady) },
					TFConfig:           config,
					OldProviderVersion: "4.61.0",
					Checks:             resource.ComposeTestCheckFunc(checks...),
				}),
			})
		})
	}

	t.Run("base test", func(t *testing.T) {
		appName := acc.RandName("basic")
		config := testAccFlinkJarApplicationDeployment(projectName, serviceName, appName, jarFile, "")
		configReplaced := testAccFlinkJarApplicationDeployment(projectName, serviceName, appName, jarFile, `
  parallelism     = 2
  restart_enabled = false
`)

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenFlinkJarApplicationDeploymentDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() { require.NoError(t, <-serviceIsReady) },
					Config:    config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttrSet(resourceName, "application_id"),
						resource.TestCheckResourceAttrSet(resourceName, "version_id"),
						resource.TestCheckResourceAttrSet(resourceName, "deployment_id"),
						resource.TestCheckResourceAttrSet(resourceName, "created_at"),
						resource.TestCheckResourceAttrSet(resourceName, "created_by"),
						resource.TestCheckResourceAttrSet(resourceName, "status"),
						resource.TestCheckResourceAttr(resourceName, "parallelism", "1"),
						resource.TestCheckResourceAttr(resourceName, "restart_enabled", "true"),
					),
				},
				{
					Config:            config,
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
					// The job reports its progress, so these change between the apply and the import.
					ImportStateVerifyIgnore: []string{"status", "job_id", "last_savepoint", "error_msg"},
				},
				{
					// The deployment has no update operation: changed values cancel and recreate it.
					Config: configReplaced,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(resourceName, "deployment_id"),
						resource.TestCheckResourceAttr(resourceName, "parallelism", "2"),
						resource.TestCheckResourceAttr(resourceName, "restart_enabled", "false"),
					),
				},
			},
		})
	})
}

func testAccCheckAivenFlinkJarApplicationDeploymentDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_flink_jar_application_deployment" {
			continue
		}

		projectName, serviceName, applicationID, deploymentID, err := schemautil.SplitResourceID4(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = c.ServiceFlinkGetJarApplicationDeployment(ctx, projectName, serviceName, applicationID, deploymentID)
		if avngen.IsNotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("flink jar application deployment %q still exists", rs.Primary.ID)
	}

	return nil
}

func testAccFlinkJarApplicationDeployment(projectName, serviceName, appName, jarFile, deploymentExtra string) string {
	return fmt.Sprintf(`
resource "aiven_flink_jar_application" "foo" {
  project      = %[1]q
  service_name = %[2]q
  name         = %[3]q
}

resource "aiven_flink_jar_application_version" "foo" {
  project        = %[1]q
  service_name   = %[2]q
  application_id = aiven_flink_jar_application.foo.application_id
  source         = %[4]q
}

resource "aiven_flink_jar_application_deployment" "foo" {
  project        = %[1]q
  service_name   = %[2]q
  application_id = aiven_flink_jar_application.foo.application_id
  version_id     = aiven_flink_jar_application_version.foo.application_version_id
  %[5]s
}
`, projectName, serviceName, appName, jarFile, deploymentExtra)
}
