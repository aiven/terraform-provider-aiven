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

func TestAccAivenMySQLDatabase(t *testing.T) {
	acc.SkipIfNotAcc(t)
	t.Parallel()

	resourceName := "aiven_mysql_database.foo"
	datasourceName := "data.aiven_mysql_database.foo"
	projectName := acc.ProjectName()

	// Creates shared MySQL
	serviceName := acc.RandName("mysql")
	cleanup, err := acc.CreateTestService(
		t.Context(),
		projectName,
		serviceName,
		acc.WithServiceType("mysql"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
	)

	require.NoError(t, err)
	defer cleanup()

	t.Run("backward compatibility test", func(t *testing.T) {
		dbName := acc.RandName("compatibility")
		config := testAccMySQLDatabaseResource(projectName, serviceName, dbName, "")
		resource.Test(t, resource.TestCase{
			PreCheck: func() { acc.TestAccPreCheck(t) },
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				TFConfig:           config,
				OldProviderVersion: "4.47.0",
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", dbName),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),

					// Datasource checks
					resource.TestCheckResourceAttr(datasourceName, "project", projectName),
					resource.TestCheckResourceAttr(datasourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(datasourceName, "database_name", dbName),
				),
			}),
		})
	})

	t.Run("base test", func(t *testing.T) {
		dbName := acc.RandName("basic")
		configTerminationNil := testAccMySQLDatabaseResource(projectName, serviceName, dbName, "")
		configTerminationTrue := testAccMySQLDatabaseResource(projectName, serviceName, dbName, "termination_protection=true")
		configTerminationFalse := testAccMySQLDatabaseResource(projectName, serviceName, dbName, "termination_protection=false")
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenMySQLDatabaseResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: configTerminationNil,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "database_name", dbName),
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),

						// Datasource checks
						resource.TestCheckResourceAttr(datasourceName, "project", projectName),
						resource.TestCheckResourceAttr(datasourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(datasourceName, "database_name", dbName),
					),
				},
				{
					// Tests that termination protection can be enabled
					Config: configTerminationTrue,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "database_name", dbName),
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
					),
				},
				{
					// Fails to remove termination_protection=true resource
					Config:      configTerminationTrue,
					Destroy:     true,
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`The resource ` + "`aiven_mysql_database`" + ` has termination protection enabled and\s+cannot be deleted`),
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
						if _, ok := rs.Primary.Attributes["database_name"]; !ok {
							return "", fmt.Errorf("expected resource '%s' to have 'database_name' attribute", resourceName)
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
						databaseName, ok := attributes["database_name"]
						if !ok {
							return errors.New("expected 'database_name' field to be set")
						}
						expectedID := fmt.Sprintf("%s/%s/%s", projectName, serviceName, databaseName)
						if !strings.EqualFold(s[0].ID, expectedID) {
							return fmt.Errorf("expected ID to match '%s', but got: %s", expectedID, s[0].ID)
						}
						return nil
					},
				},
			},
		})
	})
}

func testAccCheckAivenMySQLDatabaseResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// loop through the resources in state, verifying each database is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_mysql_database" {
			continue
		}

		projectName, serviceName, databaseName, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		list, err := c.ServiceDatabaseList(ctx, projectName, serviceName)
		if err != nil {
			return err
		}

		for _, d := range list {
			if d.DatabaseName == databaseName {
				return fmt.Errorf("mysql database %q still exists", databaseName)
			}
		}
	}

	return nil
}

func testAccMySQLDatabaseResource(project string, serviceName, dbName, terminationProtection string) string {
	return fmt.Sprintf(`
resource "aiven_mysql_database" "foo" {
  project       = %q
  service_name  = %q
  database_name = %q
  %s
}

data "aiven_mysql_database" "foo" {
  project       = aiven_mysql_database.foo.project
  service_name  = aiven_mysql_database.foo.service_name
  database_name = aiven_mysql_database.foo.database_name

  depends_on = [aiven_mysql_database.foo]
}`, project, serviceName, dbName, terminationProtection)
}
