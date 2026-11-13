package schemautil

import (
	"context"
	"fmt"
	"log"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

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
