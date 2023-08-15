package vpc_test

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
