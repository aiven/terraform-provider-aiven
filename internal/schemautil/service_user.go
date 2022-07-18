package schemautil

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/aiven/aiven-go-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceServiceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	_, err := client.ServiceUsers.Create(
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
		_, err := client.ServiceUsers.Update(projectName, serviceName, username,
			aiven.ModifyServiceUserRequest{
				NewPassword: OptionalStringPointer(d, "password"),
			})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(BuildResourceID(projectName, serviceName, username))

	return ResourceServiceUserRead(ctx, d, m)
}

func ResourceServiceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username, err := SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.ServiceUsers.Update(projectName, serviceName, username,
		aiven.ModifyServiceUserRequest{
			NewPassword: OptionalStringPointer(d, "password"),
		})
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceServiceUserRead(ctx, d, m)
}

func ResourceServiceUserRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username, err := SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	user, err := client.ServiceUsers.Get(projectName, serviceName, username)
	if err != nil {
		return diag.FromErr(ResourceReadHandleNotFound(err, d))
	}

	err = CopyServiceUserPropertiesFromAPIResponseToTerraform(d, user, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ResourceServiceUserDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username, err := SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServiceUsers.Delete(projectName, serviceName, username)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func DatasourceServiceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)

	list, err := client.ServiceUsers.List(projectName, serviceName)
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
