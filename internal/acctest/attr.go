package acctest

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// StoreAttr captures the current value of a resource attribute into target, for comparison
// in a later test step with CheckAttrDiffers or TestCheckResourceAttrPtr.
func StoreAttr(resourceName, key string, target *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		value, err := attrValue(s, resourceName, key)
		if err != nil {
			return err
		}
		*target = value
		return nil
	}
}

// CheckAttrDiffers fails when a resource attribute matches the value captured earlier by
// StoreAttr. Pairs with StoreAttr to assert that a step forced a replacement.
func CheckAttrDiffers(resourceName, key string, previous *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		value, err := attrValue(s, resourceName, key)
		if err != nil {
			return err
		}
		if value == *previous {
			return fmt.Errorf("expected %s.%s to change, got %q", resourceName, key, value)
		}
		return nil
	}
}

// attrValue returns a non-empty attribute from state, or an error identifying the missing one.
func attrValue(s *terraform.State, resourceName, key string) (string, error) {
	rs, err := ResourceFromState(s, resourceName)
	if err != nil {
		return "", err
	}
	value := rs.Primary.Attributes[key]
	if value == "" {
		return "", fmt.Errorf("attribute %q of %q is empty", key, resourceName)
	}
	return value, nil
}
