package paymentmethodlist_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenOrganizationPaymentMethodListDataSource(t *testing.T) {
	acc.SkipIfNotBeta(t)

	var (
		organizationName = acc.OrganizationName()
		dataSourceName   = "data.aiven_organization_payment_method_list.ds"
	)

	config := fmt.Sprintf(`
data "aiven_organization" "org" {
  name = %q
}

data "aiven_organization_payment_method_list" "ds" {
  organization_id = data.aiven_organization.org.id
}`, organizationName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "payment_methods.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "payment_methods.0.payment_method_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "payment_methods.0.payment_method_type"),
				),
			},
		},
	})
}
