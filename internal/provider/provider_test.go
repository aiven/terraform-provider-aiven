package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// DefaultResourceNamePrefix is the default prefix used for resource names in acceptance tests.
	DefaultResourceNamePrefix = "test-acc"

	// DefaultRandomSuffixLength is the default length of the random suffix used in acceptance tests.
	DefaultRandomSuffixLength = 10
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"aiven": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck is a helper function that is called by acceptance tests prior to any test case execution.
// It is used to perform any pre-test setup, such as environment variable validation.
func testAccPreCheck(t *testing.T) {
	if _, ok := os.LookupEnv("AIVEN_TOKEN"); !ok {
		t.Fatal("AIVEN_TOKEN environment variable must be set for acceptance tests.")
	}

	if _, ok := os.LookupEnv("AIVEN_PROJECT_NAME"); !ok {
		t.Log("AIVEN_PROJECT_NAME environment variable is not set. Some acceptance tests will be skipped.")
	}
}
