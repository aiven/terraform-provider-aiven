package pgdatabase

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func init() {
	ResourceOptions.Create = create
	ResourceOptions.Read = readView
	ResourceOptions.Delete = delete
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

func delete(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	dbName := d.Get("database_name").(string)

	err := deleteView(ctx, client, d)
	if avngen.IsNotFound(err) {
		// The resource is already gone.
		schemautil.ForgetDatabase(project, serviceName, dbName)
		return nil
	}
	if err != nil {
		return err
	}

	// Waits until database is deleted.
	err = schemautil.WaitUntilNotFound(ctx, func() error {
		_, err := findDatabaseByName(ctx, client, project, serviceName, dbName)
		if err == nil {
			return fmt.Errorf("database %q still exists", dbName)
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("waiting for database deletion failed: %w", err)
	}

	schemautil.ForgetDatabase(project, serviceName, dbName)
	return nil
}

func findDatabaseByName(ctx context.Context, client avngen.Client, project, serviceName, dbName string) (*service.DatabaseOut, error) {
	err := schemautil.CheckServiceReady(ctx, client, project, serviceName, schemautil.ServiceTypePG)
	if err != nil {
		return nil, err
	}

	databases, err := schemautil.ListServiceDatabases(ctx, client, project, serviceName)
	if err != nil {
		return nil, err
	}

	for i := range databases {
		if databases[i].DatabaseName == dbName {
			return &databases[i], nil
		}
	}

	return nil, avngen.Error{
		Message:     fmt.Sprintf("`%s` with given `database_name` not found", typeName),
		OperationID: "ServiceDatabaseList",
		Status:      404,
	}
}
