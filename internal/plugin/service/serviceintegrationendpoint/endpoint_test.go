package serviceintegrationendpoint_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenServiceIntegrationEndpoint_basic(t *testing.T) {
	resourceName := "aiven_service_integration_endpoint.bar"
	dataSourceName := "data.aiven_service_integration_endpoint.endpoint"
	suffix := acc.RandStr()
	endpointName := fmt.Sprintf("test-acc-ie-%s", suffix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegraitonEndpointResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationEndpointBasicResource(suffix, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceEndpointIntegrationAttributes(dataSourceName),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", endpointName),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "external_opensearch_logs"),
					resource.TestCheckResourceAttr(resourceName, "external_opensearch_logs_user_config.0.index_days_max", "3"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint_id", resourceName, "endpoint_id"),
				),
			},
			{
				Config: testAccServiceIntegrationEndpointBasicResource(suffix, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceEndpointIntegrationAttributes(dataSourceName),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", endpointName),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "external_opensearch_logs"),
					resource.TestCheckResourceAttr(resourceName, "external_opensearch_logs_user_config.0.index_days_max", "5"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint_id", resourceName, "endpoint_id"),
				),
			},
			{
				Config:            testAccServiceIntegrationEndpointBasicResource(suffix, 5),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Import doesn't read user config blocks: they are kept only when configured
				ImportStateVerifyIgnore: []string{"external_opensearch_logs_user_config"},
			},
		},
	})
}

func TestAccAivenServiceIntegrationEndpoint_username_password(t *testing.T) {
	resourceName := "aiven_service_integration_endpoint.bar"
	suffix := acc.RandStr()
	endpointName := fmt.Sprintf("test-acc-ie-%s", suffix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegraitonEndpointResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationEndpointUsernamePasswordResource(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceEndpointIntegrationAttributes(
						"data.aiven_service_integration_endpoint.endpoint",
					),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", endpointName),
					resource.TestCheckResourceAttr(
						resourceName, "endpoint_type", "external_schema_registry",
					),
				),
			},
			{
				Config: testAccServiceIntegrationEndpointUpdatePasswordResource(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceEndpointIntegrationAttributes(
						"data.aiven_service_integration_endpoint.endpoint",
					),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", endpointName),
					resource.TestCheckResourceAttr(
						resourceName, "endpoint_type", "external_schema_registry",
					),
				),
			},
		},
	})
}

func TestAccAivenServiceIntegrationEndpointExternalPostgresql(t *testing.T) {
	resourceName := "aiven_service_integration_endpoint.pg"
	suffix := acc.RandStr()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegraitonEndpointResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAivenServiceIntegrationEndpointExternalPostgresql(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceEndpointIntegrationAttributes(resourceName),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "external_postgresql"),
					resource.TestCheckResourceAttr(resourceName, "external_postgresql.0.port", "1234"),
					resource.TestCheckResourceAttr(resourceName, "external_postgresql.0.ssl_mode", "require"),
				),
			},
		},
	})
}

func TestAccAivenServiceIntegrationEndpoint_backwardCompat(t *testing.T) {
	resourceName := "aiven_service_integration_endpoint.bar"
	dataSourceName := "data.aiven_service_integration_endpoint.endpoint"
	suffix := acc.RandStr()
	endpointName := fmt.Sprintf("test-acc-ie-%s", suffix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acc.TestAccPreCheck(t) },
		Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
			TFConfig:           testAccServiceIntegrationEndpointBackwardCompat(suffix),
			OldProviderVersion: "4.65.0",
			Checks: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
				resource.TestCheckResourceAttr(resourceName, "endpoint_name", endpointName),
				resource.TestCheckResourceAttr(resourceName, "endpoint_type", "external_opensearch_logs"),
				resource.TestCheckResourceAttr(resourceName, "external_opensearch_logs_user_config.0.index_days_max", "3"),
				resource.TestCheckResourceAttr(dataSourceName, "project", acc.ProjectName()),
				resource.TestCheckResourceAttr(dataSourceName, "endpoint_name", endpointName),
				resource.TestCheckResourceAttr(dataSourceName, "endpoint_type", "external_opensearch_logs"),
				resource.TestCheckResourceAttr(dataSourceName, "external_opensearch_logs_user_config.0.index_days_max", "3"),
			),
		}),
	})
}

// The block keeps its SDKv2 name, see adapter.userConfigBlock.
func TestAccAivenServiceIntegrationEndpointExternalPostgresql_backwardCompat(t *testing.T) {
	resourceName := "aiven_service_integration_endpoint.pg"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acc.TestAccPreCheck(t) },
		Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
			TFConfig:           testAccAivenServiceIntegrationEndpointExternalPostgresql(acc.RandStr()),
			OldProviderVersion: "4.65.0",
			PlanOnly:           true,
			Checks: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "endpoint_type", "external_postgresql"),
				resource.TestCheckResourceAttr(resourceName, "external_postgresql.0.port", "1234"),
				resource.TestCheckResourceAttr(resourceName, "external_postgresql.0.ssl_mode", "require"),
			),
		}),
	})
}

func testAccServiceIntegrationEndpointBasicResource(suffix string, indexDaysMax int) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "bar-pg" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-pg-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  pg_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
    }
  }
}

resource "aiven_service_integration_endpoint" "bar" {
  project       = data.aiven_project.foo.project
  endpoint_name = "test-acc-ie-%s"
  endpoint_type = "external_opensearch_logs"

  external_opensearch_logs_user_config {
    url            = "https://user:passwd@logs.example.com/"
    index_prefix   = "test-acc-prefix-%s"
    index_days_max = %d
    timeout        = 10
  }
}

resource "aiven_service_integration" "bar" {
  project                 = data.aiven_project.foo.project
  integration_type        = "external_opensearch_logs"
  source_service_name     = aiven_pg.bar-pg.service_name
  destination_endpoint_id = aiven_service_integration_endpoint.bar.id
}

data "aiven_service_integration_endpoint" "endpoint" {
  project       = aiven_service_integration_endpoint.bar.project
  endpoint_name = aiven_service_integration_endpoint.bar.endpoint_name

  depends_on = [aiven_service_integration_endpoint.bar]
}`, acc.ProjectName(), suffix, suffix, suffix, indexDaysMax)
}

func testAccServiceIntegrationEndpointBackwardCompat(suffix string) string {
	return fmt.Sprintf(`
resource "aiven_service_integration_endpoint" "bar" {
  project       = %q
  endpoint_name = "test-acc-ie-%s"
  endpoint_type = "external_opensearch_logs"

  external_opensearch_logs_user_config {
    url            = "https://user:passwd@logs.example.com/"
    index_prefix   = "test-acc-prefix-%s"
    index_days_max = 3
    timeout        = 10
  }
}

data "aiven_service_integration_endpoint" "endpoint" {
  project       = aiven_service_integration_endpoint.bar.project
  endpoint_name = aiven_service_integration_endpoint.bar.endpoint_name

  depends_on = [aiven_service_integration_endpoint.bar]
}`, acc.ProjectName(), suffix, suffix)
}

func testAccServiceIntegrationEndpointUsernamePasswordResource(suffix string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = %[1]q
}

resource "aiven_pg" "bar-pg" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-pg-%[2]s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  pg_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
    }
  }
}

resource "aiven_service_integration_endpoint" "bar" {
  project       = data.aiven_project.foo.project
  endpoint_name = "test-acc-ie-%[2]s"
  endpoint_type = "external_schema_registry"

  external_schema_registry_user_config {
    url = "https://schema-registry.example.com:8081"

    authentication      = "basic"
    basic_auth_username = "username"
    basic_auth_password = "password"
  }
}

data "aiven_service_integration_endpoint" "endpoint" {
  project       = aiven_service_integration_endpoint.bar.project
  endpoint_name = aiven_service_integration_endpoint.bar.endpoint_name

  depends_on = [aiven_service_integration_endpoint.bar]
}`, acc.ProjectName(), suffix)
}

func testAccServiceIntegrationEndpointUpdatePasswordResource(suffix string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = %[1]q
}

resource "aiven_pg" "bar-pg" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-pg-%[2]s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  pg_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
    }
  }
}

resource "aiven_service_integration_endpoint" "bar" {
  project       = data.aiven_project.foo.project
  endpoint_name = "test-acc-ie-%[2]s"
  endpoint_type = "external_schema_registry"

  external_schema_registry_user_config {
    url = "https://schema-registry.example.com:8081"

    authentication      = "basic"
    basic_auth_username = "username"
    basic_auth_password = "new-password"
  }
}

data "aiven_service_integration_endpoint" "endpoint" {
  project       = aiven_service_integration_endpoint.bar.project
  endpoint_name = aiven_service_integration_endpoint.bar.endpoint_name

  depends_on = [aiven_service_integration_endpoint.bar]
}`, acc.ProjectName(), suffix)
}

func testAccCheckAivenServiceIntegraitonEndpointResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_service_integration_endpoint" {
			continue
		}

		projectName, endpointID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		// The API responds to a missing endpoint's Get with 400, so the list is checked
		list, err := c.ServiceIntegrationEndpointList(ctx, projectName)
		if err != nil {
			return err
		}

		for _, e := range list {
			if e.EndpointId == endpointID {
				return fmt.Errorf("service integration endpoint(%s) still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAivenServiceEndpointIntegrationAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project from Aiven")
		}

		if a["endpoint_name"] == "" {
			return fmt.Errorf("expected to get a endpoint_name from Aiven")
		}

		if a["endpoint_type"] == "" {
			return fmt.Errorf("expected to get an endpoint_type from Aiven")
		}

		return nil
	}
}

func testAccAivenServiceIntegrationEndpointExternalPostgresql(suffix string) string {
	return fmt.Sprintf(`
resource "aiven_pg" "pg" {
  project      = %q
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "test-acc-sr-pg-%s"
}

resource "aiven_service_integration_endpoint" "pg" {
  project       = aiven_pg.pg.project
  endpoint_name = "test-acc-external-postgresql-%s"
  endpoint_type = "external_postgresql"

  external_postgresql {
    username = aiven_pg.pg.service_username
    password = aiven_pg.pg.service_password
    host     = aiven_pg.pg.service_host
    port     = 1234
    ssl_mode = "require"
  }
}
`, acc.ProjectName(), suffix, suffix)
}
