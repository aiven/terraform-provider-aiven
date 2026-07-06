// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package validators

import (
	"net"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// CIDR creates a string validator that accepts valid CIDR notation.
func CIDR() validator.String {
	return NewStringValidator("must be a valid CIDR Value", func(v string) error {
		_, _, err := net.ParseCIDR(v)
		return err
	})
}
