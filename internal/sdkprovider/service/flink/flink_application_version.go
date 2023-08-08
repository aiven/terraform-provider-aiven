package flink

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// aivenFlinkApplicationVersionSchema is the schema of the Flink Application Version resource.
var aivenFlinkApplicationVersionSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"application_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "Application ID",
	},
	"sinks": {
		Type:             schema.TypeSet,
		Optional:         true,
		ForceNew:         true,
		Description:      "Application sinks",
		Deprecated:       "This field is deprecated and will be removed in the next major release. Use `sink` instead.",
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		ConflictsWith:    []string{"sink"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"create_table": {
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
					Description: "The CREATE TABLE statement",
				},
				"integration_id": {
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
					Description: "The integration ID",
				},
			},
		},
	},
	"sink": {
		Type:             schema.TypeSet,
		Optional:         true,
		ForceNew:         true,
		Description:      "Application sink",
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		ConflictsWith:    []string{"sinks"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"create_table": {
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
					Description: "The CREATE TABLE statement",
				},
				"integration_id": {
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
					Description: "The integration ID",
				},
			},
		},
	},
	"sources": {
		Type:             schema.TypeSet,
		Optional:         true,
		ForceNew:         true,
		Description:      "Application sources",
		Deprecated:       "This field is deprecated and will be removed in the next major release. Use `source` instead.",
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		ConflictsWith:    []string{"source"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"create_table": {
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
					Description: "The CREATE TABLE statement",
				},
				"integration_id": {
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
					Description: "The integration ID",
				},
			},
		},
	},
	"source": {
		Type:             schema.TypeSet,
		Optional:         true,
		ForceNew:         true,
		Description:      "Application source",
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		ConflictsWith:    []string{"sources"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"create_table": {
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
					Description: "The CREATE TABLE statement",
				},
				"integration_id": {
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
					Description: "The integration ID",
				},
			},
		},
	},
	"statement": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "Job SQL statement",
	},

	// Computed fields.
	"application_version_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Application version ID",
	},
	"created_at": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Application version creation time",
	},
	"created_by": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Application version creator",
	},
	"version": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Application version number",
	},
}

// ResourceFlinkApplicationVersion returns the Flink Application Version resource schema.
func ResourceFlinkApplicationVersion() *schema.Resource {
	return &schema.Resource{
		Description:   "The Flink Application Version resource allows the creation and management of Aiven Flink Application Versions.",
		CreateContext: resourceFlinkApplicationVersionCreate,
		ReadContext:   resourceFlinkApplicationVersionRead,
		DeleteContext: resourceFlinkApplicationVersionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   aivenFlinkApplicationVersionSchema,
	}
}

// resourceFlinkApplicationVersionCreate is the create function for the Flink Application Version resource.
func resourceFlinkApplicationVersionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	applicationID := d.Get("application_id").(string)

	sources := []aiven.FlinkApplicationVersionRelation{}
	sinks := []aiven.FlinkApplicationVersionRelation{}
	if d.Get("sources").(*schema.Set).Len() > 0 {
		sources = expandFlinkApplicationVersionSourcesOrSinks(d.Get("sources").(*schema.Set).List())
	}
	if d.Get("sinks").(*schema.Set).Len() > 0 {
		sinks = expandFlinkApplicationVersionSourcesOrSinks(d.Get("sinks").(*schema.Set).List())
	}
	if d.Get("source").(*schema.Set).Len() > 0 {
		sources = expandFlinkApplicationVersionSourcesOrSinks(d.Get("source").(*schema.Set).List())
	}
	if d.Get("sink").(*schema.Set).Len() > 0 {
		sinks = expandFlinkApplicationVersionSourcesOrSinks(d.Get("sink").(*schema.Set).List())
	}

	r, err := client.FlinkApplicationVersions.Create(project, serviceName, applicationID, aiven.GenericFlinkApplicationVersionRequest{
		Statement: d.Get("statement").(string),
		Sources:   sources,
		Sinks:     sinks,
	})
	if err != nil {
		return diag.Errorf("cannot create Flink Application Version: %+v - %v", expandFlinkApplicationVersionSourcesOrSinks(d.Get("sources").(*schema.Set).List()), err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, applicationID, r.ID))

	return resourceFlinkApplicationVersionRead(ctx, d, m)
}

// expandFlinkApplicationVersionSourcesOrSinks expands the sources or sinks from the Terraform schema to the Aiven API.
func expandFlinkApplicationVersionSourcesOrSinks(sources []interface{}) []aiven.FlinkApplicationVersionRelation {
	var result []aiven.FlinkApplicationVersionRelation
	for _, source := range sources {
		sourceMap := source.(map[string]interface{})
		result = append(result, aiven.FlinkApplicationVersionRelation{
			CreateTable:   sourceMap["create_table"].(string),
			IntegrationID: sourceMap["integration_id"].(string),
		})
	}

	return result
}

// resourceFlinkApplicationVersionDelete is the delete function for the Flink Application Version resource.
func resourceFlinkApplicationVersionDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, applicationID, version, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return diag.Errorf("cannot read Flink Application Version resource ID: %v", err)
	}

	_, err = client.FlinkApplicationVersions.Delete(project, serviceName, applicationID, version)
	if err != nil {
		return diag.Errorf("error deleting Flink Application Version: %v", err)
	}

	return nil
}

// resourceFlinkApplicationVersionRead is the read function for the Flink Application Version resource.
func resourceFlinkApplicationVersionRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, applicationID, version, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return diag.Errorf("cannot read Flink Application Version resource ID: %v", err)
	}

	r, err := client.FlinkApplicationVersions.Get(project, serviceName, applicationID, version)
	if err != nil {
		return diag.Errorf("cannot get Flink Application Version: %v", err)
	}

	if err := d.Set("project", project); err != nil {
		return diag.Errorf("error setting Flink Application Version `project` field: %s", err)
	}

	if err := d.Set("service_name", serviceName); err != nil {
		return diag.Errorf("error setting Flink Application Version `service_name` field: %s", err)
	}

	if err := d.Set("application_id", applicationID); err != nil {
		return diag.Errorf("error setting Flink Application Version `application_id` field: %s", err)
	}

	if err := d.Set("statement", r.Statement); err != nil {
		return diag.Errorf("error setting Flink Application Version `statement` field: %s", err)
	}

	if err := d.Set("sources", flattenFlinkApplicationVersionSourcesOrSinks(r.Sources)); err != nil {
		return diag.Errorf("error setting Flink Application Version `sources` field: %s", err)
	}

	if err := d.Set("sinks", flattenFlinkApplicationVersionSourcesOrSinks(r.Sinks)); err != nil {
		return diag.Errorf("error setting Flink Application Version `sinks` field: %s", err)
	}

	if err := d.Set("source", flattenFlinkApplicationVersionSourcesOrSinks(r.Sources)); err != nil {
		return diag.Errorf("error setting Flink Application Version `source` field: %s", err)
	}

	if err := d.Set("sink", flattenFlinkApplicationVersionSourcesOrSinks(r.Sinks)); err != nil {
		return diag.Errorf("error setting Flink Application Version `sink` field: %s", err)
	}

	if err := d.Set("application_version_id", r.ID); err != nil {
		return diag.Errorf("error setting Flink Application Version `application_version_id` field: %s", err)
	}
	if err := d.Set("version", r.Version); err != nil {
		return diag.Errorf("error setting Flink Application Version `version` field: %s", err)
	}
	if err := d.Set("created_at", r.CreatedAt); err != nil {
		return diag.Errorf("error setting Flink Application Version `created_at` field: %s", err)
	}
	if err := d.Set("created_by", r.CreatedBy); err != nil {
		return diag.Errorf("error setting Flink Application Version `created_by` field: %s", err)
	}

	return nil
}

// flattenFlinkApplicationVersionSourcesOrSinks is a helper function to flatten the sources and sinks fields.
func flattenFlinkApplicationVersionSourcesOrSinks(sources []aiven.FlinkApplicationVersionRelation) []map[string]interface{} {
	var result []map[string]interface{}
	for _, source := range sources {
		result = append(result, map[string]interface{}{
			"create_table":   source.CreateTable,
			"integration_id": source.IntegrationID,
		})
	}

	return result
}
