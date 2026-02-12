package application

import (
	"context"
	"fmt"
	"net/http"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// planModifier resolves application name to application_id for the data source.
// When name is provided instead of application_id, it lists all applications and finds the matching one.
func planModifier(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if d.Get("application_id").(string) != "" {
		return nil
	}

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	name := d.Get("name").(string)
	apps, err := client.ServiceFlinkListApplications(ctx, project, serviceName)
	if err != nil {
		return err
	}

	for _, app := range apps {
		if app.Name == name {
			err = d.Set("application_id", app.Id)
			if err != nil {
				return err
			}

			return d.SetID(project, serviceName, app.Id)
		}
	}

	return &avngen.Error{
		Status:      http.StatusNotFound,
		OperationID: "ServiceFlinkListApplications",
		Message:     fmt.Sprintf("flink application %q not found in project %q, service %q", name, project, serviceName),
	}
}
