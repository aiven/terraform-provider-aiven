// Package kafkaschemaconfiguration implements the aiven_kafka_schema_configuration resource and data source.
package kafkaschemaconfiguration

import (
	"context"
	"errors"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkaschemaregistry"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var errRegistryDisabled = errors.New("schema registry is disabled")

func registryEnabled(ctx context.Context, client avngen.Client, d adapter.ResourceData) (bool, error) {
	svc, err := schemautil.GetService(ctx, client, d.Get("project").(string), d.Get("service_name").(string))
	if err != nil {
		return false, err
	}
	return svc.UserConfig["schema_registry"] == true, nil
}

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	// Without a level, Create makes no API call, and Read would poll a disabled registry until the timeout.
	enabled, err := registryEnabled(ctx, client, d)
	if err != nil {
		return err
	}
	if !enabled {
		return errRegistryDisabled
	}
	if err := putCompatibilityLevel(ctx, client, d); err != nil {
		return err
	}
	return d.SetID(
		d.Get("project").(string),
		d.Get("service_name").(string),
	)
}

func updateView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return putCompatibilityLevel(ctx, client, d)
}

func readView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	enabled, err := registryEnabled(ctx, client, d)
	if err != nil {
		return err
	}
	if !enabled {
		return fmt.Errorf("%w: %w", adapter.ErrNotFound, errRegistryDisabled)
	}

	level, err := client.ServiceSchemaRegistryGlobalConfigGet(ctx, project, serviceName)
	if err != nil {
		return err
	}

	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("compatibility_level", string(level)); err != nil {
		return err
	}
	return d.SetID(project, serviceName)
}

func putCompatibilityLevel(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	level := d.Get("compatibility_level").(string)
	if level == "" {
		return nil
	}
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	_, err := client.ServiceSchemaRegistryGlobalConfigPut(
		ctx,
		project,
		serviceName,
		&kafkaschemaregistry.ServiceSchemaRegistryGlobalConfigPutIn{
			Compatibility: kafkaschemaregistry.CompatibilityType(level),
		},
	)
	return err
}

// deleteView resets the global compatibility level to BACKWARD. The API has no delete endpoint.
func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	enabled, err := registryEnabled(ctx, client, d)
	if err != nil || !enabled {
		return err
	}

	_, err = client.ServiceSchemaRegistryGlobalConfigPut(
		ctx,
		project,
		serviceName,
		&kafkaschemaregistry.ServiceSchemaRegistryGlobalConfigPutIn{
			Compatibility: kafkaschemaregistry.CompatibilityTypeBackward,
		},
	)
	return err
}
