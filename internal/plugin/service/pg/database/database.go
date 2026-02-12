package database

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
	ResourceOptions.Read = read
	ResourceOptions.Delete = delete
	DataSourceOptions.Read = read
}

func create(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	err := schemautil.CheckDbConflict(ctx, client, d.Get("project").(string), d.Get("service_name").(string), d.Get("database_name").(string))
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

func read(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	err := schemautil.CheckServiceIsPowered(ctx, client, d.Get("project").(string), d.Get("service_name").(string))
	if err != nil {
		return fmt.Errorf("service is powered off: %w", err)
	}

	return readView(ctx, client, d)
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
	err := schemautil.CheckServiceIsPowered(ctx, client, project, serviceName)
	if err != nil {
		return nil, err
	}

	list, err := client.ServiceDatabaseList(ctx, project, serviceName)
	if err != nil {
		return nil, err
	}

	for _, db := range list {
		if db.DatabaseName == dbName {
			return &db, nil
		}
	}

	return nil, avngen.Error{
		Message:     fmt.Sprintf("`%s` with given `database_name` not found", typeName),
		OperationID: "ServiceDatabaseList",
		Status:      404,
	}
}
