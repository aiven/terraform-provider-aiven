package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
)

func init() {
	resource.AddTestSweepers("aiven_service", &resource.Sweeper{
		Name: "aiven_service",
		F:    sweepServices,
	})
}

func sweepServices(region string) error {
	client, err := sharedClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*aiven.Client)

	projects, err := conn.Projects.List()
	if err != nil {
		return fmt.Errorf("error retrieving a list of projects : %s", err)
	}

	for _, project := range projects {
		if strings.Contains(project.Name, "test-acc-") {
			services, err := conn.Services.List(project.Name)
			if err != nil {
				return fmt.Errorf("error retrieving a list of services for a project `%s`: %s", project.Name, err)
			}

			for _, service := range services {
				if err := conn.Services.Delete(project.Name, service.Name); err != nil {
					return fmt.Errorf("error destroying service %s during sweep: %s", service.Name, err)
				}
			}
		}
	}

	// sweep projects
	if err := sweepProjects(region); err != nil {
		return err
	}

	return nil
}

// Elasticsearch service tests
func TestAccAivenService_es(t *testing.T) {
	t.Parallel()

	resourceName := "aiven_service.bar-es"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceCommonAttributes("data.aiven_service.service-es"),
					testAccCheckAivenServiceESAttributes("data.aiven_service.service-es"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-es-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "service_type", "elasticsearch"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
				),
			},
		},
	})
}

// PG service tests
func TestAccAivenService_pg(t *testing.T) {
	t.Parallel()
	resourceName := "aiven_service.bar-pg"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceCommonAttributes("data.aiven_service.service-pg"),
					testAccCheckAivenServicePGAttributes("data.aiven_service.service-pg"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-pg-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "service_type", "pg"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
				),
			},
		},
	})
}

func testAccElasticsearchServiceResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo-es" {
			project = "test-acc-pr-es-%s"
			card_id="%s"	
		}
		
		resource "aiven_service" "bar-es" {
			project = aiven_project.foo-es.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-%s"
			service_type = "elasticsearch"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			elasticsearch_user_config {
				elasticsearch_version = 7

			}
		}
		
		data "aiven_service" "service-es" {
			service_name = aiven_service.bar-es.service_name
			project = aiven_project.foo-es.project
		}
		`, name, os.Getenv("AIVEN_CARD_ID"), name)
}

func testAccPGServiceResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo-pg" {
			project = "test-acc-pr-pg-%s"
			card_id="%s"	
		}
		
		resource "aiven_service" "bar-pg" {
			project = aiven_project.foo-pg.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-%s"
			service_type = "pg"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			pg_user_config {
				pg_version = 11
			}
		}
		
		data "aiven_service" "service-pg" {
			service_name = aiven_service.bar-pg.service_name
			project = aiven_project.foo-pg.project
		}
		`, name, os.Getenv("AIVEN_CARD_ID"), name)
}

func testAccCheckAivenServiceESAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if !strings.Contains(a["service_type"], "elasticsearch") {
			return fmt.Errorf("expected to get a correct service type from Aiven, got :%s", a["service_type"])
		}

		return nil
	}
}
func testAccCheckAivenServicePGAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if !strings.Contains(a["service_type"], "pg") {
			return fmt.Errorf("expected to get a correct service type from Aiven, got :%s", a["service_type"])
		}

		return nil
	}
}

func testAccCheckAivenServiceCommonAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] service attributes %v", a)

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service name from Aiven")
		}

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["service_uri"] == "" {
			return fmt.Errorf("expected to get a service_uri from Aiven")
		}

		if a["maintenance_window_dow"] != "monday" {
			return fmt.Errorf("expected to get a service.maintenance_window_dow from Aiven")
		}

		if a["service_password"] == "" {
			return fmt.Errorf("expected to get a service_password from Aiven")
		}

		if a["service_port"] == "" {
			return fmt.Errorf("expected to get a service_port from Aiven")
		}

		if a["service_type"] == "" {
			return fmt.Errorf("expected to get a service_type from Aiven")
		}

		if a["service_username"] == "" {
			return fmt.Errorf("expected to get a service_username from Aiven")
		}

		if a["maintenance_window_time"] != "10:00:00" {
			return fmt.Errorf("expected to get a service.maintenance_window_time from Aiven")
		}

		return nil
	}
}

func testAccCheckAivenServiceResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each aiven_service is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_service" {
			continue
		}

		projectName, serviceName := splitResourceID2(rs.Primary.ID)
		p, err := c.Services.Get(projectName, serviceName)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}
		}

		if p != nil {
			return fmt.Errorf("service (%s) still exists", rs.Primary.ID)
		}
	}

	// check if all projects were destroyed as well
	if err := testAccCheckAivenProjectResourceDestroy(s); err != nil {
		return err
	}

	return nil
}

func Test_flattenServiceComponents(t *testing.T) {
	type args struct {
		r *aiven.Service
	}
	tests := []struct {
		name string
		args args
		want []map[string]interface{}
	}{
		{
			"",
			args{r: &aiven.Service{
				Components: []*aiven.ServiceComponents{
					{
						Component: "grafana",
						Host:      "aive-public-grafana.aiven.io",
						Port:      433,
						Route:     "public",
						Usage:     "primary",
					},
				},
			}},
			[]map[string]interface{}{
				{
					"component": "grafana",
					"host":      "aive-public-grafana.aiven.io",
					"port":      433,
					"route":     "public",
					"usage":     "primary",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flattenServiceComponents(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("flattenServiceComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}
