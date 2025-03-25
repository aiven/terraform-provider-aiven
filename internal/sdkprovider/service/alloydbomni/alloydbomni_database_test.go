package alloydbomni_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenAlloyDBOmniDatabase_basic(t *testing.T) {
	resourceName := "aiven_alloydbomni_database.foo"
	projectName := acc.ProjectName()
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenAlloyDBOmniDatabaseResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlloyDBOmniDatabaseResource(projectName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "database_name", fmt.Sprintf("test-acc-db-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "lc_ctype", "en_US.UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "lc_collate", "en_US.UTF-8"),
				),
			},
			{
				Config:       testAccAlloyDBOmniDatabaseResource(projectName, rName),
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
						return fmt.Errorf("expected project to match '%s', got: '%s'", projectName, attributes["project_name"])
					}
					databaseName, ok := attributes["database_name"]
					if !ok {
						return errors.New("expected 'database_name' field to be set")
					}
					if _, ok := attributes["lc_ctype"]; !ok {
						return errors.New("expected 'lc_ctype' field to be set")
					}
					if _, ok := attributes["lc_collate"]; !ok {
						return errors.New("expected 'lc_collate' field to be set")
					}
					expectedID := fmt.Sprintf("%s/test-acc-sr-%s/%s", projectName, rName, databaseName)
					if !strings.EqualFold(s[0].ID, expectedID) {
						return fmt.Errorf("expected ID to match '%s', but got: %s", expectedID, s[0].ID)
					}
					return nil
				},
			},
		},
	})
}

func testAccCheckAivenAlloyDBOmniDatabaseResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each database is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_alloydbomni_database" {
			continue
		}

		projectName, serviceName, databaseName, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		db, err := c.Databases.Get(ctx, projectName, serviceName, databaseName)
		if err != nil {
			var e aiven.Error
			if errors.As(err, &e) && e.Status != 404 {
				return err
			}
		}

		if db != nil {
			return fmt.Errorf("database (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAlloyDBOmniDatabaseResource(project string, name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_alloydbomni" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  alloydbomni_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
    }
  }
}

resource "aiven_alloydbomni_database" "foo" {
  project       = aiven_alloydbomni.bar.project
  service_name  = aiven_alloydbomni.bar.service_name
  database_name = "test-acc-db-%s"
  lc_ctype      = "en_US.UTF-8"
  lc_collate    = "en_US.UTF-8"
}

data "aiven_alloydbomni_database" "database" {
  project       = aiven_alloydbomni_database.foo.project
  service_name  = aiven_alloydbomni_database.foo.service_name
  database_name = aiven_alloydbomni_database.foo.database_name

  depends_on = [aiven_alloydbomni_database.foo]
}`, project, name, name)
}
