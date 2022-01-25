// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var compatibilityLevels = []string{
	"BACKWARD",
	"BACKWARD_TRANSITIVE",
	"FORWARD",
	"FORWARD_TRANSITIVE",
	"FULL",
	"FULL_TRANSITIVE",
	"NONE",
}

var aivenKafkaSchemaConfigurationSchema = map[string]*schema.Schema{
	"project":      commonSchemaProjectReference,
	"service_name": commonSchemaServiceNameReference,

	"compatibility_level": {
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(compatibilityLevels, false),
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			// When a compatibility level is not set to any value and consequently is null (empty string).
			// Allow ignoring those.
			return new == ""
		},
		Description: complex("Kafka Schemas compatibility level.").possibleValues(stringSliceToInterfaceSlice(compatibilityLevels)...).build(),
	},
}

func resourceKafkaSchemaConfiguration() *schema.Resource {
	return &schema.Resource{
		Description:   "The Kafka Schema Configuration resource allows the creation and management of Aiven Kafka Schema Configurations.",
		CreateContext: resourceKafkaSchemaConfigurationCreate,
		UpdateContext: resourceKafkaSchemaConfigurationUpdate,
		ReadContext:   resourceKafkaSchemaConfigurationRead,
		DeleteContext: resourceKafkaSchemaConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKafkaSchemaConfigurationState,
		},

		Schema: aivenKafkaSchemaConfigurationSchema,
	}
}

func resourceKafkaSchemaConfigurationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, serviceName := schemautil.SplitResourceID2(d.Id())

	_, err := m.(*aiven.Client).KafkaGlobalSchemaConfig.Update(
		project,
		serviceName,
		aiven.KafkaSchemaConfig{
			CompatibilityLevel: d.Get("compatibility_level").(string),
		})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKafkaSchemaConfigurationRead(ctx, d, m)
}

// resourceKafkaSchemaConfigurationCreate Kafka Schemas global configuration cannot be created but only updated
func resourceKafkaSchemaConfigurationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	_, err := m.(*aiven.Client).KafkaGlobalSchemaConfig.Update(
		project,
		serviceName,
		aiven.KafkaSchemaConfig{
			CompatibilityLevel: d.Get("compatibility_level").(string),
		})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceKafkaSchemaConfigurationRead(ctx, d, m)
}

func resourceKafkaSchemaConfigurationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, serviceName := schemautil.SplitResourceID2(d.Id())

	r, err := m.(*aiven.Client).KafkaGlobalSchemaConfig.Get(project, serviceName)
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
	}

	if err := d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("compatibility_level", r.CompatibilityLevel); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// resourceKafkaSchemaConfigurationDelete Kafka Schemas configuration cannot be deleted, therefore
// on delete event configuration will be set to the default setting
func resourceKafkaSchemaConfigurationDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, serviceName := schemautil.SplitResourceID2(d.Id())

	_, err := m.(*aiven.Client).KafkaGlobalSchemaConfig.Update(
		project,
		serviceName,
		aiven.KafkaSchemaConfig{
			CompatibilityLevel: "BACKWARD",
		})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaSchemaConfigurationState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	di := resourceKafkaSchemaConfigurationRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get kafka schema configuration: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}
