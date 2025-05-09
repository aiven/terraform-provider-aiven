package kafkaschema

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/aiven/go-client-codegen/handler/kafkaschemaregistry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenKafkaSchemaRegistryACLSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"permission": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(kafkaschemaregistry.PermissionTypeChoices(), false),
		Description:  userconfig.Desc("Kafka Schema Registry permission to grant.").ForceNew().PossibleValuesString(kafkaschemaregistry.PermissionTypeChoices()...).Build(),
	},
	"resource": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("Resource name pattern for the Schema Registry ACL entry.").ForceNew().Build(),
	},
	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: schemautil.GetACLUserValidateFunc(),
		Description:  userconfig.Desc("Username pattern for the ACL entry.").ForceNew().Build(),
	},

	// computed
	"acl_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Kafka Schema Registry ACL ID",
	},
}

func ResourceKafkaSchemaRegistryACL() *schema.Resource {
	return &schema.Resource{
		Description:   "The Resource Kafka Schema Registry ACL resource allows the creation and management of Schema Registry ACLs for an Aiven Kafka service.",
		CreateContext: resourceKafkaSchemaRegistryACLCreate,
		ReadContext:   resourceKafkaSchemaRegistryACLRead,
		DeleteContext: resourceKafkaSchemaRegistryACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenKafkaSchemaRegistryACLSchema,
	}
}

func resourceKafkaSchemaRegistryACLCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	acl, err := client.KafkaSchemaRegistryACLs.Create(
		ctx,
		project,
		serviceName,
		aiven.CreateKafkaSchemaRegistryACLRequest{
			Permission: d.Get("permission").(string),
			Resource:   d.Get("resource").(string),
			Username:   d.Get("username").(string),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, acl.ID))

	return resourceKafkaSchemaRegistryACLRead(ctx, d, m)
}

func resourceKafkaSchemaRegistryACLRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, aclID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	acl, err := client.KafkaSchemaRegistryACLs.Get(ctx, project, serviceName, aclID)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	err = copyKafkaSchemaRegistryACLPropertiesFromAPIResponseToTerraform(d, acl, project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaSchemaRegistryACLDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, aclID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.KafkaSchemaRegistryACLs.Delete(ctx, projectName, serviceName, aclID)
	if common.IsCritical(err) {
		return diag.FromErr(err)
	}

	return nil
}

func copyKafkaSchemaRegistryACLPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	acl *aiven.KafkaSchemaRegistryACL,
	project string,
	serviceName string,
) error {
	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("resource", acl.Resource); err != nil {
		return err
	}
	if err := d.Set("permission", acl.Permission); err != nil {
		return err
	}
	if err := d.Set("username", acl.Username); err != nil {
		return err
	}

	return d.Set("acl_id", acl.ID)
}
