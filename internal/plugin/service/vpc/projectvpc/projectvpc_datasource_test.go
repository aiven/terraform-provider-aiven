package projectvpc_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// TestDataSourceConfigValidation drives the `aiven_project_vpc` datasource
// against the noop provider server to exercise its config validators. The
// invalid subtests fail at plan-time on a validator violation; the valid
// subtests pass validation and hit the noop client error during read.
func TestDataSourceConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      string
		expectError *regexp.Regexp
	}{
		{
			name:        "no_discriminator",
			config:      `data "aiven_project_vpc" "x" {}`,
			expectError: regexp.MustCompile(`(?s)Exactly one of these attributes must be configured:\s+\[project_vpc_id,cloud_name,vpc_id\]`),
		},
		{
			name: "both_discriminators",
			config: `data "aiven_project_vpc" "x" {
  vpc_id  = "my-project/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  project = "my-project"
}`,
			expectError: regexp.MustCompile(`(?s)Attribute "project" cannot be specified when "vpc_id" is specified`),
		},
		{
			name: "conflicting_lookup_keys",
			config: `data "aiven_project_vpc" "x" {
  project        = "my-project"
  cloud_name     = "google-europe-west1"
  project_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
}`,
			expectError: regexp.MustCompile(`(?s)Exactly one of these attributes must be configured:\s+\[project_vpc_id,cloud_name,vpc_id\]`),
		},
		{
			name:        "by_vpc_id",
			config:      `data "aiven_project_vpc" "x" { vpc_id = "my-project/1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" }`,
			expectError: acc.ErrNoopErrorRegex,
		},
		{
			name: "by_project_and_cloud_name",
			config: `data "aiven_project_vpc" "x" {
  project    = "my-project"
  cloud_name = "google-europe-west1"
}`,
			expectError: acc.ErrNoopErrorRegex,
		},
		{
			name: "by_project_and_project_vpc_id",
			config: `data "aiven_project_vpc" "x" {
  project        = "my-project"
  project_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
}`,
			expectError: acc.ErrNoopErrorRegex,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acc.NoopProviderServer(),
				Steps: []resource.TestStep{{
					PlanOnly:    true,
					Config:      tc.config,
					ExpectError: tc.expectError,
				}},
			})
		})
	}
}
