package serviceintegration_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenServiceIntegrationEndpoint_basic(t *testing.T) {
	resourceName := "aiven_service_integration_endpoint.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegraitonEndpointResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationEndpointBasicResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceEndpointIntegrationAttributes("data.aiven_service_integration_endpoint.endpoint"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", fmt.Sprintf("test-acc-ie-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "external_opensearch_logs"),
				),
			},
		},
	})
}

func TestAccAivenServiceIntegrationEndpoint_username_password(t *testing.T) {
	resourceName := "aiven_service_integration_endpoint.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegraitonEndpointResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationEndpointUsernamePasswordResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceEndpointIntegrationAttributes(
						"data.aiven_service_integration_endpoint.endpoint",
					),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(
						resourceName, "endpoint_name", fmt.Sprintf("test-acc-ie-%s", rName),
					),
					resource.TestCheckResourceAttr(
						resourceName, "endpoint_type", "external_schema_registry",
					),
				),
			},
			{
				Config: testAccServiceIntegrationEndpointUpdatePasswordResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceEndpointIntegrationAttributes(
						"data.aiven_service_integration_endpoint.endpoint",
					),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(
						resourceName, "endpoint_name", fmt.Sprintf("test-acc-ie-%s", rName),
					),
					resource.TestCheckResourceAttr(
						resourceName, "endpoint_type", "external_schema_registry",
					),
				),
			},
		},
	})
}

func testAccServiceIntegrationEndpointBasicResource(name string) string {
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
    index_days_max = 3
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
}`, acc.ProjectName(), name, name, name)
}

func testAccServiceIntegrationEndpointUsernamePasswordResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%[1]s"
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
}`, acc.ProjectName(), name)
}

func testAccServiceIntegrationEndpointUpdatePasswordResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%[1]s"
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
}`, acc.ProjectName(), name)
}

func testAccCheckAivenServiceIntegraitonEndpointResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each aiven_service_integration_endpoint is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_service_integration_endpoint" {
			continue
		}

		projectName, endpointID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		i, err := c.ServiceIntegrationEndpoints.Get(ctx, projectName, endpointID)
		if common.IsCritical(err) {
			return err
		}

		if i != nil {
			return fmt.Errorf("common integration endpoint(%s) still exists", rs.Primary.ID)
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

func TestAccAivenServiceIntegrationEndpointExternalPostgresql(t *testing.T) {
	resourceName := "aiven_service_integration_endpoint.pg"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegraitonEndpointResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAivenServiceIntegrationEndpointExternalPostgresql(rName),
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

func testAccAivenServiceIntegrationEndpointExternalPostgresql(name string) string {
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
`, acc.ProjectName(), name, name)
}
