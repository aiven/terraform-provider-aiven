package mysql

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
)

func aivenMySQLSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeMySQL)
	s[schemautil.ServiceTypeMySQL] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "MySQL server-provided values.",
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "MySQL connection URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
				"params": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "MySQL connection parameters.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"host": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "MySQL host IP or name.",
							},
							"port": {
								Type:        schema.TypeInt,
								Computed:    true,
								Sensitive:   true,
								Description: "MySQL port.",
							},
							"sslmode": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "MySQL SSL mode setting. Always set to \"require\".",
							},
							"user": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "The username for the admin service user.",
							},
							"password": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "The password for the admin service user.",
							},
							"database_name": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "Thr name of the primary MySQL database.",
							},
						},
					},
				},
				"replica_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "MySQL replica URI for services with a replica.",
					Sensitive:   true,
				},
				"standby_uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "MySQL standby connection URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
				"syncing_uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "MySQL syncing connection URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
			},
		},
	}
	return s
}
func ResourceMySQL() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [Aiven for MySQL®](https://aiven.io/docs/products/mysql) service.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeMySQL),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: schemautil.CustomizeDiffGenericService(schemautil.ServiceTypeMySQL),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts:       schemautil.DefaultResourceTimeouts(),
		Schema:         aivenMySQLSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.MySQL(),
	}
}
