package user

import (
	"context"
	"fmt"
	"net/http"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/clickhouse"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

func init() {
	ResourceOptions.Update = updateView
}

func updateView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if d.HasChange("password") || d.HasChange("password_wo_version") {
		password, err := client.ServiceClickHousePasswordReset(
			ctx,
			d.Get("project").(string),
			d.Get("service_name").(string),
			d.Get("uuid").(string),
			&clickhouse.ServiceClickHousePasswordResetIn{
				Password: util.NilIfZero(d.Get("password").(string), d.Get("password_wo").(string)),
			},
		)
		if err != nil {
			return err
		}

		return d.Flatten(&map[string]any{"password": password}, flattenModifier(ctx, client))
	}

	return nil
}

func planModifier(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if d.Get("uuid").(string) != "" {
		return nil
	}

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)

	users, err := client.ServiceClickHouseUserList(ctx, project, serviceName)
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Name == username {
			return d.Flatten(&user)
		}
	}

	return avngen.Error{
		Message:     fmt.Sprintf("clickhouse user %q not found in project %q, service %q", username, project, serviceName),
		OperationID: "ServiceClickHouseUserList",
		Status:      http.StatusNotFound,
	}
}

func expandModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		if v, ok := d.GetOk("password_wo"); ok {
			// Sets Write-only password to the password field in the API request
			dto["password"] = v
		}
		return nil
	}
}

func flattenModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		if _, ok := d.Schema().Properties["password_wo_version"]; ok && d.Get("password_wo_version").(int) != 0 {
			// Clear password from state when using write-only password.
			delete(dto, "password")
		}

		return nil
	}
}
