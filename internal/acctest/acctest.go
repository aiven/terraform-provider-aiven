package acctest

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/server"
)

var (
	testAivenClient              *aiven.Client
	testAivenClientOnce          sync.Once
	TestProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"aiven": func() (tfprotov6.ProviderServer, error) {
			return server.NewMuxServer(context.Background(), "test")
		},
	}
)

var (
	// ErrMustSetBetaEnvVar is an error that is returned when the PROVIDER_AIVEN_ENABLE_BETA environment variable is not
	// set, but it is required for the concrete acceptance test to run.
	ErrMustSetBetaEnvVar = "PROVIDER_AIVEN_ENABLE_BETA must be set for this test to run"

	// ErrMustSetOrganizationUserIDEnvVar is an error that is returned when the AIVEN_ORGANIZATION_USER_ID environment
	// variable is not set, but it is required for the concrete acceptance test to run.
	ErrMustSetOrganizationUserIDEnvVar = "AIVEN_ORGANIZATION_USER_ID must be set for this test to run"
)

// GetTestAivenClient returns a new Aiven client that can be used for acceptance tests.
func GetTestAivenClient() *aiven.Client {
	testAivenClientOnce.Do(func() {
		client, err := common.NewAivenClient()
		if err != nil {
			log.Fatal(err)
		}
		testAivenClient = client
	})
	return testAivenClient
}

// commonTestDependencies is a struct that contains common dependencies that are used by acceptance tests.
type commonTestDependencies struct {
	// t is the testing.T instance that is used for acceptance tests.
	t *testing.T

	// isBeta is a flag that indicates whether the provider is in beta mode.
	isBeta bool
	// organizationName is the name of the organization that is used for acceptance tests.
	organizationName string
	// organizationUserID is the ID of the organization user that is used for acceptance tests.
	organizationUserID *string
}

// IsBeta returns a flag that indicates whether the provider is in beta mode.
// If skip is true, then this function will skip the test if the provider is not in beta mode.
func (d *commonTestDependencies) IsBeta(skip bool) bool {
	if skip && !d.isBeta {
		d.t.Skip(ErrMustSetBetaEnvVar)
	}

	return d.isBeta
}

// OrganizationName returns the name of the organization that is used for acceptance tests.
func (d *commonTestDependencies) OrganizationName() string {
	return d.organizationName
}

// OrganizationUserID returns the ID of the organization user that is used for acceptance tests.
// If skip is true, then this function will skip the test if the organization user ID is not set.
func (d *commonTestDependencies) OrganizationUserID(skip bool) *string {
	if skip && d.organizationUserID == nil {
		d.t.Skip(ErrMustSetOrganizationUserIDEnvVar)
	}

	return d.organizationUserID
}

// CommonTestDependencies returns a new commonTestDependencies struct that contains common dependencies that are
// used by acceptance tests.
// nolint:revive // Ignore unexported type error because this type is not meant to be used outside of this package.
func CommonTestDependencies(t *testing.T) *commonTestDependencies {
	// We mimic the real error message that is returned by Terraform when the acceptance tests are skipped.
	//
	// This is done because the tests that use this function are running it before the real Terraform check takes
	// place, and we want to avoid false positively running this function when the acceptance tests are not actually
	// ran, e.g. if unit tests are ran instead.
	//
	// See https://github.com/hashicorp/terraform-plugin-testing/blob/v1.6.0/helper/resource/testing.go#L849-L857 for
	// more details on the real check.
	if _, ok := os.LookupEnv("TF_ACC"); !ok {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	deps := &commonTestDependencies{
		t: t,

		isBeta: util.IsBeta(),
	}

	organizationName, ok := os.LookupEnv("AIVEN_ORGANIZATION_NAME")
	if !ok {
		t.Fatal("AIVEN_ORGANIZATION_NAME environment variable must be set for acceptance tests")
	}
	deps.organizationName = organizationName

	organizationUserID, ok := os.LookupEnv("AIVEN_ORGANIZATION_USER_ID")
	if ok {
		deps.organizationUserID = &organizationUserID
	}

	return deps
}

const (
	// DefaultResourceNamePrefix is the default prefix used for resource names in acceptance tests.
	DefaultResourceNamePrefix = "test-acc"

	// DefaultRandomSuffixLength is the default length of the random suffix used in acceptance tests.
	DefaultRandomSuffixLength = 10
)

func RandStr() string {
	return acctest.RandStringFromCharSet(DefaultRandomSuffixLength, acctest.CharSetAlphaNum)
}

// TestAccPreCheck is a helper function that is called by acceptance tests prior to any test case execution.
// It is used to perform any pre-test setup, such as environment variable validation.
func TestAccPreCheck(t *testing.T) {
	if _, ok := os.LookupEnv("AIVEN_TOKEN"); !ok {
		t.Fatal("AIVEN_TOKEN environment variable must be set for acceptance tests")
	}

	if _, ok := os.LookupEnv("AIVEN_PROJECT_NAME"); !ok {
		t.Log("AIVEN_PROJECT_NAME environment variable is not set. Some acceptance tests will be skipped")
	}
}

// TestAccCheckAivenServiceResourceDestroy verifies that the given service is destroyed.
func TestAccCheckAivenServiceResourceDestroy(s *terraform.State) error {
	c := GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each service is destroyed
	for n, rs := range s.RootModule().Resources {
		// ignore datasource
		if strings.HasPrefix(n, "data.") {
			continue
		}

		r := func() []string {
			return []string{
				"aiven_influxdb",
				"aiven_grafana",
				"aiven_mysql",
				"aiven_redis",
				"aiven_pg",
				"aiven_cassandra",
				"aiven_m3db",
				"aiven_flink",
				"aiven_opensearch",
				"aiven_kafka",
				"aiven_kafka_connector",
				"aiven_kafka_connect",
				"aiven_clickhouse",
			}
		}
		if sort.SearchStrings(r(), rs.Type) > 0 {
			continue
		}

		projectName, serviceName, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		p, err := c.Services.Get(ctx, projectName, serviceName)
		if err != nil {
			if !aiven.IsNotFound(err) {
				return err
			}
		}

		if p != nil {
			return fmt.Errorf("common (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

// ResourceFromState returns a resource state from the given Terraform state.
func ResourceFromState(state *terraform.State, name string) (*terraform.ResourceState, error) {
	rs, ok := state.RootModule().Resources[name]
	if !ok {
		return nil, fmt.Errorf(errmsg.ResourceNotFound, name)
	}

	return rs, nil
}
