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
		Description: "MySQL specific server provided values",
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "MySQL master connection URIs",
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
					Description: "MySQL connection parameters",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"host": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "MySQL host IP or name",
							},
							"port": {
								Type:        schema.TypeInt,
								Computed:    true,
								Sensitive:   true,
								Description: "MySQL port",
							},
							"sslmode": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "MySQL sslmode setting (currently always \"require\")",
							},
							"user": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "MySQL admin user name",
							},
							"password": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "MySQL admin user password",
							},
							"database_name": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "Primary MySQL database name",
							},
						},
					},
				},
				"replica_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "MySQL replica URI for services with a replica",
					Sensitive:   true,
				},
				"standby_uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "MySQL standby connection URIs",
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
					Description: "MySQL syncing connection URIs",
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
		Description:   "The MySQL resource allows the creation and management of Aiven MySQL services.",
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
