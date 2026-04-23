package awsprovision

import (
	"context"
	"errors"
	"net/http"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/byoc"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func init() {
	ResourceOptions.Delete = deleteNoop
}

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	req := new(byoc.CustomCloudEnvironmentProvisionIn)
	err := d.Expand(req)
	if err != nil {
		return err
	}

	rsp, err := client.CustomCloudEnvironmentProvision(
		ctx,
		d.Get("organization_id").(string),
		d.Get("custom_cloud_environment_id").(string),
		req,
	)
	if err != nil {
		if isAlreadyProvisioned(err) {
			return nil
		}
		return err
	}

	return d.Flatten(rsp)
}

func deleteNoop(_ context.Context, _ avngen.Client, _ adapter.ResourceData) error {
	return nil
}

func isAlreadyProvisioned(err error) bool {
	var e avngen.Error
	return errors.As(err, &e) && e.Status == http.StatusConflict
}
