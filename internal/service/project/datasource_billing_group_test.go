package project_test

import (
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAivenBillingGroupDataSource_basic(t *testing.T) {
	datasourceName := "data.aiven_billing_group.bar"
	resourceName := "aiven_billing_group.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBillingGroupResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckTypeSetElemAttr(datasourceName, "billing_emails.*", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "billing_emails.*", datasourceName, "billing_emails.0"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "billing_emails.*", datasourceName, "billing_emails.1"),
				),
			},
		},
	})
}
