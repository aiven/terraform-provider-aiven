package database

import (
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenDatabaseMySQLSchema = schemautil.DatabaseCommonSchema

func ResourceMySQLDatabase() *schema.Resource {
	return &schema.Resource{
		Description:   "The MySQL Database resource allows the creation and management of Aiven MySQL Databases.",
		CreateContext: schemautil.ResourceDatabaseCreate,
		ReadContext:   schemautil.ResourceDatabaseRead,
		DeleteContext: schemautil.ResourceDatabaseDelete,
		UpdateContext: schemautil.ResourceDatabaseUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schemautil.ResourceDatabaseState,
		},
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: aivenDatabaseMySQLSchema,
	}
}
