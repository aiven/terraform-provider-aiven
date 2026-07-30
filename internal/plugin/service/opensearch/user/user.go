package user

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/serviceuser"
)

func init() {
	ResourceOptions.Update = serviceuser.ResetPassword
	ResourceOptions.RefreshStateCheck = serviceuser.PasswordIsReady
}

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return serviceuser.Create(ctx, client, d)
}

func flattenModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return serviceuser.PasswordFlatten
}
