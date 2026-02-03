package database_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenClickHouseDatabase(t *testing.T) {
	resourceName := "aiven_clickhouse_database.foo"
	datasourceName := "data.aiven_clickhouse_database.foo"
	projectName := acc.ProjectName()

	// Creates shared ClickHouse service
	serviceName := acc.RandName("clickhouse")
	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		serviceName,
		acc.WithServiceType("clickhouse"),
		acc.WithPlan("startup-8"),
		acc.WithCloud("google-europe-west1"),
	)

	client, err := acc.GetTestGenAivenClient()
	require.NoError(t, err)

	t.Run("backward compatibility test", func(t *testing.T) {
		dbName := acc.RandName("compat")
		config := testAccClickHouseDatabaseWithDatasource(projectName, serviceName, dbName, "")
		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() { acc.TestAccPreCheck(t) },
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				PreConfig:          func() { require.NoError(t, <-serviceIsReady) },
				TFConfig:           config,
				OldProviderVersion: "4.50.0",
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "name", dbName),

					// Datasource checks
					resource.TestCheckResourceAttr(datasourceName, "project", projectName),
					resource.TestCheckResourceAttr(datasourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(datasourceName, "name", dbName),
				),
			}),
		})
	})

	t.Run("base test", func(t *testing.T) {
		dbName := acc.RandName("basic")
		configTerminationNil := testAccClickHouseDatabaseWithDatasource(projectName, serviceName, dbName, "")
		configTerminationTrue := testAccClickHouseDatabaseWithDatasource(projectName, serviceName, dbName, "termination_protection=true")
		configTerminationFalse := testAccClickHouseDatabaseWithDatasource(projectName, serviceName, dbName, "termination_protection=false")
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenClickHouseDatabaseResourceDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() { require.NoError(t, <-serviceIsReady) },
					Config:    configTerminationNil,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "name", dbName),
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),

						// Datasource checks
						resource.TestCheckResourceAttr(datasourceName, "project", projectName),
						resource.TestCheckResourceAttr(datasourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(datasourceName, "name", dbName),
					),
				},
				{
					// Tests that termination protection can be enabled
					Config: configTerminationTrue,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "name", dbName),
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
					),
				},
				{
					// Fails to remove termination_protection=true resource
					Config:      configTerminationTrue,
					Destroy:     true,
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`The resource ` + "`aiven_clickhouse_database`" + ` has termination protection enabled`),
				},
				{
					// Sets to false: tests that "false" bool value is applied
					Config: configTerminationFalse,
					Check:  resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				},
				{
					// Adds back termination_protection=true to test removal, see next step
					Config: configTerminationTrue,
					Check:  resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
				},
				{
					// Removing termination_protection disables it, gets false.
					Config: configTerminationNil,
					Check:  resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				},
				{
					Config:       configTerminationNil,
					ResourceName: resourceName,
					ImportState:  true,
					ImportStateIdFunc: func(s *terraform.State) (string, error) {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return "", fmt.Errorf("expected resource '%s' to be present in the state", resourceName)
						}
						if _, ok := rs.Primary.Attributes["name"]; !ok {
							return "", fmt.Errorf("expected resource '%s' to have 'name' attribute", resourceName)
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
						databaseName, ok := attributes["name"]
						if !ok {
							return errors.New("expected 'name' field to be set")
						}
						expectedID := fmt.Sprintf("%s/%s/%s", projectName, serviceName, databaseName)
						if !strings.EqualFold(s[0].ID, expectedID) {
							return fmt.Errorf("expected ID to match '%s', but got: %s", expectedID, s[0].ID)
						}
						return nil
					},
				},
				{
					// This resource has RemoveMissing set to true, so it should be recreated if it's missing (deleted).
					Config: testAccClickHouseDatabase(projectName, serviceName, dbName),
					PreConfig: func() {
						err := client.ServiceClickHouseDatabaseDelete(t.Context(), projectName, serviceName, dbName)
						require.NoError(t, err)
					},
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
					Check: resource.ComposeTestCheckFunc(
						// Doesn't have ID as it will be recreated
						resource.TestCheckNoResourceAttr(resourceName, "id"),
					),
				},
				{
					// Resource is recreated after being applied
					Config: testAccClickHouseDatabase(projectName, serviceName, dbName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s", projectName, serviceName, dbName)),
						func(state *terraform.State) error {
							list, err := client.ServiceClickHouseDatabaseList(t.Context(), projectName, serviceName)
							require.NoError(t, err)
							for _, db := range list {
								if db.Name == dbName {
									return nil
								}
							}
							return fmt.Errorf("clickhouse database %q not found after recreation", dbName)
						},
					),
				},
			},
		})
	})
}

func testAccCheckAivenClickHouseDatabaseResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// loop through the resources in state, verifying each database is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_clickhouse_database" {
			continue
		}

		projectName, serviceName, databaseName, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		list, err := c.ServiceClickHouseDatabaseList(ctx, projectName, serviceName)
		if err != nil {
			return err
		}

		for _, d := range list {
			if d.Name == databaseName {
				return fmt.Errorf("clickhouse database %q still exists", databaseName)
			}
		}
	}

	return nil
}

func testAccClickHouseDatabase(project, serviceName, dbName string) string {
	return fmt.Sprintf(`
resource "aiven_clickhouse_database" "foo" {
  project      = %q
  service_name = %q
  name         = %q
}
`, project, serviceName, dbName)
}

func testAccClickHouseDatabaseWithDatasource(project string, serviceName, dbName, terminationProtection string) string {
	return fmt.Sprintf(`
resource "aiven_clickhouse_database" "foo" {
  project      = %q
  service_name = %q
  name         = %q
  %s
}

data "aiven_clickhouse_database" "foo" {
  project      = aiven_clickhouse_database.foo.project
  service_name = aiven_clickhouse_database.foo.service_name
  name         = aiven_clickhouse_database.foo.name

  depends_on = [aiven_clickhouse_database.foo]
}`, project, serviceName, dbName, terminationProtection)
}
