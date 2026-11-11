package serviceintegrationendpoint

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

type operation = func(context.Context, avngen.Client, adapter.ResourceData) error

func init() {
	ResourceOptions.Update = adapter.UpdateUserConfig("endpoint_type", ResourceOptions.Update)
	ResourceOptions.Read = notFound(ResourceOptions.Read)
	ResourceOptions.Delete = notFound(ResourceOptions.Delete)
	DataSourceOptions.Read = notFound(DataSourceOptions.Read)
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
