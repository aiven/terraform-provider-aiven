package jarapplication_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenFlinkJarApplication(t *testing.T) {
	acc.SkipIfNotBeta(t)

	resourceName := "aiven_flink_jar_application.foo"
	versionResourceName := "aiven_flink_jar_application_version.foo"
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

	client, err := acc.GetTestGenAivenClient()
	require.NoError(t, err)

	// The SDKv2 state carries an application_versions entry, which the new provider must read
	// without asking for a change: the second step's empty plan proves it.
	t.Run("backward compatibility test", func(t *testing.T) {
		appName := acc.RandName("compat")
		config := testAccFlinkJarApplicationVersion(projectName, serviceName, appName, jarFile)
		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() { acc.TestAccPreCheck(t) },
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				PreConfig:          func() { require.NoError(t, <-serviceIsReady) },
				TFConfig:           config,
				OldProviderVersion: "4.61.0",
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "name", appName),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_by"),
					// The version resource is not migrated, its state must not change either.
					resource.TestCheckResourceAttr(versionResourceName, "file_info.0.file_status", "READY"),
					resource.TestCheckResourceAttrSet(versionResourceName, "source_checksum"),
				),
			}),
		})
	})

	// Read-only collections are computed attributes, so they hold what the API reports.
	// Re-applying the same configuration must not diff.
	t.Run("read only attributes", func(t *testing.T) {
		appName := acc.RandName("readonly")
		config := testAccFlinkJarApplicationVersion(projectName, serviceName, appName, jarFile)
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenFlinkJarApplicationDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() { require.NoError(t, <-serviceIsReady) },
					Config:    config,
					Check:     resource.TestCheckResourceAttr(versionResourceName, "file_info.0.file_status", "READY"),
				},
				{
					// The jar is uploaded after the application is created,
					// so the version lands on this step's refresh.
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "application_versions.#", "1"),
						resource.TestCheckResourceAttrPair(
							resourceName, "application_versions.0.id",
							versionResourceName, "application_version_id",
						),
						resource.TestCheckResourceAttr(resourceName, "application_versions.0.file_info.0.file_status", "READY"),
						// Nothing is deployed, so the deployment stays empty.
						resource.TestCheckResourceAttr(resourceName, "current_deployment.#", "0"),
					),
				},
			},
		})
	})

	t.Run("current deployment cleared", func(t *testing.T) {
		testAccAivenFlinkJarApplicationCurrentDeploymentCleared(
			t,
			projectName,
			serviceName,
			jarFile,
			serviceIsReady,
		)
	})

	t.Run("base test", func(t *testing.T) {
		appName := acc.RandName("basic")
		config := testAccFlinkJarApplication(projectName, serviceName, appName)
		updatedName := acc.RandName("updated")
		configUpdated := testAccFlinkJarApplication(projectName, serviceName, updatedName)
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenFlinkJarApplicationDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() { require.NoError(t, <-serviceIsReady) },
					Config:    config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "name", appName),
						resource.TestCheckResourceAttrSet(resourceName, "application_id"),
						resource.TestCheckResourceAttrSet(resourceName, "created_at"),
						resource.TestCheckResourceAttrSet(resourceName, "created_by"),
						resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
						resource.TestCheckResourceAttrSet(resourceName, "updated_by"),
						// No deployment exists yet.
						resource.TestCheckResourceAttr(resourceName, "current_deployment.#", "0"),
					),
					// No version has been uploaded, so the API reports an empty list.
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							resourceName,
							tfjsonpath.New("application_versions"),
							knownvalue.ListExact([]knownvalue.Check{}),
						),
					},
				},
				{
					// Test update: application name can be changed in-place
					Config: configUpdated,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", updatedName),
						resource.TestCheckResourceAttrSet(resourceName, "application_id"),
					),
				},
				{
					// Test import: verify all fields are populated
					Config:       configUpdated,
					ResourceName: resourceName,
					ImportState:  true,
					ImportStateIdFunc: func(s *terraform.State) (string, error) {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return "", fmt.Errorf("expected resource '%s' to be present in the state", resourceName)
						}
						return rs.Primary.ID, nil
					},
					ImportStateCheck: func(s []*terraform.InstanceState) error {
						if len(s) != 1 {
							return fmt.Errorf("expected only one instance to be imported, state: %#v", s)
						}
						attributes := s[0].Attributes
						if !strings.EqualFold(attributes["project"], projectName) {
							return fmt.Errorf("expected project to match '%s', got: '%s'", projectName, attributes["project"])
						}
						if !strings.EqualFold(attributes["service_name"], serviceName) {
							return fmt.Errorf("expected service_name to match '%s', got: '%s'", serviceName, attributes["service_name"])
						}
						for _, k := range []string{"application_id", "name", "created_at", "created_by", "updated_at", "updated_by"} {
							if attributes[k] == "" {
								return fmt.Errorf("expected %q to be set after import", k)
							}
						}
						expectedID := fmt.Sprintf("%s/%s/%s", projectName, serviceName, attributes["application_id"])
						if !strings.EqualFold(s[0].ID, expectedID) {
							return fmt.Errorf("expected ID to match '%s', but got: %s", expectedID, s[0].ID)
						}
						return nil
					},
				},
				{
					// RemoveMissing: delete via API, verify plan detects drift and wants to recreate
					Config: configUpdated,
					PreConfig: func() {
						apps, err := client.ServiceFlinkListJarApplications(t.Context(), projectName, serviceName)
						require.NoError(t, err)
						for _, app := range apps {
							if app.Name == updatedName {
								_, err = client.ServiceFlinkDeleteJarApplication(t.Context(), projectName, serviceName, app.Id)
								require.NoError(t, err)
								return
							}
						}
						t.Fatalf("flink jar application %q not found for drift test", updatedName)
					},
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
				},
				{
					// Resource is recreated after drift
					Config: configUpdated,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", updatedName),
						resource.TestCheckResourceAttrSet(resourceName, "application_id"),
					),
				},
			},
		})
	})
}

func testAccAivenFlinkJarApplicationCurrentDeploymentCleared(
	t *testing.T,
	projectName, serviceName, jarFile string,
	serviceIsReady <-chan error,
) {
	const (
		applicationResourceName = "aiven_flink_jar_application.stale_deployment"
		deploymentResourceName  = "aiven_flink_jar_application_deployment.stale_deployment"
	)

	applicationName := acc.RandName("current-deployment")

	client, err := acc.GetTestGenAivenClient()
	require.NoError(t, err)

	var applicationID, deploymentID string
	waitForCurrentDeployment := func(expectedDeploymentID string) {
		t.Helper()

		message := "current deployment did not disappear from the jar application"
		if expectedDeploymentID != "" {
			message = fmt.Sprintf("deployment %q did not appear on the jar application", expectedDeploymentID)
		}

		require.Eventually(t, func() bool {
			application, err := client.ServiceFlinkGetJarApplication(
				t.Context(), projectName, serviceName, applicationID,
			)
			if err != nil {
				return false
			}

			if expectedDeploymentID == "" {
				return application.CurrentDeployment == nil
			}

			return application.CurrentDeployment != nil && application.CurrentDeployment.Id == expectedDeploymentID
		}, 3*time.Minute, 2*time.Second, message)
	}

	withDeployment := testAccFlinkJarApplicationStaleCurrentDeploymentConfig(
		projectName, serviceName, applicationName, jarFile, true,
	)
	withoutDeployment := testAccFlinkJarApplicationStaleCurrentDeploymentConfig(
		projectName, serviceName, applicationName, jarFile, false,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenFlinkJarApplicationDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() { require.NoError(t, <-serviceIsReady) },
				Config:    withDeployment,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(applicationResourceName, "application_id"),
					resource.TestCheckResourceAttrSet(deploymentResourceName, "deployment_id"),
					func(state *terraform.State) error {
						application, ok := state.RootModule().Resources[applicationResourceName]
						if !ok {
							return fmt.Errorf("resource %q not found in state", applicationResourceName)
						}

						deployment, ok := state.RootModule().Resources[deploymentResourceName]
						if !ok {
							return fmt.Errorf("resource %q not found in state", deploymentResourceName)
						}

						applicationID = application.Primary.Attributes["application_id"]
						deploymentID = deployment.Primary.Attributes["deployment_id"]
						if applicationID == "" || deploymentID == "" {
							return fmt.Errorf("application or deployment ID is empty")
						}

						return nil
					},
				),
			},
			{
				// The application was read before its dependent deployment was created.
				// Wait for the API to expose the deployment, then refresh the parent state.
				PreConfig: func() { waitForCurrentDeployment(deploymentID) },
				Config:    withDeployment,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(applicationResourceName, "current_deployment.#", "1"),
					resource.TestCheckResourceAttrPair(
						applicationResourceName, "current_deployment.0.id",
						deploymentResourceName, "deployment_id",
					),
				),
			},
			{
				// Removing the deployment drives its cancel/delete state machine to completion.
				Config: withoutDeployment,
			},
			{
				// Prove that the API has cleared current_deployment before refreshing Terraform state.
				PreConfig: func() { waitForCurrentDeployment("") },
				Config:    withoutDeployment,
				Check: resource.TestCheckResourceAttr(
					applicationResourceName, "current_deployment.#", "0",
				),
			},
		},
	})
}

func testAccCheckAivenFlinkJarApplicationDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_flink_jar_application" {
			continue
		}

		projectName, serviceName, applicationID, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = c.ServiceFlinkGetJarApplication(ctx, projectName, serviceName, applicationID)
		if err == nil {
			return fmt.Errorf("flink jar application %q still exists", applicationID)
		}
	}

	return nil
}

func testAccFlinkJarApplication(project, serviceName, appName string) string {
	return fmt.Sprintf(`
resource "aiven_flink_jar_application" "foo" {
  project      = %q
  service_name = %q
  name         = %q
}
`, project, serviceName, appName)
}

// testAccFlinkJarApplicationVersion uploads a jar, so the API reports a version.
func testAccFlinkJarApplicationVersion(project, serviceName, appName, jarFile string) string {
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
`, project, serviceName, appName, jarFile)
}

func testAccFlinkJarApplicationStaleCurrentDeploymentConfig(
	projectName, serviceName, applicationName, jarFile string,
	withDeployment bool,
) string {
	config := fmt.Sprintf(`
resource "aiven_flink_jar_application" "stale_deployment" {
  project      = %[1]q
  service_name = %[2]q
  name         = %[3]q
}

resource "aiven_flink_jar_application_version" "stale_deployment" {
  project        = %[1]q
  service_name   = %[2]q
  application_id = aiven_flink_jar_application.stale_deployment.application_id
  source         = %[4]q
}
`, projectName, serviceName, applicationName, jarFile)

	if !withDeployment {
		return config
	}

	return config + `
resource "aiven_flink_jar_application_deployment" "stale_deployment" {
  project        = aiven_flink_jar_application.stale_deployment.project
  service_name   = aiven_flink_jar_application.stale_deployment.service_name
  application_id = aiven_flink_jar_application.stale_deployment.application_id
  version_id     = aiven_flink_jar_application_version.stale_deployment.application_version_id
}
`
}
