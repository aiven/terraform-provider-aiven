package database

import (
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const defaultLC = "en_US.UTF-8"

// handleLcDefaults checks if the lc values have actually changed
func handleLcDefaults(_, old, new string, _ *schema.ResourceData) bool {
	// NOTE! not all database resources return lc_* values even if
	// they are set when the database is created; best we can do is
	// to assume it was created using the default value.
	return new == "" || (old == "" && new == defaultLC) || old == new
}

func aivenDatabasePGSchema() map[string]*schema.Schema {
	aivenDatabasePGSchema := schemautil.DatabaseCommonSchema
	aivenDatabasePGSchema["lc_collate"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Default:          defaultLC,
		ForceNew:         true,
		DiffSuppressFunc: handleLcDefaults,
		Description:      schemautil.Complex("Default string sort order (`LC_COLLATE`) of the database.").DefaultValue(defaultLC).ForceNew().Build(),
	}

	aivenDatabasePGSchema["lc_ctype"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Default:          defaultLC,
		ForceNew:         true,
		DiffSuppressFunc: handleLcDefaults,
		Description:      schemautil.Complex("Default character classification (`LC_CTYPE`) of the database.").DefaultValue(defaultLC).ForceNew().Build(),
	}
	return aivenDatabasePGSchema
}

func ResourcePGDatabase() *schema.Resource {
	return &schema.Resource{
		Description:   "The Database resource allows the creation and management of Aiven PG Databases.",
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

		Schema: aivenDatabasePGSchema(),
	}
}
