package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"time"
)

func aivenKafkaSchema() map[string]*schema.Schema {
	aivenKafkaSchema := serviceCommonSchema()
	aivenKafkaSchema["default_acl"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "Create default wildcard Kafka ACL",
	}
	aivenKafkaSchema[ServiceTypeKafka] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Kafka server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"access_cert": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka client certificate",
					Optional:    true,
					Sensitive:   true,
				},
				"access_key": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka client certificate key",
					Optional:    true,
					Sensitive:   true,
				},
				"connect_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka Connect URI, if any",
					Optional:    true,
					Sensitive:   true,
				},
				"rest_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka REST URI, if any",
					Optional:    true,
					Sensitive:   true,
				},
				"schema_registry_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Schema Registry URI, if any",
					Optional:    true,
					Sensitive:   true,
				},
			},
		},
	}
	aivenKafkaSchema[ServiceTypeKafka+"_user_config"] = &schema.Schema{
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "Kafka user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeKafka].(map[string]interface{})),
		},
	}

	return aivenKafkaSchema
}

func resourceKafka() *schema.Resource {
	return &schema.Resource{
		Create: resourceKafkaCreate,
		Read:   resourceServiceRead,
		Update: resourceServiceUpdate,
		Delete: resourceServiceDelete,
		Exists: resourceServiceExists,
		Importer: &schema.ResourceImporter{
			State: resourceServiceState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: aivenKafkaSchema(),
	}
}

func resourceKafkaCreate(d *schema.ResourceData, m interface{}) error {
	if err := resourceServiceCreateWrapper(ServiceTypeKafka)(d, m); err != nil {
		return err
	}

	// if default_acl=false delete default wildcard Kafka ACL that is automatically created
	if d.Get("default_acl").(bool) == false {
		client := m.(*aiven.Client)
		project := d.Get("project").(string)
		serviceName := d.Get("service_name").(string)

		list, err := client.KafkaACLs.List(project, serviceName)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return fmt.Errorf("cannot get a list of kafka acl's: %w", err)
			}
		}

		for _, acl := range list {
			if acl.Username == "*" && acl.Topic == "*" && acl.Permission == "admin" {
				err := client.KafkaACLs.Delete(project, serviceName, acl.ID)
				if err != nil {
					return fmt.Errorf("cannot delete default wildcard kafka acl: %w", err)
				}
			}
		}
	}

	return nil
}
