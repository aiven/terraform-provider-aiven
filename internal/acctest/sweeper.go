package acctest

import (
	"fmt"
	"os"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// SharedClient returns a service Aiven Client setup needed for the sweeper
func SharedClient(region string) (interface{}, error) {
	if os.Getenv("AIVEN_TOKEN") == "" {
		return nil, fmt.Errorf("must provide environment variable AIVEN_TOKEN ")
	}

	// configures a default client, using the above env var
	client, err := aiven.NewTokenClient(os.Getenv("AIVEN_TOKEN"), "terraform-provider-aiven-acc/")
	if err != nil {
		return nil, fmt.Errorf("error getting Aiven client")
	}

	return client, nil
}

func SweepDatabases(region string) error {
	client, err := SharedClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*aiven.Client)

	projects, err := conn.Projects.List()
	if err != nil {
		return fmt.Errorf("error retrieving a list of projects : %s", err)
	}

	for _, project := range projects {
		if project.Name == os.Getenv("AIVEN_PROJECT_NAME") {
			services, err := conn.Services.List(project.Name)
			if err != nil {
				return fmt.Errorf("error retrieving a list of services for a project `%s`: %s", project.Name, err)
			}

			for _, service := range services {
				dbs, err := conn.Databases.List(project.Name, service.Name)
				if err != nil {
					if err.(aiven.Error).Status == 403 || err.(aiven.Error).Status == 501 {
						continue
					}

					return fmt.Errorf("error retrieving a list of databases for a service `%s`: %s", service.Name, err)
				}

				for _, db := range dbs {
					if db.DatabaseName == "defaultdb" {
						continue
					}

					err = conn.Databases.Delete(project.Name, service.Name, db.DatabaseName)
					if err != nil {
						return fmt.Errorf("error destroying database `%s` during sweep: %s", db.DatabaseName, err)
					}
				}
			}
		}
	}

	return nil
}
