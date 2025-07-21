package pg

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

const defaultLC = "en_US.UTF-8"

// handleLcDefaults checks if the lc values have actually changed
func handleLcDefaults(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	// NOTE! not all database resources return lc_* values even if
	// they are set when the database is created; best we can do is
	// to assume it was created using the default value.
	return newValue == "" || (oldValue == "" && newValue == defaultLC) || oldValue == newValue
}

var aivenPGDatabaseSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"database_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The name of the service database.").ForceNew().Build(),
	},
	"lc_collate": {
		Type:             schema.TypeString,
		Optional:         true,
		Default:          defaultLC,
		ForceNew:         true,
		DiffSuppressFunc: handleLcDefaults,
		Description:      userconfig.Desc("Default string sort order (`LC_COLLATE`) of the database.").DefaultValue(defaultLC).ForceNew().Build(),
	},
	"lc_ctype": {
		Type:             schema.TypeString,
		Optional:         true,
		Default:          defaultLC,
		ForceNew:         true,
		DiffSuppressFunc: handleLcDefaults,
		Description:      userconfig.Desc("Default character classification (`LC_CTYPE`) of the database.").DefaultValue(defaultLC).ForceNew().Build(),
	},
	"termination_protection": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: userconfig.Desc(`Terraform client-side deletion protection, which prevents the database from being deleted by Terraform. It's recommended to enable this for any production databases containing critical data.`).DefaultValue(false).Build(),
	},
}

func ResourcePGDatabase() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages a database in an Aiven for PostgreSQL® service.",
		CreateContext: common.WithGenClient(resourcePGDatabaseCreate),
		ReadContext:   common.WithGenClient(resourcePGDatabaseRead),
		DeleteContext: common.WithGenClient(resourcePGDatabaseDelete),
		UpdateContext: common.WithGenClient(resourcePGDatabaseUpdate),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenPGDatabaseSchema,
	}
}

func resourcePGDatabaseCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	databaseName := d.Get("database_name").(string)

	// Proves the database does not exist before creating it.
	err := schemautil.CheckDbConflict(ctx, client, projectName, serviceName, databaseName)
	if err != nil {
		return err
	}

	err = client.ServiceDatabaseCreate(
		ctx,
		projectName,
		serviceName,
		&service.ServiceDatabaseCreateIn{
			Database:  databaseName,
			LcCollate: util.NilIfZero(d.Get("lc_collate").(string)),
			LcCtype:   util.NilIfZero(d.Get("lc_ctype").(string)),
		},
	)

	switch {
	case avngen.IsAlreadyExists(err):
		// We have already checked for the existence of the database.
		// Getting a conflict here means the client retried the request.
	case err != nil:
		return err
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, databaseName))
	return resourcePGDatabaseRead(ctx, d, client)
}

// resourcePGDatabaseUpdate update is not really possible, except for the virtual field "termination_protection"
func resourcePGDatabaseUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	return resourcePGDatabaseRead(ctx, d, client)
}

func resourcePGDatabaseRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, databaseName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	database, err := getDatabaseByName(ctx, client, projectName, serviceName, databaseName)
	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	if err := d.Set("database_name", database.DatabaseName); err != nil {
		return err
	}
	if err := d.Set("project", projectName); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("lc_collate", database.LcCollate); err != nil {
		return err
	}
	if err := d.Set("lc_ctype", database.LcCtype); err != nil {
		return err
	}

	return nil
}

func resourcePGDatabaseDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, databaseName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	if d.Get("termination_protection").(bool) {
		return fmt.Errorf("cannot delete a database termination_protection is enabled")
	}

	err = client.ServiceDatabaseDelete(ctx, projectName, serviceName, databaseName)
	if err != nil {
		return err
	}

	err = schemautil.WaitUntilNotFound(ctx, func() error {
		_, err := getDatabaseByName(ctx, client, projectName, serviceName, databaseName)
		return err
	})
	if err != nil {
		return err
	}

	schemautil.ForgetDatabase(projectName, serviceName, databaseName)
	return nil
}

func getDatabaseByName(ctx context.Context, client avngen.Client, project, serviceName, dbName string) (*service.DatabaseOut, error) {
	err := schemautil.CheckServiceIsPowered(ctx, client, project, serviceName)
	if err != nil {
		return nil, err
	}

	list, err := client.ServiceDatabaseList(ctx, project, serviceName)
	if err != nil {
		return nil, err
	}
	for _, db := range list {
		if db.DatabaseName == dbName {
			return &db, nil
		}
	}
	return nil, schemautil.NewNotFound("database %q not found", dbName)
}
