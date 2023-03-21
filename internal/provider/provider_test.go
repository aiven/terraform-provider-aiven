// Package provider is the implementation of the Aiven provider.
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
// nolint:unused // TODO: Remove this once we have acceptance tests.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"aiven": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck is a helper function that is called by acceptance tests prior to any test case execution.
// It is used to perform any pre-test setup, such as environment variable validation.
// nolint:unused // TODO: Remove this once we have acceptance tests.
func testAccPreCheck(t *testing.T) {}
