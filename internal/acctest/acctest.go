package acctest

import (
	"fmt"
	"log"
	"os"
	"sort"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/provider"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	TestAccProviders         map[string]*schema.Provider
	TestAccProvider          *schema.Provider
	TestAccProviderFactories map[string]func() (*schema.Provider, error)
)

func init() {
	TestAccProvider = provider.Provider()
	TestAccProviders = map[string]*schema.Provider{
		"aiven": TestAccProvider,
	}
	TestAccProviderFactories = map[string]func() (*schema.Provider, error){
		"aiven": func() (*schema.Provider, error) {
			return TestAccProvider, nil
		},
	}
}

func TestAccPreCheck(t *testing.T) {
	if v := os.Getenv("AIVEN_TOKEN"); v == "" {
		t.Log(v)
		t.Fatal("AIVEN_TOKEN must be set for acceptance tests")
	}

	// Provider a project name with enough credits to run acceptance
	// tests or project name with the assigned payment card.
	if v := os.Getenv("AIVEN_PROJECT_NAME"); v == "" {
		log.Print("[WARNING] AIVEN_PROJECT_NAME must be set for some acceptance tests")
	}
}

func TestAccCheckAivenServiceResourceDestroy(s *terraform.State) error {
	c := TestAccProvider.Meta().(*aiven.Client)
	// loop through the resources in state, verifying each service is destroyed
	for _, rs := range s.RootModule().Resources {
		var r []string

		if sort.SearchStrings(r, rs.Type) > 0 {
			continue
		}

		if len(schemautil.SplitResourceID(rs.Primary.ID, 2)) == 2 {
			projectName, serviceName := schemautil.SplitResourceID2(rs.Primary.ID)
			p, err := c.Services.Get(projectName, serviceName)
			if err != nil {
				if err.(aiven.Error).Status != 404 {
					return err
				}
			}

			if p != nil {
				return fmt.Errorf("common (%s) still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}
