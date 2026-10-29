package acctest

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

var (
	cachedVersion string
	versionOnce   sync.Once
)

// GetLastStableVersion returns the most recent published version of the Aiven provider.
// Returns an error if the version cannot be detected from environment or git tags.
func GetLastStableVersion() (string, error) {
	var err error
	versionOnce.Do(func() {
		if envVersion := os.Getenv("AIVEN_BACKWARD_COMPAT_VERSION"); envVersion != "" {
			cachedVersion = envVersion
			return
		}

		cachedVersion, err = gitTags()
	})

	if cachedVersion == "" && err == nil {
		err = fmt.Errorf("version not found: set AIVEN_BACKWARD_COMPAT_VERSION or ensure git tags are available")
	}

	return cachedVersion, err
}

// ExternalAivenProvider returns an ExternalProvider configuration for Aiven.
// For testing against a specific version of the Aiven provider.
func ExternalAivenProvider(t *testing.T, version string) resource.ExternalProvider {
	t.Helper()

	if version == "" || strings.TrimSpace(version) == "latest" {
		v, err := GetLastStableVersion()
		if err != nil {
			t.Fatalf("failed to get last stable version: %v", err)
		}
		version = v
	}

	return resource.ExternalProvider{
		Source:            "aiven/aiven",
		VersionConstraint: version,
	}
}

// gitTags attempts to get the latest version from git tags.
func gitTags() (string, error) {
	// get the latest tag sorted by version
	cmd := exec.Command("git", "tag", "-l", "v*", "--sort=-version:refname")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute git command: %w", err)
	}

	// get the first line (latest tag)
	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		return "", fmt.Errorf("no git tags found matching pattern 'v*'")
	}

	version := strings.TrimSpace(lines[0])
	return strings.TrimPrefix(version, "v"), nil
}

// BackwardCompatConfig holds configuration for backward compatibility tests.
type BackwardCompatConfig struct {
	// TFConfig is the Terraform configuration (same for both steps)
	TFConfig string

	// Checks are the test checks to run in both steps
	Checks resource.TestCheckFunc

	// PreConfig Same as resource.TestStep.PreConfig but for the first step only
	PreConfig func()

	// OldProviderVersion is the version to test against (defaults to latest stable)
	OldProviderVersion string

	// PlanOnly controls whether verification step runs with PlanOnly mode (defaults to false).
	// When false, verification step performs a refresh during plan (`terraform plan`) to allow Read functions to migrate state.
	// When true, verification only compares state to schema without calling Read functions
	// (equivalent to `terraform plan -refresh=false`).
	//
	// e.g. Adding optional fields can be breaking for PlanOnly mode since old state lacks these fields
	// and Read function doesn't run to migrate them. This may affect some CI pipelines using `plan -refresh=false`.
	PlanOnly bool

	// ExpectNonEmptyPlan allows the verification step to change resources, which ConfigPlanChecks should narrow down,
	// e.g., a service drops the user config options SDKv2 stored while the resource under test doesn't change.
	ExpectNonEmptyPlan bool

	// ConfigPlanChecks are the plan checks of the verification step.
	ConfigPlanChecks resource.ConfigPlanChecks
}

// BackwardCompatibilitySteps creates a two-step backward compatibility test:
// Step 1: Create resources with old (published) provider version
// Step 2: Verify new provider can read state without changes
//
// When switching from old provider to new provider with SAME config,
// the plan MUST be empty (no changes needed), unless ExpectNonEmptyPlan is set.
func BackwardCompatibilitySteps(t *testing.T, config BackwardCompatConfig) []resource.TestStep {
	t.Helper()

	version := config.OldProviderVersion
	if version == "" {
		v, err := GetLastStableVersion()
		if err != nil {
			t.Fatalf("failed to get last stable version: %v", err)
		}
		version = v
	}

	return []resource.TestStep{
		// create resources with old provider version
		{
			ExternalProviders: map[string]resource.ExternalProvider{
				"aiven": ExternalAivenProvider(t, version),
			},
			Config:    config.TFConfig,
			Check:     config.Checks,
			PreConfig: config.PreConfig,
		},
		// switch to new provider and verify plan is empty
		{
			ProtoV6ProviderFactories: TestProtoV6ProviderFactories,
			Config:                   config.TFConfig,
			PlanOnly:                 config.PlanOnly,
			ExpectNonEmptyPlan:       config.ExpectNonEmptyPlan,
			ConfigPlanChecks:         config.ConfigPlanChecks,
			Check:                    config.Checks,
		},
	}
}

// UpgradePlanChecks expects both plans of a PlanOnly verification step to keep unchanged as is
// and to update updated in place, e.g., services that drop the user config options SDKv2 stored.
func UpgradePlanChecks(unchanged, updated []string) resource.ConfigPlanChecks {
	checks := make([]plancheck.PlanCheck, 0, len(unchanged)+len(updated))
	for _, v := range unchanged {
		checks = append(checks, plancheck.ExpectResourceAction(v, plancheck.ResourceActionNoop))
	}
	for _, v := range updated {
		checks = append(checks, plancheck.ExpectResourceAction(v, plancheck.ResourceActionUpdate))
	}
	return resource.ConfigPlanChecks{PostApplyPreRefresh: checks, PostApplyPostRefresh: checks}
}
