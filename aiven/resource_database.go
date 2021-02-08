// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"

	"github.com/aiven/aiven-go-client"
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

var aivenDatabaseSchema = map[string]*schema.Schema{
	"project": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Project to link the database to",
		ForceNew:    true,
	},
	"service_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Service to link the database to",
		ForceNew:    true,
	},
	"database_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Service database name",
		ForceNew:    true,
	},
	"lc_collate": {
		Type:             schema.TypeString,
		Optional:         true,
		Default:          defaultLC,
		Description:      "Default string sort order (LC_COLLATE) of the database. Default value: en_US.UTF-8",
		ForceNew:         true,
		DiffSuppressFunc: handleLcDefaults,
	},
	"lc_ctype": {
		Type:             schema.TypeString,
		Optional:         true,
		Default:          defaultLC,
		Description:      "Default character classification (LC_CTYPE) of the database. Default value: en_US.UTF-8",
		ForceNew:         true,
		DiffSuppressFunc: handleLcDefaults,
	},
	"termination_protection": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
		Description: `It is a Terraform client-side deletion protections, which prevents the database
			from being deleted by Terraform. It is recommended to enable this for any production
			databases containing critical data.`,
	},
}

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseCreate,
		ReadContext:   resourceDatabaseRead,
		DeleteContext: resourceDatabaseDelete,
		UpdateContext: resourceDatabaseUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDatabaseState,
		},

		// TODO: add user config
		Schema: aivenDatabaseSchema,
	}
}

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	databaseName := d.Get("database_name").(string)
	_, err := client.Databases.Create(
		projectName,
		serviceName,
		aiven.CreateDatabaseRequest{
			Database:  databaseName,
			LcCollate: optionalString(d, "lc_collate"),
			LcType:    optionalString(d, "lc_ctype"),
		},
	)
	if err != nil && !aiven.IsAlreadyExists(err) {
		return diag.FromErr(err)
	}

	d.SetId(buildResourceID(projectName, serviceName, databaseName))

	return resourceDatabaseRead(ctx, d, m)
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceDatabaseRead(ctx, d, m)
}

func resourceDatabaseRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, databaseName := splitResourceID3(d.Id())
	database, err := client.Databases.Get(projectName, serviceName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", database.DatabaseName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project", projectName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("lc_collate", database.LcCollate); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("lc_ctype", database.LcType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("termination_protection", d.Get("termination_protection")); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, databaseName := splitResourceID3(d.Id())

	if d.Get("termination_protection").(bool) {
		return diag.Errorf("cannot delete a database termination_protection is enabled")
	}

	err := client.Databases.Delete(projectName, serviceName, databaseName)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDatabaseState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<database_name>", d.Id())
	}

	di := resourceDatabaseRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get database: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}
