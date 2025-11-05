package clickhouse

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/clickhouse"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenClickhouseUserSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: schemautil.GetServiceUserValidateFunc(),
		Description:  userconfig.Desc("The name of the ClickHouse user.").ForceNew().Build(),
	},
	"password": {
		Type:        schema.TypeString,
		Sensitive:   true,
		Computed:    true,
		Description: "The password of the ClickHouse user (generated). Empty when using `password_wo`.",
	},
	"password_wo": {
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		RequiredWith:  []string{"password_wo_version"},
		ConflictsWith: []string{"password"},
		ValidateFunc:  validation.StringIsNotEmpty,
		Description:   "The password of the ClickHouse user (write-only, not stored in state). Must be used with `password_wo_version`. Cannot be empty.",
	},
	"password_wo_version": {
		Type:         schema.TypeInt,
		Optional:     true,
		RequiredWith: []string{"password_wo"},
		ValidateFunc: validation.IntAtLeast(1),
		Description:  "Version number for `password_wo`. Increment this to rotate the password. Must be >= 1.",
	},
	"uuid": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "UUID of the ClickHouse user.",
	},
	"required": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Indicates if a ClickHouse user is required.",
	},
}

func ResourceClickhouseUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages a ClickHouse user.",
		CreateContext: common.WithGenClient(resourceClickhouseUserCreate),
		ReadContext:   common.WithGenClient(resourceClickhouseUserRead),
		UpdateContext: common.WithGenClient(resourceClickhouseUserUpdate),
		DeleteContext: common.WithGenClient(resourceClickhouseUserDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts:      schemautil.DefaultResourceTimeouts(),
		CustomizeDiff: customizeDiffPasswordWoVersion,

		Schema: aivenClickhouseUserSchema,
	}
}

func resourceClickhouseUserCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)

	u, err := client.ServiceClickHouseUserCreate(
		ctx,
		projectName,
		serviceName,
		&clickhouse.ServiceClickHouseUserCreateIn{
			Name: username,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot create ClickHouse user: %w", err)
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, u.Uuid))

	// check if using write-only password or auto-generated password
	passWo := d.Get("password_wo").(string)
	if passWo != "" { // use custom write-only password
		if err = resetPassword(ctx, d, client, projectName, serviceName, u.Uuid, &passWo); err != nil {
			return err
		}
	} else if u.Password != nil { // use auto-generated password
		if err = d.Set("password", *u.Password); err != nil {
			return err
		}
	}

	return resourceClickhouseUserRead(ctx, d, client)
}

func resourceClickhouseUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, uuid, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	var user *clickhouse.UserOut
	ul, err := client.ServiceClickHouseUserList(ctx, projectName, serviceName)
	if err != nil {
		return err
	}

	for _, u := range ul {
		if u.Uuid == uuid {
			user = &u
			break
		}
	}

	if user == nil {
		return schemautil.ResourceReadHandleNotFound(fmt.Errorf("user %q not found", d.Id()), d)
	}

	if err = d.Set("project", projectName); err != nil {
		return err
	}
	if err = d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err = d.Set("username", user.Name); err != nil {
		return err
	}
	if err = d.Set("uuid", user.Uuid); err != nil {
		return err
	}
	if err = d.Set("required", user.Required); err != nil {
		return err
	}

	return nil
}

func resourceClickhouseUserUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	if !d.HasChange("password_wo_version") {
		return resourceClickhouseUserRead(ctx, d, client)
	}

	projectName, serviceName, uuid, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	// check if user is removing write-only password (transition to generated password)
	if d.GetRawConfig().GetAttr("password_wo_version").IsNull() {
		if err = resetPassword(ctx, d, client, projectName, serviceName, uuid, nil); err != nil {
			return err
		}
	} else { // rotate write-only password
		customPassword := d.Get("password_wo").(string)
		if err = resetPassword(ctx, d, client, projectName, serviceName, uuid, &customPassword); err != nil {
			return err
		}
	}

	return resourceClickhouseUserRead(ctx, d, client)
}

func resourceClickhouseUserDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, uuid, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	err = client.ServiceClickHouseUserDelete(ctx, projectName, serviceName, uuid)
	if common.IsCritical(err) {
		return err
	}

	return nil
}

// resetPassword handles password reset scenarios for a ClickHouse user.
// If customPassword is provided, sets it as write-only (clears password field).
// If customPassword is nil, generates new auto-generated password (clears password_wo_version).
func resetPassword(
	ctx context.Context,
	d *schema.ResourceData,
	client avngen.Client,
	projectName, serviceName, uuid string,
	customPassword *string,
) error {
	newPassword, err := client.ServiceClickHousePasswordReset(
		ctx,
		projectName,
		serviceName,
		uuid,
		&clickhouse.ServiceClickHousePasswordResetIn{
			Password: customPassword,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot reset ClickHouse user password: %w", err)
	}

	// update state based on password
	if customPassword != nil { // write-only password
		if err = d.Set("password", nil); err != nil { // clear the computed password field and persist version
			return err
		}
	} else { // auto-generated password
		if err = d.Set("password", newPassword); err != nil {
			return err
		}

		if err = d.Set("password_wo_version", nil); err != nil {
			return err
		}
	}

	return nil
}

// customizeDiffPasswordWoVersion ensures that password_wo_version only increases.
// Allows removal of write-only password by removing password_wo_version.
// This enforces the policy that write-only passwords can only be rotated forward and follow the same UX as other providers.
func customizeDiffPasswordWoVersion(_ context.Context, diff *schema.ResourceDiff, _ any) error {
	if diff.HasChange("password_wo_version") {
		oldRaw, newRaw := diff.GetChange("password_wo_version")

		// initial setting
		if oldRaw == nil || oldRaw.(int) == 0 {
			return nil
		}

		oldVersion := oldRaw.(int)
		newVersion := newRaw.(int)

		// allow removal
		if newVersion == 0 {
			return nil
		}

		// prevent decrement
		if newVersion < oldVersion {
			return fmt.Errorf("password_wo_version must be incremented (old: %d, new: %d). Decrementing version is not allowed", oldVersion, newVersion)
		}
	}

	return nil
}
