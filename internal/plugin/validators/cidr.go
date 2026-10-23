// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package validators

import (
	"errors"
	"net"
	"net/netip"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// CIDR creates a string validator that accepts valid CIDR notation.
func CIDR() validator.String {
	return NewStringValidator("must be a valid CIDR Value", func(v string) error {
		_, _, err := net.ParseCIDR(v)
		return err
	})
}

// IPv4CIDR creates a string validator that accepts IPv4 CIDR notation only.
// IPv4-mapped IPv6 prefixes such as "::ffff:10.0.0.0/120" are rejected.
func IPv4CIDR() validator.String {
	return NewStringValidator("must be a valid IPv4 CIDR", func(v string) error {
		p, err := netip.ParsePrefix(v)
		if err != nil {
			return err
		}
		if !p.Addr().Is4() {
			return errors.New("not an IPv4 CIDR")
		}
		return nil
	})
}
