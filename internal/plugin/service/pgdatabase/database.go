package pgdatabase

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func init() {
	ResourceOptions.Create = create
	ResourceOptions.Read = readView
	ResourceOptions.Delete = deleteAndForget
	DataSourceOptions.Read = readView
}

func create(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	err := schemautil.CheckDbConflict(ctx, client, d.Get("project").(string), d.Get("service_name").(string), schemautil.ServiceTypePG, d.Get("database_name").(string))
	if err != nil {
		return err
	}

	err = createView(ctx, client, d)
	if err != nil {
		return err
	}

	// We have already checked for the existence of the database.
	// Getting a conflict here means the client retried the request.
	if avngen.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// deleteAndForget forgets the database, so CheckDbConflict doesn't report it when it's created again.
// deleteStateGone polls readView until it's gone.
func deleteAndForget(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	err := deleteView(ctx, client, d)
	if err == nil || avngen.IsNotFound(err) {
		schemautil.ForgetDatabase(d.Get("project").(string), d.Get("service_name").(string), d.Get("database_name").(string))
	}
	return err
}
