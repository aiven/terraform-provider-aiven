package kafka

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenKafkaACLSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"permission": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice([]string{"admin", "read", "readwrite", "write"}, false),
		Description:  userconfig.Desc("Permissions to grant.").ForceNew().PossibleValues("admin", "read", "readwrite", "write").Build(),
	},
	"topic": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("Topics that the permissions apply to.").ForceNew().Build(),
	},
	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: schemautil.GetACLUserValidateFunc(),
		Description:  userconfig.Desc("Usernames to grant permissions to.").ForceNew().Build(),
	},

	// computed
	"acl_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Kafka ACL ID.",
	},
}

func ResourceKafkaACL() *schema.Resource {
	return &schema.Resource{
		Description: `
Creates and manages an [access control list](https://aiven.io/docs/products/kafka/concepts/acl) (ACL) entry for an Aiven for Apache Kafka® service.

ACL entries grant users rights to produce, consume, and manage Kafka topics.
`,
		CreateContext: resourceKafkaACLCreate,
		ReadContext:   resourceKafkaACLRead,
		DeleteContext: resourceKafkaACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenKafkaACLSchema,
	}
}

func resourceKafkaACLCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	acl, err := client.KafkaACLs.Create(
		ctx,
		project,
		serviceName,
		aiven.CreateKafkaACLRequest{
			Permission: d.Get("permission").(string),
			Topic:      d.Get("topic").(string),
			Username:   d.Get("username").(string),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, acl.ID))

	return resourceKafkaACLRead(ctx, d, m)
}

func resourceKafkaACLRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, aclID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	acl, err := kafkaACLCache{}.Read(ctx, project, serviceName, aclID, client)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	err = copyKafkaACLPropertiesFromAPIResponseToTerraform(d, &acl, project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaACLDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, aclID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.KafkaACLs.Delete(ctx, projectName, serviceName, aclID)
	if common.IsCritical(err) {
		return diag.FromErr(err)
	}

	return nil
}

func copyKafkaACLPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	acl *aiven.KafkaACL,
	project string,
	serviceName string,
) error {
	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("topic", acl.Topic); err != nil {
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
