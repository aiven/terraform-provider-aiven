package schemautil

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func ResourceServiceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	_, err := client.ServiceUsers.Create(
		ctx,
		projectName,
		serviceName,
		aiven.CreateServiceUserRequest{
			Username: username,
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("password"); ok {
		_, err := client.ServiceUsers.Update(ctx, projectName, serviceName, username,
			aiven.ModifyServiceUserRequest{
				NewPassword: OptionalStringPointer(d, "password"),
			})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Retry because the user may not be immediately available
	// todo: Retry NotFound user might be not available immediately
	d.SetId(BuildResourceID(projectName, serviceName, username))
	return ResourceServiceUserRead(ctx, d, m)
}

func ResourceServiceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username, err := SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("password") {
		_, err = client.ServiceUsers.Update(ctx, projectName, serviceName, username,
			aiven.ModifyServiceUserRequest{
				NewPassword: OptionalStringPointer(d, "password"),
			})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceServiceUserRead(ctx, d, m)
}

func ResourceServiceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username, err := SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// User password might be Null https://api.aiven.io/doc/#tag/Service/operation/ServiceUserGet
	// > Account password. A null value indicates a user overridden password.
	var user *aiven.ServiceUser
	err = retry.Do(
		func() error {
			user, err = client.ServiceUsers.Get(ctx, projectName, serviceName, username)
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

func DatasourceServiceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)

	list, err := client.ServiceUsers.List(ctx, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, u := range list {
		if u.Username == userName {
			d.SetId(BuildResourceID(projectName, serviceName, userName))
			return ResourceServiceUserRead(ctx, d, m)
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
