package alloydbomni

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/pg"
)

func DatasourceAlloyDBOmniUser() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about an Aiven for AlloyDB Omni service user.",
		ReadContext: common.WithGenClient(pg.DatasourcePGUserRead),
		Schema:      pg.DatasourcePGUserSchema(),
	}
}
