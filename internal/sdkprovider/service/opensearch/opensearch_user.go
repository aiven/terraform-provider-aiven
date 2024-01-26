// Package opensearch implements the Aiven OpenSearch service.
package opensearch

import (
	"context"
	"errors"
	"strings"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// errOpenSearchConfiguredDirectly is an error that we expect to be returned by the Aiven client when the OpenSearch
// Security Plugin is enabled for the OpenSearch service, and the user is trying to manage the OpenSearch users via
// the Aiven API.
const errOpenSearchConfiguredDirectly = "access to service is configured directly by opensearch security"

var aivenOpenSearchUserSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: schemautil.GetServiceUserValidateFunc(),
		Description:  userconfig.Desc("The actual name of the OpenSearch User.").ForceNew().Referenced().Build(),
	},
	"password": {
		Type:             schema.TypeString,
		Optional:         true,
		Sensitive:        true,
		Computed:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Description:      "The password of the OpenSearch User.",
	},

	// computed fields
	"type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Type of the user account. Tells whether the user is the primary account or a regular account.",
	},
}

func ResourceOpenSearchUser() *schema.Resource {
	return &schema.Resource{
		Description:   "The OpenSearch User resource allows the creation and management of Aiven OpenSearch Users.",
		CreateContext: resourceOpenSearchUserCreate,
		ReadContext:   resourceOpenSearchUserRead,
		UpdateContext: resourceOpenSearchUserUpdate,
		DeleteContext: resourceOpenSearchUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenOpenSearchUserSchema,
	}
}

// detourSecurityPluginEnabledCheck checks if the OpenSearch Security Plugin is enabled for the OpenSearch service.
// If it is enabled, it returns an error, and the resource is not allowed to be created, read or updated.
func detourSecurityPluginEnabledCheck(ctx context.Context, d *schema.ResourceData, m any) error {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	r, err := client.OpenSearchSecurityPluginHandler.Get(ctx, project, serviceName)
	if err == nil && r.SecurityPluginAdminEnabled {
		return errors.New("when the OpenSearch Security Plugin is enabled, OpenSearch users are being " +
			"managed by it; delete the aiven_opensearch_user resource(s), and manage the users via the " +
			"OpenSearch Security Plugin instead; any changes to the aiven_opensearch_user resource(s) are not " +
			"going to have any effect now")
	}

	return err
}

// resourceOpenSearchUserCreate creates a OpenSearch User.
func resourceOpenSearchUserCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if err := detourSecurityPluginEnabledCheck(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return schemautil.ResourceServiceUserCreate(ctx, d, m)
}

// resourceOpenSearchUserRead reads a OpenSearch User into the Terraform state.
func resourceOpenSearchUserRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	diags := schemautil.ResourceServiceUserRead(ctx, d, m)

	if diags == nil {
		return nil
	}

	var e *diag.Diagnostic

	for _, v := range diags {
		if v.Severity == diag.Error {
			if e != nil {
				panic("multiple errors in diags; this should never happen")
			}

			e = &v
		}
	}

	if err := detourSecurityPluginEnabledCheck(ctx, d, m); err != nil &&
		strings.Contains(strings.ToLower(e.Summary), errOpenSearchConfiguredDirectly) {
		return schemautil.ErrorToDiagWarning(err)
	}

	return diags
}

// resourceOpenSearchUserUpdate updates a OpenSearch User.
func resourceOpenSearchUserUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if err := detourSecurityPluginEnabledCheck(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return schemautil.ResourceServiceUserUpdate(ctx, d, m)
}

// resourceOpenSearchUserDelete deletes a OpenSearch User.
func resourceOpenSearchUserDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServiceUsers.Delete(ctx, projectName, serviceName, username)
	if common.IsCritical(err) {
		var e *aiven.Error

		// This is a special case where the user is not managed by Aiven, but by the OpenSearch Security plugin.
		// We don't want to fail on destroy operations if the OpenSearch Security Plugin is enabled,
		// because the users of our provider wouldn't want to have obsolete resources in their manifests, so we
		// nullify the error instead of returning it, and the resource is allowed to be destroyed, while
		// performing a no-op.
		if errors.As(err, &e) && strings.Contains(strings.ToLower(e.Message), errOpenSearchConfiguredDirectly) {
			return nil
		}

		return diag.FromErr(err)
	}

	return nil
}
