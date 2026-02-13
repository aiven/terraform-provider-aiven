package application_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenFlinkApplication(t *testing.T) {
	resourceName := "aiven_flink_application.foo"
	datasourceName := "data.aiven_flink_application.foo"
	projectName := acc.ProjectName()

	serviceName := acc.RandName("flink")
	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		serviceName,
		acc.WithServiceType("flink"),
		acc.WithPlan("business-4"),
		acc.WithCloud("google-europe-west1"),
	)

	client, err := acc.GetTestGenAivenClient()
	require.NoError(t, err)

	t.Run("backward compatibility test", func(t *testing.T) {
		appName := acc.RandName("compat")
		config := testAccFlinkApplication(projectName, serviceName, appName)
		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() { acc.TestAccPreCheck(t) },
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				PreConfig:          func() { require.NoError(t, <-serviceIsReady) },
				TFConfig:           config,
				OldProviderVersion: "4.50.0",
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "name", appName),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_by"),
				),
			}),
		})
	})

	t.Run("base test", func(t *testing.T) {
		appName := acc.RandName("basic")
		config := testAccFlinkApplicationWithDatasource(projectName, serviceName, appName)
		updatedName := acc.RandName("updated")
		configUpdated := testAccFlinkApplication(projectName, serviceName, updatedName)
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenFlinkApplicationDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() { require.NoError(t, <-serviceIsReady) },
					Config:    config,
					Check: resource.ComposeTestCheckFunc(
						// Resource checks
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "name", appName),
						resource.TestCheckResourceAttrSet(resourceName, "application_id"),
						resource.TestCheckResourceAttrSet(resourceName, "created_at"),
						resource.TestCheckResourceAttrSet(resourceName, "created_by"),
						resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
						resource.TestCheckResourceAttrSet(resourceName, "updated_by"),

						// Datasource checks
						resource.TestCheckResourceAttr(datasourceName, "project", projectName),
						resource.TestCheckResourceAttr(datasourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(datasourceName, "name", appName),
						resource.TestCheckResourceAttrSet(datasourceName, "application_id"),
						resource.TestCheckResourceAttrSet(datasourceName, "created_at"),
						resource.TestCheckResourceAttrSet(datasourceName, "created_by"),
						resource.TestCheckResourceAttrSet(datasourceName, "updated_at"),
						resource.TestCheckResourceAttrSet(datasourceName, "updated_by"),
					),
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
						if attributes["application_id"] == "" {
							return fmt.Errorf("expected 'application_id' to be set after import")
						}
						if attributes["name"] == "" {
							return fmt.Errorf("expected 'name' to be set after import")
						}
						if attributes["created_at"] == "" {
							return fmt.Errorf("expected 'created_at' to be set after import")
						}
						if attributes["created_by"] == "" {
							return fmt.Errorf("expected 'created_by' to be set after import")
						}
						if attributes["updated_at"] == "" {
							return fmt.Errorf("expected 'updated_at' to be set after import")
						}
						if attributes["updated_by"] == "" {
							return fmt.Errorf("expected 'updated_by' to be set after import")
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
					Config: testAccFlinkApplication(projectName, serviceName, updatedName),
					PreConfig: func() {
						// Get the application ID from state by listing
						apps, err := client.ServiceFlinkListApplications(t.Context(), projectName, serviceName)
						require.NoError(t, err)
						for _, app := range apps {
							if app.Name == updatedName {
								_, err = client.ServiceFlinkDeleteApplication(t.Context(), projectName, serviceName, app.Id)
								require.NoError(t, err)
								return
							}
						}
						t.Fatalf("flink application %q not found for drift test", updatedName)
					},
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
				},
				{
					// Resource is recreated after drift
					Config: testAccFlinkApplication(projectName, serviceName, updatedName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", updatedName),
						resource.TestCheckResourceAttrSet(resourceName, "application_id"),
					),
				},
			},
		})
	})
}

func testAccCheckAivenFlinkApplicationDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_flink_application" {
			continue
		}

		projectName, serviceName, applicationID, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = c.ServiceFlinkGetApplication(ctx, projectName, serviceName, applicationID)
		if err == nil {
			return fmt.Errorf("flink application %q still exists", applicationID)
		}
	}

	return nil
}

func testAccFlinkApplication(project, serviceName, appName string) string {
	return fmt.Sprintf(`
resource "aiven_flink_application" "foo" {
  project      = %q
  service_name = %q
  name         = %q
}
`, project, serviceName, appName)
}

func testAccFlinkApplicationWithDatasource(project, serviceName, appName string) string {
	return fmt.Sprintf(`
resource "aiven_flink_application" "foo" {
  project      = %q
  service_name = %q
  name         = %q
}

data "aiven_flink_application" "foo" {
  project      = aiven_flink_application.foo.project
  service_name = aiven_flink_application.foo.service_name
  name         = aiven_flink_application.foo.name

  depends_on = [aiven_flink_application.foo]
}
`, project, serviceName, appName)
}
