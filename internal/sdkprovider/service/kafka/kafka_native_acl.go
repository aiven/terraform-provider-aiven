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
		Description:  userconfig.Desc("The name of the Kafka resource the permission applies to, such as the topic name or group ID.").ForceNew().MaxLen(256).Build(),
	},
	"resource_type": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(kafka.ResourceTypeChoices(), false),
		Description:  userconfig.Desc("The type of Kafka resource.").ForceNew().PossibleValuesString(kafka.ResourceTypeChoices()...).Build(),
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
		Description:  userconfig.Desc("Identities in `user:name` format that the permissions apply to. The `name` supports wildcards.").ForceNew().MaxLen(256).Build(),
	},
	"host": {
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 256),
		Description:  userconfig.Desc("The IP address from which a principal is allowed or denied access to the resource. Use `*` for all hosts.").ForceNew().MaxLen(256).Build(),
	},
	"operation": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(kafka.OperationTypeChoices(), false),
		Description:  userconfig.Desc("The action that a principal is allowed or denied on the Kafka resource.").ForceNew().PossibleValuesString(kafka.OperationTypeChoices()...).Build(),
	},
	"permission_type": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(kafka.KafkaAclPermissionTypeChoices(), false),
		Description:  userconfig.Desc("Specifies whether the action is explicitly allowed or denied for the principal on the specified resource.").ForceNew().PossibleValuesString(kafka.KafkaAclPermissionTypeChoices()...).Build(),
	},
}

func ResourceKafkaNativeACL() *schema.Resource {
	return &schema.Resource{
		Description: userconfig.Desc(`Creates and manages Kafka-native [access control lists](https://aiven.io/docs/products/kafka/concepts/acl) (ACLs) for an Aiven for Apache Kafka® service. ACLs control access to Kafka topics, consumer groups,
clusters, and Schema Registry.

Kafka-native ACLs provide advanced resource-level access control with fine-grained permissions, including ` + "`ALLOW`" + ` and ` + "`DENY`" + ` rules. For simplified topic-level control you can use [Aiven ACLs](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/kafka_acl).`).Build(),
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

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	acl, err := client.ServiceKafkaNativeAclAdd(
		ctx,
		projectName,
		serviceName,
		&req,
	)
	if err != nil {
		return err
	}

	err = schemautil.ResourceDataSet(
		d, acl, aivenKafkaNativeACLSchema,
		schemautil.AddForceNew("project", projectName),
		schemautil.AddForceNew("service_name", serviceName),
	)
	if err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, acl.Id))
	return resourceKafkaNativeACLRead(ctx, d, client)
}

func resourceKafkaNativeACLRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, aclID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	acl, err := client.ServiceKafkaNativeAclGet(
		ctx,
		projectName,
		serviceName,
		aclID,
	)
	if err != nil {
		return err
	}

	err = schemautil.ResourceDataSet(
		d, acl, aivenKafkaNativeACLSchema,
		schemautil.AddForceNew("project", projectName),
		schemautil.AddForceNew("service_name", serviceName),
	)
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
