package organization

import (
	"context"
	"fmt"
	"log"
	"sync"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/account"
	"github.com/aiven/go-client-codegen/handler/organization"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// permissionLock locks Upsert operation to run conflict validation.
var permissionLock sync.Mutex

// envPermissionValidateConflict by default is true.
const (
	envPermissionValidateConflict = "AIVEN_ORGANIZATION_PERMISSION_VALIDATE_CONFLICT"
	permissionRegistryDocsURL     = "https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/organization_permission"
)

var aivenOrganizationalPermissionSchema = map[string]*schema.Schema{
	"organization_id": {
		Type:        schema.TypeString,
		Description: "Organization ID.",
		Required:    true,
	},
	"resource_type": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice(organization.ResourceTypeChoices(), false),
		Description:  userconfig.Desc("Resource type.").PossibleValuesString(organization.ResourceTypeChoices()...).Build(),
	},
	"resource_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Resource ID.",
	},
	"permissions": {
		Type:        schema.TypeSet,
		Description: "Permissions to grant to principals.",
		Required:    true,
		Elem: &schema.Resource{
			Schema: permissionFields,
		},
	},
}

var permissionFields = map[string]*schema.Schema{
	"principal_type": {
		Type:        schema.TypeString,
		Required:    true,
		Description: userconfig.Desc("The type of principal.").PossibleValuesString(organization.PrincipalTypeChoices()...).Build(),
	},
	"principal_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "ID of the user or group to grant permissions to. Only active users who have accepted an [invite](https://aiven.io/docs/platform/howto/manage-org-users) to join the organization can be granted permissions.",
	},
	"permissions": {
		Type:        schema.TypeSet,
		Description: userconfig.Desc("List of [roles and permissions](https://aiven.io/docs/platform/concepts/permissions) to grant.").PossibleValuesString(account.MemberTypeChoices()...).Build(),
		Required:    true,
		Elem:        &schema.Schema{Type: schema.TypeString},
	},
	"create_time": {
		Type:        schema.TypeString,
		Description: "Time created.",
		Computed:    true,
	},
	"update_time": {
		Type:        schema.TypeString,
		Description: "Time updated.",
		Computed:    true,
	},
}

func ResourceOrganizationalPermission() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf(`Grants [roles and permissions](https://aiven.io/docs/platform/concepts/permissions)
to a principal for a resource. Permissions can be granted at the organization, organizational unit, and project level.
Unit-level permissions aren't shown in the Aiven Console.

To assign permissions to multiple users and groups on the same combination of organization ID, resource ID and resource type, don't use multiple `+"`aiven_organization_permission`"+` resources.
Instead, use multiple permission blocks as in the example usage.

**Do not use the `+"`aiven_project_user`"+` or `+"`aiven_organization_group_project`"+` resources with this resource**.

By default, Aiven Terraform Provider validates whether the resource already exists in the Aiven API. 
This validation prevents you from managing permissions for a specific resource using multiple `+"`aiven_organization_group_project`"+` resources, 
which leads to overwrites and conflicts. 
In case of a conflict, you can import the resource using the `+"`terraform import`"+` command to continue managing it.
Alternatively, you can disable this validation by setting the `+"`%s`"+` environment variable to `+"`false`"+`, 
which will cause Terraform to override the remote state.
`, envPermissionValidateConflict),
		CreateContext: common.WithGenClient(resourceOrganizationalPermissionUpsert),
		ReadContext:   common.WithGenClient(resourceOrganizationalPermissionRead),
		UpdateContext: common.WithGenClient(resourceOrganizationalPermissionUpsert),
		DeleteContext: common.WithGenClient(resourceOrganizationalPermissionDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   aivenOrganizationalPermissionSchema,
	}
}

func resourceOrganizationalPermissionUpsert(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID := d.Get("organization_id").(string)
	resourceType := d.Get("resource_type").(string)
	resourceID := d.Get("resource_id").(string)
	composeID := schemautil.BuildResourceID(orgID, resourceType, resourceID)

	// Validates new resources only.
	// Update operation will definitely have a remote state.
	if d.Id() == "" {
		permissionLock.Lock()
		defer permissionLock.Unlock()

		v, err := client.PermissionsGet(ctx, orgID, organization.ResourceType(resourceType), resourceID)
		if err != nil {
			return fmt.Errorf("failed to read remote state: %w", err)
		}

		switch {
		case len(v) == 0:
			// The remote state is empty.
			// Can proceed with the upsert.
		case util.EnvBool(envPermissionValidateConflict, true):
			// The remote state is not empty and the validation is enabled.
			return fmt.Errorf(
				"resource %q already has permissions set. "+
					"Probably there is another `aiven_organization_permission` managing it. "+
					"Please follow the [instructions](%s)",
				composeID,
				permissionRegistryDocsURL,
			)
		default:
			log.Printf(
				"[WARNING] Conflict validation is disabled. "+
					"The remote state is not empty and will be overridden. "+
					"This will cause issues if %q is managed by another resource.",
				composeID,
			)
		}
	}

	req := new(organization.PermissionsSetIn)
	err := schemautil.ResourceDataGet(d, req)
	if err != nil {
		return err
	}

	if req.Permissions == nil {
		req.Permissions = make([]organization.PermissionIn, 0)
	}

	err = client.PermissionsSet(ctx, orgID, organization.ResourceType(resourceType), resourceID, req)
	if err != nil {
		return err
	}

	d.SetId(composeID)
	return resourceOrganizationalPermissionRead(ctx, d, client)
}

func resourceOrganizationalPermissionRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID, resourceType, resourceID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	out, err := client.PermissionsGet(ctx, orgID, organization.ResourceType(resourceType), resourceID)
	if err != nil {
		return err
	}

	permissions := make([]map[string]any, 0, len(out))
	err = schemautil.Remarshal(out, &permissions)
	if err != nil {
		return err
	}

	// Removes fields that are not on the schema,
	// so it won't blow up when the DTO gets new fields with the updates
	for _, m := range permissions {
		for k := range m {
			if _, ok := permissionFields[k]; !ok {
				delete(m, k)
			}
		}
	}

	return d.Set("permissions", permissions)
}

func resourceOrganizationalPermissionDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID, resourceType, resourceID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	req := &organization.PermissionsSetIn{
		Permissions: make([]organization.PermissionIn, 0),
	}

	// Omits 404 in case there is an issue with the remote state.
	err = client.PermissionsSet(ctx, orgID, organization.ResourceType(resourceType), resourceID, req)
	return schemautil.OmitNotFound(err)
}
