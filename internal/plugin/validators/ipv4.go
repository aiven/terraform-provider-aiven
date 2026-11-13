// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package validators

import (
	"errors"
	"net/netip"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// IPv4 accepts dotted-quad IPv4 addresses, including public addresses.
// IPv6 and IPv4-mapped IPv6 addresses are rejected.
func IPv4() validator.String {
	return NewStringValidator("must be a valid IPv4 address", func(v string) error {
		address, err := netip.ParseAddr(v)
		if err != nil {
			return err
		}
		if !address.Is4() {
			return errors.New("not an IPv4 address")
		}
		return nil
	})
}
