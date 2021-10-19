package aiven

import (
	"context"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		Description:   "The Kafka resource allows the creation and management of Aiven Kafka services.",
		CreateContext: resourceKafkaCreate,
		ReadContext:   resourceServiceRead,
		UpdateContext: resourceServiceUpdate,
		DeleteContext: resourceServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: aivenKafkaSchema(),
	}
}

func resourceKafkaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if di := resourceServiceCreateWrapper(ServiceTypeKafka)(ctx, d, m); di.HasError() {
		return di
	}

	// if default_acl=false delete default wildcard Kafka ACL that is automatically created
	if !d.Get("default_acl").(bool) {
		client := m.(*aiven.Client)
		project := d.Get("project").(string)
		serviceName := d.Get("service_name").(string)

		list, err := client.KafkaACLs.List(project, serviceName)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return diag.Errorf("cannot get a list of kafka acl's: %s", err)
			}
		}

		for _, acl := range list {
			if acl.Username == "*" && acl.Topic == "*" && acl.Permission == "admin" {
				err := client.KafkaACLs.Delete(project, serviceName, acl.ID)
				if err != nil {
					return diag.Errorf("cannot delete default wildcard kafka acl: %s", err)
				}
			}
		}
	}

	return nil
}
