package alloydbomni

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/pg"
)

func ResourceAlloyDBOmniUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an Aiven for AlloyDB Omni service user.",
		CreateContext: common.WithGenClient(pg.ResourcePGUserCreate),
		UpdateContext: common.WithGenClient(pg.ResourcePGUserUpdate),
		ReadContext:   common.WithGenClient(pg.ResourcePGUserRead),
		DeleteContext: common.WithGenClient(schemautil.ResourceServiceUserDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   pg.ResourcePGUserSchema,
	}
}
