package cmk_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

func TestAccCMKResource(t *testing.T) {
	resName := os.Getenv("AIVEN_CMK_NAME")
	if resName == "" {
		t.Skip("AIVEN_CMK_NAME is not set")
	}
	project := acc.ProjectName()
	resourceName := "aiven_cmk.foo"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccCMKResource(project, resName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "cmk_provider", "gcp"),
					resource.TestCheckResourceAttr(resourceName, "resource", resName),
					resource.TestCheckResourceAttr(resourceName, "default_cmk", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", "current"),
					resource.TestCheckResourceAttrSet(resourceName, "cmk_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, err := acc.ResourceFromState(state, resourceName)
					if err != nil {
						return "", err
					}

					return util.ComposeID(rs.Primary.Attributes["project"], rs.Primary.Attributes["cmk_id"]), nil
				},
			},
		},
	})
}

func testAccCMKResource(projectName, resName string) string {
	return fmt.Sprintf(`
resource "aiven_cmk" "foo" {
  project      = %q
  resource     = %q
  cmk_provider = "gcp"
  default_cmk  = false
}
`, projectName, resName)
}
