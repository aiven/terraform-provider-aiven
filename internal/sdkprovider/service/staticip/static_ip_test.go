package staticip_test

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenResourceStaticIp(t *testing.T) {
	resourceName := "aiven_static_ip.foo"
	projectName := acc.ProjectName()
	cloudName := "google-europe-west1"
	manifest := fmt.Sprintf(`
resource "aiven_static_ip" "foo" {
  project    = "%s"
  cloud_name = "%s"
}`, projectName, cloudName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
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
					_, ok = rs.Primary.Attributes["static_ip_address_id"]
					if !ok {
						return "", fmt.Errorf("expected resource '%s' to have an 'static_ip_address_id' attribute", resourceName)
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

func TestAccAivenResourceStaticIpNonExistentIdentifier(t *testing.T) {
	resourceName := "aiven_static_ip.foo"
	projectName := acc.ProjectName()
	cloudName := "google-europe-west1"
	manifest := fmt.Sprintf(`
resource "aiven_static_ip" "foo" {
  project    = "%s"
  cloud_name = "%s"
}`, projectName, cloudName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: manifest,
			},
			{
				Config:       manifest,
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return "non-existent/identifier", nil
				},
				ExpectError: regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}
