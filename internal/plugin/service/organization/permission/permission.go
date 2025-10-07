package permission

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/aiven/go-client-codegen/handler/accountteam"
	"github.com/aiven/go-client-codegen/handler/organization"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// permissionLock locks Upsert operation to run conflict validation.
var permissionLock sync.Mutex

// envPermissionValidateConflict by default is true.
const (
	envPermissionValidateConflict = "AIVEN_ORGANIZATION_PERMISSION_VALIDATE_CONFLICT"
	permissionRegistryDocsURL     = "https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/organization_permission"
)

func NewResource() resource.Resource {
	return adapter.NewResource(aivenName, new(view), patchedSchema, newResourceModel, composeID())
}

// patchedSchema adds "permissions" enum values that are not yet in OpenAPI spec.
// They are exactly the same as for teams.
func patchedSchema(ctx context.Context) schema.Schema {
	s := newResourceSchema(ctx)
	b := s.Blocks["permissions"].(schema.SetNestedBlock)
	v := b.NestedObject.Attributes["permissions"].(schema.SetAttribute)
	v.MarkdownDescription = userconfig.
		Desc(v.MarkdownDescription).
		PossibleValuesString(accountteam.TeamTypeChoices()...).Build()
	b.NestedObject.Attributes["permissions"] = v
	return s
}

type view struct{ adapter.View }

func (vw *view) Create(ctx context.Context, plan *tfModel) diag.Diagnostics {
	permissionLock.Lock()
	defer permissionLock.Unlock()
	diags := vw.validateConflict(ctx, plan)
	if diags.HasError() {
		return diags
	}

	return vw.Update(ctx, plan, nil, nil)
}

func (vw *view) Update(ctx context.Context, plan, state, _ *tfModel) diag.Diagnostics {
	var req organization.PermissionsSetIn
	diags := expandData(ctx, plan, state, &req)
	if diags.HasError() {
		return diags
	}

	orgID := plan.OrganizationID.ValueString()
	resourceType := plan.ResourceType.ValueString()
	resourceID := plan.ResourceID.ValueString()
	err := vw.Client.PermissionsSet(ctx, orgID, organization.ResourceType(resourceType), resourceID, &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return diags
	}

	// Sets ID fields to Read() the resource
	plan.SetID(orgID, resourceType, resourceID)
	return vw.Read(ctx, plan)
}

func (vw *view) Read(ctx context.Context, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := vw.Client.PermissionsGet(
		ctx,
		state.OrganizationID.ValueString(),
		organization.ResourceType(state.ResourceType.ValueString()),
		state.ResourceID.ValueString(),
	)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	type wrapper struct {
		Permissions []organization.PermissionOut `json:"permissions"`
	}
	return flattenData(ctx, state, &wrapper{Permissions: rsp})
}

func (vw *view) Delete(ctx context.Context, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	req := &organization.PermissionsSetIn{
		Permissions: make([]organization.PermissionIn, 0),
	}

	orgID := state.OrganizationID.ValueString()
	resourceType := state.ResourceType.ValueString()
	resourceID := state.ResourceID.ValueString()
	err := vw.Client.PermissionsSet(ctx, orgID, organization.ResourceType(resourceType), resourceID, req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}
	return nil
}

func (vw *view) validateConflict(ctx context.Context, plan *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	orgID := plan.OrganizationID.ValueString()
	resourceType := plan.ResourceType.ValueString()
	resourceID := plan.ResourceID.ValueString()
	v, err := vw.Client.PermissionsGet(ctx, orgID, organization.ResourceType(resourceType), resourceID)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, fmt.Sprintf("failed to read remote state: %s", err))
		return diags
	}

	fullID := fmt.Sprintf("%s/%s/%s", orgID, resourceType, resourceID)

	switch {
	case len(v) == 0:
		// The remote state is empty.
		// Can proceed with the upsert.
	case util.EnvBool(envPermissionValidateConflict, true):
		// The remote state is not empty and the validation is enabled.
		msg := fmt.Sprintf(
			"resource conflict: The target %q already has permissions configured. "+
				"This likely indicates another `aiven_organization_permission` resource is managing these permissions "+
				"Please follow the [instructions](%s)",
			fullID,
			permissionRegistryDocsURL,
		)
		diags.AddError(errmsg.SummaryErrorCreatingResource, msg)
		return diags
	default:
		log.Printf(
			"[WARNING] Conflict validation is disabled. "+
				"The remote state is not empty and will be overridden. "+
				"This will cause issues if %q is managed by another resource.",
			fullID,
		)
	}
	return nil
}
