package aiven

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

// sharedClient returns a common Aiven Client setup needed for the sweeper
func sharedClient(region string) (interface{}, error) {
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
