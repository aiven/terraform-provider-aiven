package alloydbomni

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

const defaultLC = "en_US.UTF-8"

var aivenAlloyDBOmniDatabaseSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"database_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The name of the service database.").ForceNew().Build(),
	},
	"lc_ctype": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Default:     defaultLC,
		Description: userconfig.Desc("Default character classification (`LC_CTYPE`) of the database.").DefaultValue(defaultLC).ForceNew().Build(),
	},
	"lc_collate": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Default:     defaultLC,
		Description: userconfig.Desc("Default string sort order (`LC_COLLATE`) of the database.").DefaultValue(defaultLC).ForceNew().Build(),
	},
}

func ResourceAlloyDBOmniDatabase() *schema.Resource {
	return &schema.Resource{
		Description:        "Creates and manages a database in an Aiven for AlloyDB Omni service.",
		DeprecationMessage: deprecationMessage,
		CreateContext:      common.WithGenClient(resourceAlloyDBOmniDatabaseCreate),
		ReadContext:        common.WithGenClient(resourceAlloyDBOmniDatabaseRead),
		DeleteContext:      common.WithGenClient(resourceAlloyDBOmniDatabaseDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   aivenAlloyDBOmniDatabaseSchema,
	}
}

func resourceAlloyDBOmniDatabaseCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	req := new(service.ServiceDatabaseCreateIn)
	err := schemautil.ResourceDataGet(d, req, schemautil.RenameAlias("database_name", "database"))
	if err != nil {
		return err
	}

	err = client.ServiceDatabaseCreate(ctx, projectName, serviceName, req)
	if err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, req.Database))
	return resourceAlloyDBOmniDatabaseRead(ctx, d, client)
}

func resourceAlloyDBOmniDatabaseRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, dbName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	db, err := getDatabase(ctx, client, projectName, serviceName, dbName)
	if err != nil {
		return err
	}

	if err := d.Set("project", projectName); err != nil {
		return err
	}

	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}

	return schemautil.ResourceDataSet(
		d, db, aivenAlloyDBOmniDatabaseSchema,
		schemautil.SetForceNew("project", projectName),
		schemautil.SetForceNew("service_name", serviceName),
	)
}

func resourceAlloyDBOmniDatabaseDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, dbName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	err = client.ServiceDatabaseDelete(ctx, projectName, serviceName, dbName)
	if err != nil {
		return err
	}

	// Waits until database is deleted
	return schemautil.WaitUntilNotFound(ctx, func() error {
		_, err = getDatabase(ctx, client, projectName, serviceName, dbName)
		return err
	})
}

func getDatabase(ctx context.Context, client avngen.Client, projectName, serviceName, dbName string) (*service.DatabaseOut, error) {
	list, err := client.ServiceDatabaseList(ctx, projectName, serviceName)
	if err != nil {
		return nil, err
	}

	for _, db := range list {
		if db.DatabaseName == dbName {
			return &db, nil
		}
	}

	return nil, schemautil.NewNotFound("service database %q not found", dbName)
}
