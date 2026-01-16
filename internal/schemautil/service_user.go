package schemautil

import (
	"context"
	"fmt"
	"log"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/aiven/terraform-provider-aiven/internal/common"
)

func ResourceServiceUserCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)

	_, err := client.ServiceUserCreate(
		ctx,
		projectName,
		serviceName,
		&service.ServiceUserCreateIn{
			Username: username,
		},
	)
	if err != nil {
		return diag.FromErr(fmt.Errorf("cannot create service user: %w", err))
	}

	if err = UpsertPassword(ctx, d, client); err != nil {
		return diag.FromErr(err)
	}

	// Retry because the user may not be immediately available
	// todo: Retry NotFound user might be not available immediately
	d.SetId(BuildResourceID(projectName, serviceName, username))

	return ResourceServiceUserRead(ctx, d, client)
}

func ResourceServiceUserUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	if err := UpsertPassword(ctx, d, client); err != nil {
		return diag.FromErr(err)
	}

	return ResourceServiceUserRead(ctx, d, client)
}

func ResourceServiceUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName, serviceName, username, err := SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// check if using write-only password
	usingWriteOnlyPassword := false
	if version, ok := d.GetOk("password_wo_version"); ok {
		usingWriteOnlyPassword = version.(int) > 0
	}

	// User password might be Null https://api.aiven.io/doc/#tag/Service/operation/ServiceUserGet
	// > Account password. A null value indicates a user overridden password.
	var user *service.ServiceUserGetOut
	err = retry.Do(
		func() error {
			user, err = client.ServiceUserGet(ctx, projectName, serviceName, username)
			if err != nil {
				return retry.Unrecoverable(err)
			}
			// The field is not nullable, so we compare to an empty string
			// Only wait for password if not using write-only password
			if !usingWriteOnlyPassword && user.Password == "" {
				return fmt.Errorf("password is not received from the API")
			}
			return nil
		},
		retry.Context(ctx),
		retry.Delay(time.Second),
		retry.LastErrorOnly(true), // retry returns a list of errors by default
	)
	if err != nil {
		return diag.FromErr(ResourceReadHandleNotFound(err, d))
	}

	if err = CopyServiceUserPropertiesFromAPIResponseToTerraform(d, user, projectName, serviceName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ResourceServiceUserDelete(ctx context.Context, d ResourceData, client avngen.Client) error {
	projectName, serviceName, username, err := SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	err = client.ServiceUserDelete(ctx, projectName, serviceName, username)
	return common.OmitNotFound(err)
}

func DatasourceServiceUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)

	svc, err := client.ServiceGet(ctx, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, u := range svc.Users {
		if u.Username == userName {
			d.SetId(BuildResourceID(projectName, serviceName, userName))
			return ResourceServiceUserRead(ctx, d, client)
		}
	}

	return diag.Errorf("user %s/%s/%s not found",
		projectName, serviceName, userName)
}

func TestAccCheckAivenServiceUserAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] user service attributes %v", a)

		if a["username"] == "" {
			return fmt.Errorf("expected to get a Service User username from Aiven")
		}

		if a["password"] == "" {
			return fmt.Errorf("expected to get a Service User password from Aiven")
		}

		if a["project"] == "" {
			return fmt.Errorf("expected to get a Service User project from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a Service User service_name from Aiven")
		}

		return nil
	}
}

// ServiceUserPasswordSchema returns the standard password schema fields for service users,
// including support for write-only passwords.
func ServiceUserPasswordSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"password": {
			Type:             schema.TypeString,
			Optional:         true,
			Sensitive:        true,
			Computed:         true,
			DiffSuppressFunc: EmptyObjectDiffSuppressFunc,
			ConflictsWith:    []string{"password_wo"},
			ValidateFunc:     validation.StringLenBetween(8, 256),
			Description:      "The password of the service user (auto-generated if not provided). Must be 8-256 characters if specified.",
		},
		"password_wo": {
			Type:          schema.TypeString,
			Optional:      true,
			Sensitive:     true,
			WriteOnly:     true,
			RequiredWith:  []string{"password_wo_version"},
			ConflictsWith: []string{"password"},
			ValidateFunc:  validation.StringLenBetween(8, 256),
			Description:   "The password of the service user (write-only, not stored in state). Must be used with `password_wo_version`. Must be 8-256 characters.",
		},
		"password_wo_version": {
			Type:         schema.TypeInt,
			Optional:     true,
			RequiredWith: []string{"password_wo"},
			ValidateFunc: validation.IntAtLeast(1),
			Description:  "Version number for `password_wo`. Increment this to rotate the password. Must be >= 1.",
		},
	}
}

// CustomizeDiffServiceUserPasswordWoVersion ensures that password_wo_version only increases.
// Allows removal of write-only password by removing password_wo_version.
// This enforces the policy that write-only passwords can only be rotated forward and follow the same UX as other providers.
func CustomizeDiffServiceUserPasswordWoVersion(ctx context.Context, diff *schema.ResourceDiff, m any) error {
	if err := customizeDiffPasswordWoVersion("password_wo_version")(ctx, diff, m); err != nil {
		return err
	}

	return CustomizeDiffWriteOnlyPasswordTransitionWarning("password", "password_wo_version")(ctx, diff, m)
}

// ClearPasswordIfWriteOnly clears the password field if using write-only mode.
// Must be called after setting other fields from API response, because the password may be copied
// from the API response into state.
func ClearPasswordIfWriteOnly(d ResourceData) error {
	if version, ok := d.GetOk("password_wo_version"); ok && version.(int) > 0 {
		if err := d.Set("password", nil); err != nil {
			return err
		}
	}
	return nil
}

// UpsertPassword handles password setting and rotation for service users.
func UpsertPassword(ctx context.Context, d ResourceData, client avngen.Client) error {
	password, shouldReset := shouldResetPassword(d)
	if !shouldReset {
		return nil
	}

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)

	if password == "" { // auto-generate password
		_, err := client.ServiceUserCredentialsReset(ctx, projectName, serviceName, username)
		if err != nil {
			return fmt.Errorf("cannot reset service user password: %w", err)
		}

		return nil
	}

	// set custom password
	_, err := client.ServiceUserCredentialsModify(ctx, projectName, serviceName, username,
		&service.ServiceUserCredentialsModifyIn{
			NewPassword: &password,
			Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
		})
	if err != nil {
		return fmt.Errorf("cannot set service user password: %w", err)
	}

	return nil
}

// shouldResetPassword determines if a password reset is needed and returns the password to use if so.
func shouldResetPassword(d ResourceData) (string, bool) {
	var password string

	// write-only password takes precedence
	passwordWoAttr := d.GetRawConfig().GetAttr("password_wo")
	if !passwordWoAttr.IsNull() {
		password = passwordWoAttr.AsString()
	} else if pwd, ok := d.GetOk("password"); ok {
		password = pwd.(string)
	}

	// reset needed on create or when any password field changes
	shouldReset := d.Id() == "" || d.HasChange("password") || d.HasChange("password_wo_version")

	return password, shouldReset
}
