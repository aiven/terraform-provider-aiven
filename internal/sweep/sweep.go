//go:build sweep

package sweep

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
)

var sharedClient *aiven.Client

// SharedClient returns a common Aiven Client setup needed for the sweeper
func SharedClient() (*aiven.Client, error) {
	if os.Getenv("AIVEN_PROJECT_NAME") == "" {
		return nil, fmt.Errorf("must provide environment variable AIVEN_PROJECT_NAME ")
	}

	if sharedClient == nil {
		// configures a default client, using the above env var
		var err error
		sharedClient, err = common.NewAivenClient()
		if err != nil {
			return nil, fmt.Errorf("error getting Aiven client")
		}
	}

	return sharedClient, nil
}

func SweepServices(ctx context.Context, t string) error {
	client, err := SharedClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	projectName := os.Getenv("AIVEN_PROJECT_NAME")

	services, err := client.Services.List(ctx, projectName)
	if err != nil && !aiven.IsNotFound(err) {
		return fmt.Errorf("error retrieving a list of services for a project `%s`: %w", projectName, err)
	}

	for _, s := range services {
		if s.Type != t {
			continue
		}

		if !hasPrefixAny(s.Name, "test-tf", "test-acc", "test-examples", "k8s-") {
			continue
		}

		// if service termination_protection is on service cannot be deleted
		// update service and turn termination_protection off
		if s.TerminationProtection {
			_, err := client.Services.Update(ctx, projectName, s.Name, aiven.UpdateServiceRequest{
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

		if err := client.Services.Delete(ctx, projectName, s.Name); err != nil {
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
		F: func(_ string) error {
			return SweepServices(context.Background(), t)
		},
		Dependencies: []string{"aiven_service_integration", "aiven_service_integration_endpoint"},
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
