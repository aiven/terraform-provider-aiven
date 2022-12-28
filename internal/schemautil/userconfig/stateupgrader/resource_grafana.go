package stateupgrader

import (
	v0 "github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Grafana() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceGrafanaResourceV0().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceGrafanaStateUpgradeV0,
			Version: 0,
		},
	}
}
