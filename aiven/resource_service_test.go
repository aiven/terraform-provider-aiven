package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
				// if service termination_protection is on service cannot be deleted
				// update service and turn termination_protection off
				if service.TerminationProtection == true {
					_, err := conn.Services.Update(project.Name, service.Name, aiven.UpdateServiceRequest{
						Cloud:                 service.CloudName,
						MaintenanceWindow:     &service.MaintenanceWindow,
						Plan:                  service.Plan,
						ProjectVPCID:          service.ProjectVPCID,
						Powered:               true,
						TerminationProtection: false,
						UserConfig:            service.UserConfig,
					})

					if err != nil {
						return fmt.Errorf("error destroying service %s during sweep, disabling `termination_protection`: %s", service.Name, err)
					}
				}

				if err := conn.Services.Delete(project.Name, service.Name); err != nil {
					if err.(aiven.Error).Status != 404 {
						return fmt.Errorf("error destroying service %s during sweep: %s", service.Name, err)
					}
				}
			}
		}
	}

	return nil
}

func testAccCheckAivenServiceCommonAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["cloud_name"] == "" {
			return fmt.Errorf("expected to get a cloud_name from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service name from Aiven")
		}

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["plan"] == "" {
			return fmt.Errorf("expected to get a plan from Aiven")
		}

		if a["service_uri"] == "" {
			return fmt.Errorf("expected to get a service_uri from Aiven")
		}

		if a["maintenance_window_dow"] != "monday" {
			return fmt.Errorf("expected to get a service.maintenance_window_dow from Aiven")
		}

		// Kafka service has no username and password
		if a["service_type"] != "kafka" {
			if a["service_password"] == "" {
				return fmt.Errorf("expected to get a service_password from Aiven")
			}

			if a["service_username"] == "" {
				return fmt.Errorf("expected to get a service_username from Aiven")
			}
		}

		if a["service_port"] == "" {
			return fmt.Errorf("expected to get a service_port from Aiven")
		}

		if a["service_host"] == "" {
			return fmt.Errorf("expected to get a service_host from Aiven")
		}

		if a["service_type"] == "" {
			return fmt.Errorf("expected to get a service_type from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service_name from Aiven")
		}

		if a["state"] != "RUNNING" {
			return fmt.Errorf("expected to get a correct state from Aiven")
		}

		if a["maintenance_window_time"] != "10:00:00" {
			return fmt.Errorf("expected to get a service.maintenance_window_time from Aiven")
		}

		if a["components.#"] == "0" {
			return fmt.Errorf("expected to get a none empty list of components from Aiven")
		}

		if a["termination_protection"] == "" {
			return fmt.Errorf("expected to get a termination_protection from Aiven")
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
