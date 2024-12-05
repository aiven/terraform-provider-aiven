package kafka

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenKafkaNativeACLSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"resource_name": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 256),
		Description:  userconfig.Desc("The kafka resource name").ForceNew().MaxLen(256).Build(),
	},
	"resource_type": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(kafka.ResourceTypeChoices(), false),
		Description:  userconfig.Desc("The kafka resource type").ForceNew().PossibleValuesString(kafka.ResourceTypeChoices()...).Build(),
	},
	"pattern_type": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(kafka.PatternTypeChoices(), false),
		Description:  userconfig.Desc("Resource pattern used to match specified resources").ForceNew().PossibleValuesString(kafka.PatternTypeChoices()...).Build(),
	},
	"principal": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 256),
		Description:  userconfig.Desc("Principal is in type:name' format").ForceNew().MaxLen(256).Build(),
	},
	"host": {
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 256),
		Description:  userconfig.Desc("The host or `*` for all hosts").ForceNew().MaxLen(256).Build(),
	},
	"operation": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(kafka.OperationTypeChoices(), false),
		Description:  userconfig.Desc("The operation").ForceNew().PossibleValuesString(kafka.OperationTypeChoices()...).Build(),
	},
	"permission_type": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(kafka.KafkaAclPermissionTypeChoices(), false),
		Description:  userconfig.Desc("The permission type").ForceNew().PossibleValuesString(kafka.KafkaAclPermissionTypeChoices()...).Build(),
	},
}

func ResourceKafkaNativeACL() *schema.Resource {
	return &schema.Resource{
		Description:   userconfig.Desc(`Manages native acls in [kafka service](https://aiven.io/docs/products/kafka/concepts/acl)`).Build(),
		CreateContext: common.WithGenClient(resourceKafkaNativeACLCreate),
		ReadContext:   common.WithGenClient(resourceKafkaNativeACLRead),
		DeleteContext: common.WithGenClient(resourceKafkaNativeACLDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   aivenKafkaNativeACLSchema,
	}
}

func resourceKafkaNativeACLCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var req kafka.ServiceKafkaNativeAclAddIn
	err := schemautil.ResourceDataGet(d, &req)
	if err != nil {
		return err
	}

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	acl, err := client.ServiceKafkaNativeAclAdd(
		ctx,
		project,
		serviceName,
		&req,
	)
	if err != nil {
		return err
	}

	err = schemautil.ResourceDataSet(aivenKafkaNativeACLSchema, d, acl)
	if err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, acl.Id))
	return resourceKafkaNativeACLRead(ctx, d, client)
}

func resourceKafkaNativeACLRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	project, serviceName, aclID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	acl, err := client.ServiceKafkaNativeAclGet(
		ctx,
		project,
		serviceName,
		aclID,
	)
	if err != nil {
		return err
	}

	err = schemautil.ResourceDataSet(aivenKafkaNativeACLSchema, d, acl)
	return err
}

func resourceKafkaNativeACLDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	project, serviceName, aclID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	return client.ServiceKafkaNativeAclDelete(
		ctx,
		project,
		serviceName,
		aclID,
	)
}
