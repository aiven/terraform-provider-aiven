package governance

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/governance"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenGovernanceKafkaSubscriptionSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"username": {
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 54),
		Description:  userconfig.Desc("The name that will be used for the new service user assigned to the subscription. If not provided, name is automatically generated").ForceNew().MaxLen(54).Build(),
	},
	"acls": {
		Type:        schema.TypeSet,
		Required:    true,
		ForceNew:    true,
		MaxItems:    10,
		Description: userconfig.Desc("The permissions granted to the assigned service user").ForceNew().MaxLen(54).Build(),
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: userconfig.Desc("The acl id").Build(),
			},
			"resource_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
				Description:  userconfig.Desc("The name of the resource the permission applies to, such as the topic name or group ID in Kafka service").ForceNew().MaxLen(256).Build(),
			},
			"resource_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(governance.ResourceTypeChoices(), false),
				Description:  userconfig.Desc("The type of resource on the service.").ForceNew().PossibleValuesString(governance.ResourceTypeChoices()...).Build(),
			},
			"pattern_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: userconfig.Desc("Resource pattern used to match specified resources").PossibleValuesString(governance.PatternTypeChoices()...).Build(),
			},
			"principal": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: userconfig.Desc("Identities in `user:name` format that the permissions apply to").Build(),
			},
			"operation": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(governance.OperationTypeChoices(), false),
				Description:  userconfig.Desc("The action that will be allowed for the service user.").ForceNew().PossibleValuesString(governance.OperationTypeChoices()...).Build(),
			},
			"permission_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(governance.PermissionTypeChoices(), false),
				Description:  userconfig.Desc("Specifies whether the action is explicitly allowed or denied for the service user on the specified resource").ForceNew().PossibleValuesString(governance.PermissionTypeChoices()...).Build(),
			},
			"host": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
				Description:  userconfig.Desc("The IP address from which a principal is allowed or denied access to the resource. Use `*` for all hosts").ForceNew().MaxLen(256).Build(),
			},
		}},
	},
}

var aivenGovernanceSubscriptionSchema = map[string]*schema.Schema{
	"organization_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The ID of the organization").ForceNew().Build(),
	},
	"susbcription_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: userconfig.Desc("The ID of the subscription").Build(),
	},
	"subscription_name": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 54),
		Description:  userconfig.Desc("The name to describe the subscription").ForceNew().MaxLen(54).Build(),
	},
	"subscription_type": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(governance.SubscriptionTypeChoices(), false),
		Description:  userconfig.Desc("The type of subscription. For example KAFKA.").ForceNew().PossibleValuesString(governance.SubscriptionTypeChoices()...).Build(),
	},
	"subscription_data": {
		Type:             schema.TypeList,
		Description:      userconfig.Desc("The data defined by subscription_type.").ForceNew().Build(),
		Required:         true,
		ForceNew:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: aivenGovernanceKafkaSubscriptionSchema,
		},
		MaxItems: 1,
	},
	"owner_user_group_id": {
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 54),
		Description:  userconfig.Desc("The ID of the user group that owns the subscription").ForceNew().MaxLen(54).Build(),
	},
}

func ResourceGovernanceSubscription() *schema.Resource {
	return &schema.Resource{
		Description:   userconfig.Desc(`Creates and manages governance subscriptions for an organization. Governance subscriptions provide convenient management and access to service resources and is part of Aiven Kafka Governance`).Build(),
		CreateContext: common.WithGenClient(resourceGovernanceSubscriptionCreate),
		ReadContext:   common.WithGenClient(resourceGovernanceSubscriptionRead),
		DeleteContext: common.WithGenClient(resourceGovernanceSubscriptionDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   aivenGovernanceSubscriptionSchema,
	}
}

func resourceGovernanceSubscriptionCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var req governance.OrganizationGovernanceSubscriptionCreateIn

	req.SubscriptionName = d.Get("subscription_name").(string)
	req.SubscriptionType = governance.SubscriptionType(d.Get("subscription_type").(string))

	req.SubscriptionData = governance.SubscriptionDataIn{
		ProjectName: d.Get("subscription_data.0.project").(string),
		ServiceName: d.Get("subscription_data.0.service_name").(string),
		Username:    d.Get("Username").(string),
	}

	if ownerUserGroupID := d.Get("owner_user_group_id").(string); ownerUserGroupID != "" {
		req.OwnerUserGroupId = &ownerUserGroupID
	}

	for _, v := range d.Get("subscription_data.0.acls").(*schema.Set).List() {
		acl := v.(map[string]interface{})
		req.SubscriptionData.Acls = append(req.SubscriptionData.Acls, governance.AclIn{
			ResourceName:   acl["resource_name"].(string),
			ResourceType:   governance.ResourceType(acl["resource_type"].(string)),
			Operation:      governance.OperationType(acl["operation"].(string)),
			PermissionType: governance.PermissionType(acl["permission_type"].(string)),
			Host:           acl["host"].(string),
		})
	}

	subscription, err := client.OrganizationGovernanceSubscriptionCreate(
		ctx,
		d.Get("organization_id").(string),
		&req,
	)
	if err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(d.Get("organization_id").(string), subscription.SubscriptionId))

	return resourceGovernanceSubscriptionRead(ctx, d, client)
}

func resourceGovernanceSubscriptionRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	organizationID, subscriptionID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}
	subscription, err := client.OrganizationGovernanceSubscriptionGet(
		ctx,
		organizationID,
		subscriptionID,
	)
	if err != nil {
		return err
	}

	if err := d.Set("subscription_name", subscription.SubscriptionName); err != nil {
		return err
	}

	if err := d.Set("subscription_type", subscription.SubscriptionType); err != nil {
		return err
	}

	if err := d.Set("owner_user_group_id", subscription.OwnerUserGroupId); err != nil {
		return err
	}

	subscriptionData := make([]map[string]any, 1)
	subscriptionData[0] = make(map[string]any, 1)
	subscriptionData[0]["project"] = subscription.SubscriptionData.ProjectName
	subscriptionData[0]["service_name"] = subscription.SubscriptionData.ServiceName
	subscriptionData[0]["username"] = subscription.SubscriptionData.Username

	acls := make([]map[string]string, len(subscription.SubscriptionData.Acls))
	for i, acl := range subscription.SubscriptionData.Acls {
		acls[i] = make(map[string]string, 1)
		acls[i]["id"] = acl.Id
		acls[i]["resource_name"] = acl.ResourceName
		acls[i]["resource_type"] = string(acl.ResourceType)
		acls[i]["pattern_type"] = string(acl.PatternType)
		acls[i]["operation"] = string(acl.Operation)
		acls[i]["permission_type"] = string(acl.PermissionType)
		acls[i]["host"] = acl.Host
	}
	subscriptionData[0]["acls"] = acls

	return d.Set("subscription_data", subscriptionData)
}

func resourceGovernanceSubscriptionDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	_, err := client.OrganizationGovernanceSubscriptionDelete(
		ctx,
		d.Get("organization_id").(string),
		d.Id(),
	)
	return err
}
