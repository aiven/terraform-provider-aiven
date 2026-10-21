// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package validators

import (
	"errors"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// reUUID accepts only the canonical form the API returns. google/uuid also accepts
// braced, urn and unhyphenated forms, which would diff against the API value.
var reUUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// UUID creates a string validator that accepts UUIDs in canonical 8-4-4-4-12 hex form.
func UUID() validator.String {
	return NewStringValidator("must be a valid UUID", func(v string) error {
		if !reUUID.MatchString(v) {
			return errors.New("invalid UUID")
		}
		return nil
	})
}
