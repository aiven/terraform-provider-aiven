//go:build sweep
// +build sweep

package static_ip

import (
	"fmt"
	"os"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aiven_static_ip", &resource.Sweeper{
		Name: "aiven_static_ip",
		F:    sweepStaticIPs,
	})

}

func sweepStaticIPs(region string) error {
	client, err := sweep.SharedClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	projectName := os.Getenv("AIVEN_PROJECT_NAME")
	conn := client.(*aiven.Client)

	r, err := conn.StaticIPs.List(projectName)
	if err != nil {
		return fmt.Errorf("error retrieving a list of static_ips : %w", err)
	}

	for _, ip := range r.StaticIPs {
		err := conn.StaticIPs.Delete(
			projectName,
			aiven.DeleteStaticIPRequest{
				StaticIPAddressID: ip.StaticIPAddressID,
			})
		if err != nil && !aiven.IsNotFound(err) {
			return fmt.Errorf("error deleting stic_ip: %w", err)
		}
	}

	return nil
}
