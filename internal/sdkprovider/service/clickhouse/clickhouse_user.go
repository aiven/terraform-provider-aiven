package clickhouse

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/clickhouse"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

func aivenClickhouseUserSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"project":      schemautil.CommonSchemaProjectReference,
		"service_name": schemautil.CommonSchemaServiceNameReference,
		"username": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: schemautil.GetServiceUserValidateFunc(),
			Description:  userconfig.Desc("The name of the ClickHouse user.").ForceNew().Build(),
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

	return schemautil.MergeSchemas(s, schemautil.ServiceUserPasswordSchema())
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
		CustomizeDiff: schemautil.CustomizeDiffServiceUserPasswordWoVersion,

		Schema: aivenClickhouseUserSchema(),
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

	// handle password in priority order: password_wo > password > auto-generated
	// use GetRawConfig for password_wo because it's WriteOnly and not stored in state
	if !d.GetRawConfig().GetAttr("password_wo").IsNull() {
		passWo := d.GetRawConfig().GetAttr("password_wo").AsString() // use write-only password
		if err = resetPassword(ctx, d, client, projectName, serviceName, u.Uuid, &passWo); err != nil {
			return err
		}
	} else if password, ok := d.GetOk("password"); ok {
		pwd := password.(string)
		if err = resetPassword(ctx, d, client, projectName, serviceName, u.Uuid, &pwd); err != nil {
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
	if !d.HasChange("password_wo_version") && !d.HasChange("password") {
		return resourceClickhouseUserRead(ctx, d, client)
	}

	projectName, serviceName, uuid, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	// handle write-only password rotation
	if d.HasChange("password_wo_version") {
		if d.GetRawConfig().GetAttr("password_wo_version").IsNull() { // transition to generated password
			if err = resetPassword(ctx, d, client, projectName, serviceName, uuid, nil); err != nil {
				return err
			}
		} else {
			// rotate write-only password. Must use GetRawConfig because password_wo is WriteOnly and not stored in state
			passWo := d.GetRawConfig().GetAttr("password_wo").AsString()
			if err = resetPassword(ctx, d, client, projectName, serviceName, uuid, &passWo); err != nil {
				return err
			}
		}
	} else if d.HasChange("password") {
		password := d.Get("password").(string) // optional password changes
		if password != "" {
			if err = resetPassword(ctx, d, client, projectName, serviceName, uuid, &password); err != nil {
				return err
			}
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
// If customPassword is nil, generates a new auto-generated password and stores it in state.
// If customPassword is provided, sets it and determines storage based on whether it's write-only.
func resetPassword(
	ctx context.Context,
	d *schema.ResourceData,
	client avngen.Client,
	projectName, serviceName, uuid string,
	customPassword *string,
) error {
	usingWriteOnlyPassword := false
	if version, ok := d.GetOk("password_wo_version"); ok && version.(int) > 0 {
		usingWriteOnlyPassword = true
	}

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

	if customPassword == nil { // auto-generated password: store in state
		if err = d.Set("password", newPassword); err != nil {
			return err
		}
		if err = d.Set("password_wo_version", nil); err != nil {
			return err
		}
	} else if usingWriteOnlyPassword { // write-only password: clear the password field in state
		if err = d.Set("password", ""); err != nil {
			return err
		}
	}

	return nil
}
