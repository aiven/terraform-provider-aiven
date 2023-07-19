package vpc_test

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func importStateByName(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName: name,
		ImportState:  true,
		ImportStateIdFunc: func(s *terraform.State) (string, error) {
			root := s.RootModule()
			rs, ok := root.Resources[name]
			if !ok {
				return "", fmt.Errorf(`resource %q not found in the state`, name)
			}
			return rs.Primary.ID, nil
		},
	}
}
