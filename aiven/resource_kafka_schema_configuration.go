package aiven

import (
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

var compatibilityLevels = []string{
	"BACKWARD",
	"BACKWARD_TRANSITIVE",
	"FORWARD",
	"FORWARD_TRANSITIVE",
	"FULL",
	"FULL_TRANSITIVE",
	"NONE"}

var aivenKafkaSchemaConfigurationSchema = map[string]*schema.Schema{
	"project": {
		Type:        schema.TypeString,
		Description: "Project to link the Kafka Schemas Configuration to",
		Required:    true,
		ForceNew:    true,
	},
	"service_name": {
		Type:        schema.TypeString,
		Description: "Service to link the Kafka Schemas Configuration to",
		Required:    true,
		ForceNew:    true,
	},
	"compatibility_level": {
		Type:         schema.TypeString,
		Description:  "Kafka Schemas compatibility level",
		Required:     true,
		ValidateFunc: validation.StringInSlice(compatibilityLevels, false),
	},
}

func resourceKafkaSchemaConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceKafkaSchemaConfigurationCreate,
		Update: resourceKafkaSchemaConfigurationUpdate,
		Read:   resourceKafkaSchemaConfigurationRead,
		Delete: resourceKafkaSchemaConfigurationDelete,
		Exists: resourceKafkaSchemaConfigurationExists,
		Importer: &schema.ResourceImporter{
			State: resourceKafkaSchemaConfigurationState,
		},

		Schema: aivenKafkaSchemaConfigurationSchema,
	}
}

func resourceKafkaSchemaConfigurationUpdate(d *schema.ResourceData, m interface{}) error {
	project, serviceName := splitResourceID2(d.Id())

	_, err := m.(*aiven.Client).KafkaGlobalSchemaConfig.Update(
		project,
		serviceName,
		aiven.KafkaSchemaConfig{
			CompatibilityLevel: d.Get("compatibility_level").(string),
		})
	if err != nil {
		return err
	}

	return resourceKafkaSchemaConfigurationRead(d, m)
}

// resourceKafkaSchemaConfigurationCreate Kafka Schemas global configuration cannot be created but only updated
func resourceKafkaSchemaConfigurationCreate(d *schema.ResourceData, m interface{}) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	_, err := m.(*aiven.Client).KafkaGlobalSchemaConfig.Update(
		project,
		serviceName,
		aiven.KafkaSchemaConfig{
			CompatibilityLevel: d.Get("compatibility_level").(string),
		})
	if err != nil {
		return err
	}

	d.SetId(buildResourceID(project, serviceName))

	return resourceKafkaSchemaConfigurationRead(d, m)
}

func resourceKafkaSchemaConfigurationRead(d *schema.ResourceData, m interface{}) error {
	project, serviceName := splitResourceID2(d.Id())

	r, err := m.(*aiven.Client).KafkaGlobalSchemaConfig.Get(project, serviceName)
	if err != nil {
		return err
	}

	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("compatibility_level", r.CompatibilityLevel); err != nil {
		return err
	}

	return nil
}

// resourceKafkaSchemaConfigurationDelete Kafka Schemas configuration cannot be deleted, therefore
// on delete event configuration will be set to the default setting
func resourceKafkaSchemaConfigurationDelete(d *schema.ResourceData, m interface{}) error {
	project, serviceName := splitResourceID2(d.Id())

	_, err := m.(*aiven.Client).KafkaGlobalSchemaConfig.Update(
		project,
		serviceName,
		aiven.KafkaSchemaConfig{
			CompatibilityLevel: "BACKWARD",
		})

	return err
}

func resourceKafkaSchemaConfigurationExists(d *schema.ResourceData, m interface{}) (bool, error) {
	return resourceExists(resourceKafkaSchemaConfigurationRead(d, m))
}

func resourceKafkaSchemaConfigurationState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	err := resourceKafkaSchemaConfigurationRead(d, m)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
