package schemautil

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// Those fields are not returned by API,
// but should be in the state to make terraform work correctly
var sensitiveFields = []string{
	"admin_username",
	"admin_password",
}

// copySensitiveFields copies sensitive fields to the state which not returned by API,
// but exist in the manifest
func copySensitiveFields(oldSrc interface{}, new []map[string]interface{}) {
	old := userconfig.UnmarshalUserConfig(oldSrc)
	if len(old)*len(new) == 0 {
		return
	}

	for _, k := range sensitiveFields {
		if v, ok := old[k]; ok {
			new[0][k] = v
		}
	}
}
