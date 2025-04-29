package governance

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	governance "github.com/aiven/go-client-codegen/handler/organizationgovernance"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenGovernanceKafkaAccessSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"username": {
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 54),
		Description:  userconfig.Desc("The name for the new service user given access. If not provided, the name is automatically generated.").ForceNew().MaxLen(54).Build(),
	},
	"acls": {
		Type:        schema.TypeSet,
		Required:    true,
		ForceNew:    true,
		MaxItems:    10,
		Description: userconfig.Desc("The permissions granted to the assigned service user.").ForceNew().MaxLen(54).Build(),
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: userconfig.Desc("The ACL ID.").Build(),
			},
			"resource_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
				Description:  userconfig.Desc("The name of the resource the permission applies to, such as the topic name or group ID in the Kafka service.").ForceNew().MaxLen(256).Build(),
			},
			"resource_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(governance.ResourceTypeChoices(), false),
				Description:  userconfig.Desc("The type of resource.").ForceNew().PossibleValuesString(governance.ResourceTypeChoices()...).Build(),
			},
			"pattern_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: userconfig.Desc("Pattern used to match specified resources.").PossibleValuesString(governance.PatternTypeChoices()...).Build(),
			},
			"principal": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: userconfig.Desc("Identities in `user:name` format that the permissions apply to.").Build(),
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
				Description:  userconfig.Desc("Explicitly allows or denies the action for the service user on the specified resource.").ForceNew().PossibleValuesString(governance.PermissionTypeChoices()...).Build(),
			},
			"host": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
				Description:  userconfig.Desc("The IP address from which a principal is allowed or denied access to the resource. Use `*` for all hosts.").ForceNew().MaxLen(256).Build(),
			},
		}},
	},
}

var aivenGovernanceAccessSchema = map[string]*schema.Schema{
	"organization_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The ID of the organization.").ForceNew().Build(),
	},
	"susbcription_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: userconfig.Desc("The ID of the access.").Build(),
	},
	"access_name": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 54),
		Description:  userconfig.Desc("The name to describe the access.").ForceNew().MaxLen(54).Build(),
	},
	"access_type": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(governance.AccessTypeChoices(), false),
		Description:  userconfig.Desc("The type of access.").ForceNew().PossibleValuesString(governance.AccessTypeChoices()...).Build(),
	},
	"access_data": {
		Type:             schema.TypeList,
		Description:      userconfig.Desc("Details of the access.").ForceNew().Build(),
		Required:         true,
		ForceNew:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: aivenGovernanceKafkaAccessSchema,
		},
		MaxItems: 1,
	},
	"owner_user_group_id": {
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 54),
		Description:  userconfig.Desc("The ID of the user group that owns the access.").ForceNew().MaxLen(54).Build(),
	},
}

func ResourceGovernanceAccess() *schema.Resource {
	return &schema.Resource{
		Description:   userconfig.Desc(`Request access to an Apache Kafka topic in Aiven for Apache Kafka® Governance. [Governance](https://aiven.io/docs/products/kafka/howto/governance) helps you manage your Kafka clusters securely and efficiently through structured policies, roles, and processes. You can [manage approval workflows using Terraform and GitHub Actions](https://aiven.io/docs/products/kafka/howto/terraform-governance-approvals).`).Build(),
		CreateContext: common.WithGenClient(resourceGovernanceAccessCreate),
		ReadContext:   common.WithGenClient(resourceGovernanceAccessRead),
		DeleteContext: common.WithGenClient(resourceGovernanceAccessDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   aivenGovernanceAccessSchema,
	}
}

func resourceGovernanceAccessCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var req governance.OrganizationGovernanceAccessCreateIn

	req.AccessName = d.Get("access_name").(string)
	req.AccessType = governance.AccessType(d.Get("access_type").(string))

	req.AccessData = governance.AccessDataIn{
		ProjectName: d.Get("access_data.0.project").(string),
		ServiceName: d.Get("access_data.0.service_name").(string),
		Username:    d.Get("access_data.0.username").(string),
	}

	if ownerUserGroupID := d.Get("owner_user_group_id").(string); ownerUserGroupID != "" {
		req.OwnerUserGroupId = &ownerUserGroupID
	}

	for _, v := range d.Get("access_data.0.acls").(*schema.Set).List() {
		acl := v.(map[string]interface{})
		req.AccessData.Acls = append(req.AccessData.Acls, governance.AclIn{
			ResourceName:   acl["resource_name"].(string),
			ResourceType:   governance.ResourceType(acl["resource_type"].(string)),
			Operation:      governance.OperationType(acl["operation"].(string)),
			PermissionType: governance.PermissionType(acl["permission_type"].(string)),
			Host:           acl["host"].(string),
		})
	}

	access, err := client.OrganizationGovernanceAccessCreate(
		ctx,
		d.Get("organization_id").(string),
		&req,
	)
	if err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(d.Get("organization_id").(string), access.AccessId))

	return resourceGovernanceAccessRead(ctx, d, client)
}

func resourceGovernanceAccessRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	organizationID, accessID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}
	access, err := client.OrganizationGovernanceAccessGet(
		ctx,
		organizationID,
		accessID,
	)
	if err != nil {
		return err
	}

	if err := d.Set("access_name", access.AccessName); err != nil {
		return err
	}

	if err := d.Set("access_type", access.AccessType); err != nil {
		return err
	}

	if err := d.Set("owner_user_group_id", access.OwnerUserGroupId); err != nil {
		return err
	}

	accessData := make([]map[string]any, 1)
	accessData[0] = make(map[string]any, 1)
	accessData[0]["project"] = access.AccessData.ProjectName
	accessData[0]["service_name"] = access.AccessData.ServiceName
	accessData[0]["username"] = access.AccessData.Username

	acls := make([]map[string]string, len(access.AccessData.Acls))
	for i, acl := range access.AccessData.Acls {
		acls[i] = make(map[string]string, 1)
		acls[i]["id"] = acl.Id
		acls[i]["resource_name"] = acl.ResourceName
		acls[i]["resource_type"] = string(acl.ResourceType)
		acls[i]["pattern_type"] = string(acl.PatternType)
		acls[i]["operation"] = string(acl.Operation)
		acls[i]["permission_type"] = string(acl.PermissionType)
		acls[i]["host"] = acl.Host
	}
	accessData[0]["acls"] = acls

	return d.Set("access_data", accessData)
}

func resourceGovernanceAccessDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	organizationID, accessID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}
	_, err = client.OrganizationGovernanceAccessDelete(
		ctx,
		organizationID,
		accessID,
	)
	return err
}
