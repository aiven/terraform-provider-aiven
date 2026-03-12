package staticip_test

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aiven/go-client-codegen/handler/staticip"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenResourceStaticIp(t *testing.T) {
	projectName := acc.ProjectName()
	resourceName := "aiven_static_ip.foo"
	cloudName := "google-europe-west1"

	manifest := fmt.Sprintf(`
resource "aiven_static_ip" "foo" {
  project    = "%s"
  cloud_name = "%s"
}`, projectName, cloudName)

	manifestWithTerminationProtection := fmt.Sprintf(`
resource "aiven_static_ip" "foo" {
  project                = "%s"
  cloud_name             = "%s"
  termination_protection = true
}`, projectName, cloudName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: manifest,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", cloudName),
					resource.TestCheckResourceAttrSet(resourceName, "static_ip_address_id"),
					resource.TestCheckResourceAttr(resourceName, "state", string(staticip.StaticIPStateTypeCreated)),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				Config: manifestWithTerminationProtection,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
				),
			},
			{
				Config:      manifestWithTerminationProtection,
				Destroy:     true,
				ExpectError: regexp.MustCompile(`The resource ` + "`aiven_static_ip`" + ` has termination protection enabled`),
			},
			{
				// Removing termination protection goes true -> false.
				Config: manifest,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
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

func TestAccAivenStaticIP_backwardCompat(t *testing.T) {
	resourceName := "aiven_static_ip.foo"
	projectName := acc.ProjectName()
	cloudName := "google-europe-west1"
	config := fmt.Sprintf(`
resource "aiven_static_ip" "foo" {
  project    = "%s"
  cloud_name = "%s"
}`, projectName, cloudName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acc.TestAccPreCheck(t) },
		Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
			TFConfig:           config,
			OldProviderVersion: "4.47.0",
			Checks: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "project", projectName),
				resource.TestCheckResourceAttr(resourceName, "cloud_name", cloudName),
				resource.TestCheckResourceAttrSet(resourceName, "static_ip_address_id"),
				resource.TestCheckResourceAttrSet(resourceName, "state"),
			),
		}),
	})
}
