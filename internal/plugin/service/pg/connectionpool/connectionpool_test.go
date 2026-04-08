package connectionpool_test

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

func TestAccAivenConnectionPool(t *testing.T) {
	projectName := acc.ProjectName()
	serviceName := acc.RandName("pool")
	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		serviceName,
		acc.WithServiceType("pg"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
	)

	t.Run("with user", func(t *testing.T) {
		resourceName := "aiven_connection_pool.foo"
		dataSourceName := "data.aiven_connection_pool.pool"
		databaseName := acc.RandName("db")
		userName := acc.RandName("user")
		poolName := acc.RandName("pool")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenConnectionPoolResourceDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccConnectionPoolResource(projectName, serviceName, databaseName, userName, poolName, 25),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "database_name", databaseName),
						resource.TestCheckResourceAttr(resourceName, "username", userName),
						resource.TestCheckResourceAttr(resourceName, "pool_name", poolName),
						resource.TestCheckResourceAttr(resourceName, "pool_size", "25"),
						resource.TestCheckResourceAttr(resourceName, "pool_mode", "transaction"),
						resource.TestCheckResourceAttr(dataSourceName, "project", projectName),
						resource.TestCheckResourceAttr(dataSourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(dataSourceName, "database_name", databaseName),
						resource.TestCheckResourceAttr(dataSourceName, "username", userName),
						resource.TestCheckResourceAttr(dataSourceName, "pool_name", poolName),
						resource.TestCheckResourceAttr(dataSourceName, "pool_size", "25"),
						resource.TestCheckResourceAttr(dataSourceName, "pool_mode", "transaction"),
					),
				},
				{
					Config: testAccConnectionPoolResource(projectName, serviceName, databaseName, userName, poolName, 42),
					Check:  resource.TestCheckResourceAttr(resourceName, "pool_size", "42"),
				},
			},
		})
	})

	t.Run("without user", func(t *testing.T) {
		resourceName := "aiven_connection_pool.foo"
		databaseName := acc.RandName("db")
		poolName := acc.RandName("pool")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenConnectionPoolResourceDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccConnectionPoolNoUserResource(projectName, serviceName, databaseName, poolName, 25),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "database_name", databaseName),
						resource.TestCheckResourceAttr(resourceName, "pool_name", poolName),
						resource.TestCheckResourceAttr(resourceName, "pool_size", "25"),
						resource.TestCheckResourceAttr(resourceName, "pool_mode", "transaction"),
					),
				},
				{
					Config: testAccConnectionPoolNoUserResource(projectName, serviceName, databaseName, poolName, 42),
					Check:  resource.TestCheckResourceAttr(resourceName, "pool_size", "42"),
				},
			},
		})
	})

	const oldProviderVersion = "4.53.0"

	t.Run("with user (backward compatibility)", func(t *testing.T) {
		resourceName := "aiven_connection_pool.foo"
		databaseName := acc.RandName("db")
		userName := acc.RandName("user")
		poolName := acc.RandName("pool")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() { acc.TestAccPreCheck(t) },
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				TFConfig: testAccConnectionPoolResource(projectName, serviceName, databaseName, userName, poolName, 25),
				PreConfig: func() {
					require.NoError(t, <-serviceIsReady)
				},
				OldProviderVersion: oldProviderVersion,
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", databaseName),
					resource.TestCheckResourceAttr(resourceName, "username", userName),
					resource.TestCheckResourceAttr(resourceName, "pool_name", poolName),
					resource.TestCheckResourceAttr(resourceName, "pool_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "pool_mode", "transaction"),
				),
			}),
		})
	})

	t.Run("without user (backward compatibility)", func(t *testing.T) {
		resourceName := "aiven_connection_pool.foo"
		databaseName := acc.RandName("db")
		poolName := acc.RandName("pool")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() { acc.TestAccPreCheck(t) },
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				TFConfig: testAccConnectionPoolNoUserResource(projectName, serviceName, databaseName, poolName, 25),
				PreConfig: func() {
					require.NoError(t, <-serviceIsReady)
				},
				OldProviderVersion: oldProviderVersion,
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", databaseName),
					resource.TestCheckResourceAttr(resourceName, "pool_name", poolName),
					resource.TestCheckResourceAttr(resourceName, "pool_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "pool_mode", "transaction"),
				),
			}),
		})
	})
}

func testAccConnectionPoolNoUserResource(projectName, serviceName, databaseName, poolName string, poolSize int) string {
	return fmt.Sprintf(`
resource "aiven_pg_database" "foo" {
  project       = %[1]q
  service_name  = %[2]q
  database_name = %[3]q
}

resource "aiven_connection_pool" "foo" {
  project       = %[1]q
  service_name  = %[2]q
  database_name = aiven_pg_database.foo.database_name
  pool_name     = %[4]q
  pool_size     = %[5]d
  pool_mode     = "transaction"
}

data "aiven_connection_pool" "pool" {
  project      = aiven_connection_pool.foo.project
  service_name = aiven_connection_pool.foo.service_name
  pool_name    = aiven_connection_pool.foo.pool_name
}`, projectName, serviceName, databaseName, poolName, poolSize)
}

func testAccConnectionPoolResource(projectName, serviceName, databaseName, username, poolName string, poolSize int) string {
	return fmt.Sprintf(`
resource "aiven_pg_user" "foo" {
  service_name = %[2]q
  project      = %[1]q
  username     = %[5]q
}

resource "aiven_pg_database" "foo" {
  project       = %[1]q
  service_name  = %[2]q
  database_name = %[3]q
}

resource "aiven_connection_pool" "foo" {
  project       = %[1]q
  service_name  = %[2]q
  database_name = aiven_pg_database.foo.database_name
  username      = aiven_pg_user.foo.username
  pool_name     = %[4]q
  pool_size     = %[6]d
  pool_mode     = "transaction"
}

data "aiven_connection_pool" "pool" {
  project      = aiven_connection_pool.foo.project
  service_name = aiven_connection_pool.foo.service_name
  pool_name    = aiven_connection_pool.foo.pool_name
}`, projectName, serviceName, databaseName, poolName, username, poolSize)
}

func testAccCheckAivenConnectionPoolResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_connection_pool" {
			continue
		}

		projectName, serviceName, poolName, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		rsp, err := c.ServiceGet(ctx, projectName, serviceName)
		if err != nil {
			if avngen.IsNotFound(err) {
				continue
			}
			return err
		}

		for _, p := range rsp.ConnectionPools {
			if p.PoolName == poolName {
				return fmt.Errorf("connection pool (%[1]q) still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}
