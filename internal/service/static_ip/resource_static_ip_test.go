package static_ip_test

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAivenResourceStaticIp(t *testing.T) {
	resourceName := "aiven_static_ip.foo"
	projectName := os.Getenv("AIVEN_PROJECT_NAME")
	cloudName := "google-europe-west1"
	manifest := fmt.Sprintf(`
			resource "aiven_static_ip" "foo" {
			  project    = "%s"
			  cloud_name = "%s"
			}`,
		projectName, cloudName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: manifest,
			},
			{
				Config:       manifest,
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("expected resource '%s' to be present in the state", resourceName)
					}
					id, ok := rs.Primary.Attributes["static_ip_address_id"]
					if !ok {
						return "", fmt.Errorf("expected resource '%s' to have an 'static_ip_address_id' attribute", resourceName)
					}
					return fmt.Sprintf("%s/%s", projectName, id), nil
				},
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					if len(s) != 1 {
						return fmt.Errorf("expected only one instance to be imported, state: %#v", s)
					}
					attributes := s[0].Attributes
					if !strings.EqualFold(attributes["project"], projectName) {
						return fmt.Errorf("expected project to match '%s', got: '%s'", projectName, attributes["project_name"])
					}
					if !strings.EqualFold(attributes["cloud_name"], cloudName) {
						return fmt.Errorf("expected cloud to match '%s', got: '%s'", cloudName, attributes["cloud_name"])
					}
					if _, ok := attributes["static_ip_address_id"]; !ok {
						return errors.New("expected 'static_ip_address_id' field to be set")
					}
					if _, ok := attributes["state"]; !ok {
						return errors.New("expected 'state' field to be set")
					}
					return nil
				},
			},
		},
	})
}
