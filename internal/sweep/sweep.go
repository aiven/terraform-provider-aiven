//go:build sweep
// +build sweep

package sweep

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/aiven/aiven-go-client"
)

var sharedClient *aiven.Client

// sharedClient returns a common Aiven Client setup needed for the sweeper
func SharedClient(region string) (interface{}, error) {
	if os.Getenv("AIVEN_TOKEN") == "" {
		return nil, fmt.Errorf("must provide environment variable AIVEN_TOKEN ")
	}
	if os.Getenv("AIVEN_PROJECT_NAME") == "" {
		return nil, fmt.Errorf("must provide environment variable AIVEN_PROJECT_NAME ")
	}

	if sharedClient == nil {
		// configures a default client, using the above env var
		var err error
		sharedClient, err = aiven.NewTokenClient(os.Getenv("AIVEN_TOKEN"), "terraform-provider-aiven-acc/")
		if err != nil {
			return nil, fmt.Errorf("error getting Aiven client")
		}
	}

	return sharedClient, nil
}

func SweepServices(region, t string) error {
	client, err := SharedClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	projectName := os.Getenv("AIVEN_PROJECT_NAME")
	conn := client.(*aiven.Client)

	services, err := conn.Services.List(projectName)
	if err != nil && !aiven.IsNotFound(err) {
		return fmt.Errorf("error retrieving a list of services for a project `%s`: %w", projectName, err)
	}

	for _, s := range services {
		if s.Type != t {
			continue
		}

		if !hasPrefixAny(s.Name, "test-acc", "test-examples", "k8s-") {
			continue
		}

		// if service termination_protection is on service cannot be deleted
		// update service and turn termination_protection off
		if s.TerminationProtection {
			_, err := conn.Services.Update(projectName, s.Name, aiven.UpdateServiceRequest{
				Cloud:                 s.CloudName,
				MaintenanceWindow:     &s.MaintenanceWindow,
				Plan:                  s.Plan,
				ProjectVPCID:          s.ProjectVPCID,
				Powered:               true,
				TerminationProtection: false,
				UserConfig:            s.UserConfig,
			})

			if err != nil {
				return fmt.Errorf("error disabling `termination_protection` for service '%s' during sweep: %s", s.Name, err)
			}
		}

		if err := conn.Services.Delete(projectName, s.Name); err != nil {
			if err != nil && !aiven.IsNotFound(err) {
				return fmt.Errorf("error destroying service %s during sweep: %s", s.Name, err)
			}
		}
	}
	return nil
}

func AddServiceSweeper(t string) {
	resource.AddTestSweepers("aiven_"+t, &resource.Sweeper{
		Name: "aiven_" + t,
		F: func(r string) error {
			return SweepServices(r, t)
		},
		Dependencies: []string{"aiven_service_integration"},
	})
}

func hasPrefixAny(s string, prefix ...string) bool {
	for _, ss := range prefix {
		if strings.HasPrefix(s, ss) {
			return true
		}
	}
	return false
}
