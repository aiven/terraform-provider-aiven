package database_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenPGDatabase(t *testing.T) {
	const resourceName = "aiven_pg_database.foo"
	const datasourceName = "data.aiven_pg_database.foo"
	const oldProviderVersion = "4.47.0"
	projectName := acc.ProjectName()

	// Creates shared PG
	serviceName := acc.RandName("pg")
	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		serviceName,
		acc.WithServiceType("pg"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
	)

	client, err := acc.GetTestGenAivenClient()
	require.NoError(t, err)

	t.Run("backward compatibility test (default lc_*)", func(t *testing.T) {
		dbName := acc.RandName("compatibility")
		config := testAccPGDatabaseWithDatasource(projectName, serviceName, dbName, "")
		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() { acc.TestAccPreCheck(t) },
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				PreConfig:          func() { require.NoError(t, <-serviceIsReady) },
				TFConfig:           config,
				OldProviderVersion: oldProviderVersion,
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", dbName),
					resource.TestCheckResourceAttr(resourceName, "lc_collate", "en_US.UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "lc_ctype", "en_US.UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),

					// Datasource checks
					resource.TestCheckResourceAttr(datasourceName, "project", projectName),
					resource.TestCheckResourceAttr(datasourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(datasourceName, "database_name", dbName),
				),
			}),
		})
	})

	t.Run("backward compatibility test (explicit lc_*)", func(t *testing.T) {
		dbName := acc.RandName("compatibility-explicitlc")
		config := testAccPGDatabaseWithDatasource(projectName, serviceName, dbName, `  lc_collate = "en_US.UTF-8"
	  lc_ctype   = "en_US.UTF-8"`)
		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() { acc.TestAccPreCheck(t) },
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				PreConfig:          func() { require.NoError(t, <-serviceIsReady) },
				TFConfig:           config,
				OldProviderVersion: oldProviderVersion,
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", dbName),
					resource.TestCheckResourceAttr(resourceName, "lc_collate", "en_US.UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "lc_ctype", "en_US.UTF-8"),
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
		configTerminationNil := testAccPGDatabaseWithDatasource(projectName, serviceName, dbName, "")
		configTerminationTrue := testAccPGDatabaseWithDatasource(projectName, serviceName, dbName, "  termination_protection = true")
		configTerminationFalse := testAccPGDatabaseWithDatasource(projectName, serviceName, dbName, "  termination_protection = false")
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenPGDatabaseResourceDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() { require.NoError(t, <-serviceIsReady) },
					Config:    configTerminationNil,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "database_name", dbName),
						resource.TestCheckResourceAttr(resourceName, "lc_collate", "en_US.UTF-8"),
						resource.TestCheckResourceAttr(resourceName, "lc_ctype", "en_US.UTF-8"),
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
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
					),
				},
				{
					// Fails to remove termination_protection=true resource
					Config:      configTerminationTrue,
					Destroy:     true,
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`The resource ` + "`aiven_pg_database`" + ` has termination protection enabled and\s+cannot be deleted`),
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
				{
					// Powers off the service to get ErrServicePoweredOff.
					Config: configTerminationNil,
					PreConfig: func() {
						err := servicePowerOn(t, projectName, serviceName, false)
						require.NoError(t, err)
					},
					ExpectError: regexp.MustCompile(schemautil.ErrServicePoweredOff.Error()),
				},
				{
					// Powers on the service: a database can't be destroyed while the service is powered off.
					Config: configTerminationNil,
					PreConfig: func() {
						err := servicePowerOn(t, projectName, serviceName, true)
						require.NoError(t, err)
					},
				},
				{
					// This resource has RemoveMissing set to true, so it should be recreated if it's missing (deleted).
					Config: testAccPGDatabase(projectName, serviceName, dbName),
					PreConfig: func() {
						err := client.ServiceDatabaseDelete(t.Context(), projectName, serviceName, dbName)
						require.NoError(t, err)
						// The provider caches the initial list of databases per service.
						// Clear it to mimic a fresh provider process after an out-of-band delete.
						schemautil.ForgetDatabase(projectName, serviceName, dbName)
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
					Config: testAccPGDatabase(projectName, serviceName, dbName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s", projectName, serviceName, dbName)),
						func(state *terraform.State) error {
							list, err := client.ServiceDatabaseList(t.Context(), projectName, serviceName)
							require.NoError(t, err)
							for _, db := range list {
								if db.DatabaseName == dbName {
									return nil
								}
							}
							return fmt.Errorf("pg database %q not found after recreation", dbName)
						},
					),
				},
			},
		})
	})
}

func servicePowerOn(t *testing.T, projectName, serviceName string, on bool) error {
	// Each test step should start with a clean service powered map
	schemautil.ServicePoweredForget(projectName, serviceName)

	client, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := t.Context()
	req := &service.ServiceUpdateIn{Powered: lo.ToPtr(on)}
	_, err = client.ServiceUpdate(ctx, projectName, serviceName, req)
	if err != nil || !on {
		// Powering off happens immediately, so we don't need to wait.
		return err
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s, err := client.ServiceGet(ctx, projectName, serviceName)
			if err != nil {
				return err
			}

			if s.State == service.ServiceStateTypeRunning {
				return nil
			}
		}
	}
}

func testAccCheckAivenPGDatabaseResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// loop through the resources in state, verifying each database is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_pg_database" {
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
				return fmt.Errorf("pg database %q still exists", databaseName)
			}
		}
	}

	return nil
}

func testAccPGDatabase(project, serviceName, dbName string) string {
	return fmt.Sprintf(`
resource "aiven_pg_database" "foo" {
  project       = %q
  service_name  = %q
  database_name = %q
}
`, project, serviceName, dbName)
}

func testAccPGDatabaseWithDatasource(project string, serviceName, dbName, extra string) string {
	return fmt.Sprintf(`
resource "aiven_pg_database" "foo" {
  project       = %q
  service_name  = %q
  database_name = %q
%s
}

data "aiven_pg_database" "foo" {
  project       = aiven_pg_database.foo.project
  service_name  = aiven_pg_database.foo.service_name
  database_name = aiven_pg_database.foo.database_name

  depends_on = [aiven_pg_database.foo]
}`, project, serviceName, dbName, extra)
}
