package permission

import (
	"context"
	"fmt"
	"log"
	"sync"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/accountteam"
	"github.com/aiven/go-client-codegen/handler/organization"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
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
	return adapter.NewResource(adapter.ResourceOptions{
		TypeName:       typeName,
		IDFields:       idFields(),
		Schema:         patchedSchema,
		SchemaInternal: resourceSchemaInternal(),
		Read:           readPermission,
		Create:         createPermission,
		Update:         updatePermission,
		Delete:         deletePermission,
	})
}

// patchedSchema adds "permissions" enum values that are not yet in OpenAPI spec.
// They are exactly the same as for teams.
func patchedSchema(ctx context.Context) schema.Schema {
	s := resourceSchema(ctx)
	b := s.Blocks["permissions"].(schema.SetNestedBlock)
	v := b.NestedObject.Attributes["permissions"].(schema.SetAttribute)
	v.MarkdownDescription = userconfig.
		Desc(v.MarkdownDescription).
		PossibleValuesString(accountteam.TeamTypeChoices()...).Build()
	b.NestedObject.Attributes["permissions"] = v
	return s
}

func createPermission(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	permissionLock.Lock()
	defer permissionLock.Unlock()
	err := validateConflict(ctx, client, d)
	if err != nil {
		return err
	}

	return updatePermission(ctx, client, d)
}

func updatePermission(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	var req organization.PermissionsSetIn
	err := d.Expand(&req)
	if err != nil {
		return err
	}

	orgID := d.Get("organization_id").(string)
	resourceType := d.Get("resource_type").(string)
	resourceID := d.Get("resource_id").(string)
	err = client.PermissionsSet(ctx, orgID, organization.ResourceType(resourceType), resourceID, &req)
	if err != nil {
		return err
	}

	// Sets ID fields to Read() the resource
	err = d.SetID(orgID, resourceType, resourceID)
	if err != nil {
		return err
	}
	return readPermission(ctx, client, d)
}

func readPermission(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	rsp, err := client.PermissionsGet(
		ctx,
		d.Get("organization_id").(string),
		organization.ResourceType(d.Get("resource_type").(string)),
		d.Get("resource_id").(string),
	)
	if err != nil {
		return err
	}
	return d.Flatten(&map[string]any{"permissions": rsp})
}

func deletePermission(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	orgID := d.Get("organization_id").(string)
	resourceType := d.Get("resource_type").(string)
	resourceID := d.Get("resource_id").(string)
	return client.PermissionsSet(ctx, orgID, organization.ResourceType(resourceType), resourceID, &organization.PermissionsSetIn{
		Permissions: make([]organization.PermissionIn, 0),
	})
}

func validateConflict(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	orgID := d.Get("organization_id").(string)
	resourceType := d.Get("resource_type").(string)
	resourceID := d.Get("resource_id").(string)
	v, err := client.PermissionsGet(ctx, orgID, organization.ResourceType(resourceType), resourceID)
	if err != nil {
		return err
	}

	fullID := fmt.Sprintf("%s/%s/%s", orgID, resourceType, resourceID)

	switch {
	case len(v) == 0:
		// The remote state is empty.
		// Can proceed with the upsert.
	case util.EnvBool(envPermissionValidateConflict, true):
		// The remote state is not empty and the validation is enabled.
		return fmt.Errorf(
			"resource conflict: The target %q already has permissions configured. "+
				"This likely indicates another `aiven_organization_permission` resource is managing these permissions "+
				"Please follow the [instructions](%s)",
			fullID,
			permissionRegistryDocsURL,
		)
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
