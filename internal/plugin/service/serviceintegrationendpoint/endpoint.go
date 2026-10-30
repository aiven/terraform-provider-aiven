package serviceintegrationendpoint

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

type operation = func(context.Context, avngen.Client, adapter.ResourceData) error

func init() {
	ResourceOptions.Read = checkRedactedCreds(notFound(ResourceOptions.Read))
	ResourceOptions.Delete = notFound(ResourceOptions.Delete)
	DataSourceOptions.Read = checkRedactedCreds(notFound(DataSourceOptions.Read))
}

// notFound returns adapter.ErrNotFound for a missing endpoint: the API responds with 400, not 404.
func notFound(op operation) operation {
	return func(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
		err := op(ctx, client, d)
		var e avngen.Error
		if errors.As(err, &e) && e.Status == http.StatusBadRequest && strings.Contains(e.Message, "does not exist") {
			return fmt.Errorf("%w: %w", adapter.ErrNotFound, err)
		}
		return err
	}
}

// checkRedactedCreds fails when the read returns redacted credentials, though it requests secrets.
// Only the read requests secrets, so the Create and Update responses aren't checked.
func checkRedactedCreds(read operation) operation {
	return func(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
		err := read(ctx, client, d)
		if err != nil {
			return err
		}

		config, _ := d.Get("endpoint_config").(map[string]any)
		return schemautil.ContainsRedactedCreds(config)
	}
}
