package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func TestAccAivenAccountDataSource_basic(t *testing.T) {
	datasourceName := "data.aiven_account.account"
	resourceName := "aiven_account.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "account_id", resourceName, "account_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "owner_team_id", resourceName, "owner_team_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "tenant_id", resourceName, "tenant_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "create_time", resourceName, "create_time"),
					resource.TestCheckResourceAttrPair(datasourceName, "update_time", resourceName, "update_time"),
				),
			},
		},
	})
}
