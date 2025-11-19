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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

	password := d.Get("password").(string)
	if password != "" {
		_, err := client.ServiceUserCredentialsModify(ctx, projectName, serviceName, username,
			&service.ServiceUserCredentialsModifyIn{
				NewPassword: &password,
				Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
			})
		if err != nil {
			return diag.FromErr(fmt.Errorf("cannot update service user password: %w", err))
		}
	}

	// Retry because the user may not be immediately available
	// todo: Retry NotFound user might be not available immediately
	d.SetId(BuildResourceID(projectName, serviceName, username))
	return ResourceServiceUserRead(ctx, d, client)
}

func ResourceServiceUserUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName, serviceName, username, err := SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("password") {
		password := d.Get("password").(string)
		_, err = client.ServiceUserCredentialsModify(ctx, projectName, serviceName, username,
			&service.ServiceUserCredentialsModifyIn{
				NewPassword: &password,
				Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
			})
		if err != nil {
			return diag.FromErr(fmt.Errorf("cannot update service user password: %w", err))
		}
	}

	return ResourceServiceUserRead(ctx, d, client)
}

func ResourceServiceUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName, serviceName, username, err := SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
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
			if user.Password == "" {
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

	err = CopyServiceUserPropertiesFromAPIResponseToTerraform(d, user, projectName, serviceName)
	if err != nil {
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
	return OmitNotFound(err)
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
