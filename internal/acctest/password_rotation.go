package acctest

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// ServicePasswordTestOptions configures the password rotation test scenario
type ServicePasswordTestOptions struct {
	// ResourceType is the Terraform resource type (e.g., "aiven_mysql", "aiven_pg")
	ResourceType string

	// Username is the expected service username (e.g., "avnadmin", "default", "custom")
	Username string

	// ServicePrefix is used to generate service names (e.g., "test-acc-mysql")
	// If empty, will be auto-generated from ResourceType.
	// Only used if ConfigBasic and ConfigWO are nil
	ServicePrefix string

	// CloudName for generated configs (default: "google-europe-west1")
	CloudName string

	// Plan for generated configs (default: "startup-4")
	Plan string

	// ConfigBasic is an optional custom config function for service without password
	// If nil, a default config will be generated using ServicePrefix, CloudName, and Plan
	ConfigBasic func(name string) string

	// ConfigWO is an optional custom config function for service with write-only password field
	// If nil, a default config will be generated using ServicePrefix, CloudName, and Plan
	ConfigWO func(name, password string, version int) string
}

// TestAccCheckAivenServiceWriteOnlyPassword tests the full lifecycle of write-only passwords
func TestAccCheckAivenServiceWriteOnlyPassword(t *testing.T, opts ServicePasswordTestOptions) {
	t.Skip("This test will be enabled once services support write-only passwords")

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("%s.test", opts.ResourceType)
	password1 := acctest.RandStringFromCharSet(12, acctest.CharSetAlphaNum)
	password2 := acctest.RandStringFromCharSet(12, acctest.CharSetAlphaNum)
	password3 := acctest.RandStringFromCharSet(12, acctest.CharSetAlphaNum)

	cb := configBasic(opts)
	cWO := configWO(opts)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestProtoV6ProviderFactories,
		CheckDestroy:             TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				// Step 1: Create service with auto-generated password
				Config: cb(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_username", opts.Username),
					resource.TestCheckResourceAttrSet(resourceName, "service_password"), // verify password is in state
					resource.TestCheckNoResourceAttr(resourceName, "service_password_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "service_password_wo_version"),
					TestAccPasswordHasGeneratedPassword(resourceName),
				),
			},
			{
				// Step 2: Migrate to write-only password with explicit value
				Config: cWO(rName, password1, 1),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						ExpectOnlyAttributesChanged(resourceName, "service_password", "service_password_wo_version"),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_username", opts.Username),
					TestAccPasswordHasWOPassword(resourceName),
					TestAccPasswordMatches(resourceName, password1), // verify actual password
				),
			},
			{
				// Step 3: Rotate password by incrementing version
				Config: cWO(rName, password2, 2),
				Check: resource.ComposeTestCheckFunc(
					TestAccPasswordHasWOPassword(resourceName),
					TestAccPasswordMatches(resourceName, password2), // verify password rotated
				),
			},
			{
				// Step 4: Another rotation
				Config: cWO(rName, password3, 3),
				Check: resource.ComposeTestCheckFunc(
					TestAccPasswordHasWOPassword(resourceName),
					TestAccPasswordMatches(resourceName, password3),
				),
			},
			{
				// Step 5: Try to decrement password version (should trigger error)
				Config:             cWO(rName, password2, 2),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("must be incremented .* Decrementing version is not allowed"),
			},
			{
				// Step 6: Change password without incrementing version (should be ignored)
				Config: cWO(rName, password1, 3),
				Check: resource.ComposeTestCheckFunc(
					TestAccPasswordHasWOPassword(resourceName),
					TestAccPasswordMatches(resourceName, password3), // matches previous password
					resource.TestCheckResourceAttr(resourceName, "service_password_wo_version", "3"),
				),
			},
			{
				// Step 7: Test short password (should trigger validation error)
				Config:      cWO(rName, "1234", 4),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("expected length of service_password_wo to be in the range .8 - 256."),
			},
			{
				// Step 8: Switch back to auto-generated password
				Config: cb(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_username", opts.Username),
					resource.TestCheckResourceAttrSet(resourceName, "service_password"), // password back in state
					resource.TestCheckNoResourceAttr(resourceName, "service_password_wo"),
					resource.TestCheckResourceAttr(resourceName, "service_password_wo_version", "0"), // version reset to 0
					TestAccPasswordHasGeneratedPassword(resourceName),
					TestAccPasswordNotMatches(resourceName, password3), // verify password changed from the custom one
				),
			},
		},
	})
}

// TestAccPasswordHasWOPassword validates that resource has a write-only password.
// Works for both services and service_users.
func TestAccPasswordHasWOPassword(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		info, err := extractResourceInfo(rs)
		if err != nil {
			return err
		}

		// validate local state
		localPassword := rs.Primary.Attributes[info.passwordField]
		if localPassword != "" {
			return fmt.Errorf("local state: %s should be empty for write-only mode, got %q", info.passwordField, localPassword)
		}

		woVersion := rs.Primary.Attributes[info.passwordWoVerField]
		if woVersion == "" || woVersion == "0" {
			return fmt.Errorf("local state: %s should be set and non-zero for write-only mode, got %q", info.passwordWoVerField, woVersion)
		}

		// validate API state
		remotePassword, err := fetchRemotePassword(rs)
		if err != nil {
			return err
		}

		if remotePassword == "" {
			return fmt.Errorf("remote state: password should exist in API for write-only mode")
		}

		return nil
	}
}

// TestAccPasswordMatches validates that the password matches the expected value.
// Works for both services and service_users.
func TestAccPasswordMatches(resourceName, expectedPassword string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		info, err := extractResourceInfo(rs)
		if err != nil {
			return err
		}

		// For write-only mode, local password should be empty
		localPassword := rs.Primary.Attributes[info.passwordField]
		woVersion := rs.Primary.Attributes[info.passwordWoVerField]
		isWOMode := woVersion != "" && woVersion != "0"

		if isWOMode && localPassword != "" {
			return fmt.Errorf("local state: %s should be empty when using write-only password, got %q", info.passwordField, localPassword)
		}

		// validate API state
		remotePassword, err := fetchRemotePassword(rs)
		if err != nil {
			return err
		}

		if remotePassword != expectedPassword {
			return fmt.Errorf(
				"remote state: password mismatch: got %q, want %q",
				remotePassword,
				expectedPassword,
			)
		}

		return nil
	}
}

// TestAccPasswordNotMatches validates that the password has changed from a previous value.
// Works for both services and service_users.
func TestAccPasswordNotMatches(resourceName, oldPassword string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		// validate API state
		remotePassword, err := fetchRemotePassword(rs)
		if err != nil {
			return err
		}

		if remotePassword == oldPassword {
			return fmt.Errorf("remote state: password still matches old value: %q", oldPassword)
		}

		return nil
	}
}

// TestAccPasswordHasGeneratedPassword validates that the resource has an auto-generated password.
// Optionally validates password has specified prefix (e.g., "AVN").
// Works for both services and service_users.
func TestAccPasswordHasGeneratedPassword(resourceName string, expectedPrefix ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		info, err := extractResourceInfo(rs)
		if err != nil {
			return err
		}

		// validate local state
		localPassword := rs.Primary.Attributes[info.passwordField]
		if localPassword == "" {
			return fmt.Errorf("local state: %s should be set for generated password mode", info.passwordField)
		}

		woVersion := rs.Primary.Attributes[info.passwordWoVerField]
		if woVersion != "" && woVersion != "0" {
			return fmt.Errorf("local state: %s should be empty or 0 for generated password mode, got %q", info.passwordWoVerField, woVersion)
		}

		// validate API state
		remotePassword, err := fetchRemotePassword(rs)
		if err != nil {
			return err
		}

		if remotePassword == "" {
			return fmt.Errorf("remote state: password should exist in API for generated password mode")
		}

		if remotePassword != localPassword {
			return fmt.Errorf("remote state: password mismatch with local state: got %q, want %q", remotePassword, localPassword)
		}

		// prefix validation (optional)
		if len(expectedPrefix) > 0 && expectedPrefix[0] != "" {
			if !strings.HasPrefix(remotePassword, expectedPrefix[0]) {
				return fmt.Errorf("remote state: password should start with %q for generated password mode, got %q", expectedPrefix[0], remotePassword)
			}
		}

		return nil
	}
}

// TestAccPasswordHasCustomPassword validates that resource has a custom password equal to expectedPassword.
// Only works for service_users (services don't support custom passwords).
func TestAccPasswordHasCustomPassword(resourceName, expectedPassword string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		info, err := extractResourceInfo(rs)
		if err != nil {
			return err
		}

		if !info.isServiceUser {
			return fmt.Errorf("custom password mode is only supported for service_user resources")
		}

		// validate local state
		localPassword := rs.Primary.Attributes[info.passwordField]
		if localPassword != expectedPassword {
			return fmt.Errorf("local state: password mismatch: got %q, want %q", localPassword, expectedPassword)
		}

		if rs.Primary.Attributes[info.passwordWoField] != "" {
			return fmt.Errorf("local state: %s should be empty for custom password mode, got %q", info.passwordWoField, rs.Primary.Attributes[info.passwordWoField])
		}

		woVersion := rs.Primary.Attributes[info.passwordWoVerField]
		if woVersion != "" && woVersion != "0" {
			return fmt.Errorf("local state: %s should be empty or 0 for custom password mode, got %q", info.passwordWoVerField, woVersion)
		}

		// validate API state
		remotePassword, err := fetchRemotePassword(rs)
		if err != nil {
			return err
		}

		if remotePassword != expectedPassword {
			return fmt.Errorf(
				"remote state: password mismatch: got %q, want %q",
				remotePassword,
				expectedPassword,
			)
		}

		return nil
	}
}

// resourceInfo contains extracted information about a service or service_user resource
type resourceInfo struct {
	projectName        string
	serviceName        string
	username           string
	passwordField      string
	passwordWoField    string
	passwordWoVerField string
	isServiceUser      bool
}

// extractResourceInfo detects resource type and extracts relevant information
func extractResourceInfo(rs *terraform.ResourceState) (*resourceInfo, error) {
	if strings.HasSuffix(rs.Type, "_user") {
		// service_user resource has 3-part ID
		projectName, serviceName, username, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err == nil {
			return &resourceInfo{
				projectName:        projectName,
				serviceName:        serviceName,
				username:           username,
				passwordField:      "password",
				passwordWoField:    "password_wo",
				passwordWoVerField: "password_wo_version",
				isServiceUser:      true,
			}, nil
		}
	}

	// check if it's a service resource (should have service_type in attributes)
	serviceType, ok := rs.Primary.Attributes["service_type"]
	if !ok || serviceType == "" {
		return nil, fmt.Errorf("unsupported resource type %q: must be a service or a service user resource", rs.Type)
	}

	// service resource has 2-part ID
	projectName, serviceName, err := schemautil.SplitResourceID2(rs.Primary.ID)
	if err != nil {
		return nil, fmt.Errorf("error parsing service ID: %w", err)
	}

	username := rs.Primary.Attributes["service_username"]
	if username == "" {
		username = schemautil.DefaultServiceUsername(serviceType)
	}

	return &resourceInfo{
		projectName:        projectName,
		serviceName:        serviceName,
		username:           username,
		passwordField:      "service_password",
		passwordWoField:    "service_password_wo",
		passwordWoVerField: "service_password_wo_version",
		isServiceUser:      false,
	}, nil
}

// fetchRemotePassword retrieves the password from the API for either service or service_user
func fetchRemotePassword(rs *terraform.ResourceState) (string, error) {
	info, err := extractResourceInfo(rs)
	if err != nil {
		return "", err
	}

	genClient, err := GetTestGenAivenClient()
	if err != nil {
		return "", fmt.Errorf("error getting generated client: %w", err)
	}

	var password string
	err = retry.Do(
		func() error {
			user, err := genClient.ServiceUserGet(context.Background(), info.projectName, info.serviceName, info.username)
			if err != nil {
				return fmt.Errorf("error getting %s user from API: %w", info.username, err)
			}
			if user.Password == "" {
				return fmt.Errorf("fetched password for user %s is empty", info.username)
			}
			password = user.Password
			return nil
		},
		retry.Attempts(10),
		retry.Delay(1*time.Second),
	)
	if err != nil {
		return "", fmt.Errorf("failed to fetch non-empty password after retries: %w", err)
	}

	return password, nil
}

// servicePrefix returns the service prefix to use, either from opts or auto-generated
func servicePrefix(opts ServicePasswordTestOptions) string {
	if opts.ServicePrefix != "" {
		return opts.ServicePrefix
	}

	return "test-acc-" + strings.TrimPrefix(opts.ResourceType, "aiven_")
}

// configBasic returns the config function to use (custom or generated)
func configBasic(opts ServicePasswordTestOptions) func(string) string {
	if opts.ConfigBasic != nil {
		return opts.ConfigBasic
	}

	return func(name string) string {
		return defaultServiceConfig(opts, name)
	}
}

// configWO returns the config function to use (custom or generated)
func configWO(opts ServicePasswordTestOptions) func(string, string, int) string {
	if opts.ConfigWO != nil {
		return opts.ConfigWO
	}

	return func(name, password string, version int) string {
		return defaultServiceWOConfig(opts, name, password, version)
	}
}

// defaultServiceConfig creates a default service configuration without password fields
func defaultServiceConfig(opts ServicePasswordTestOptions, name string) string {
	cloudName := opts.CloudName
	if cloudName == "" {
		cloudName = "google-europe-west1"
	}
	plan := opts.Plan
	if plan == "" {
		plan = "startup-4"
	}
	sp := servicePrefix(opts)

	return fmt.Sprintf(`
resource "%s" "test" {
  project                 = "%s"
  cloud_name              = "%s"
  plan                    = "%s"
  service_name            = "%s-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}
`, opts.ResourceType, ProjectName(), cloudName, plan, sp, name)
}

// defaultServiceWOConfig creates a default service configuration with write-only password field
func defaultServiceWOConfig(opts ServicePasswordTestOptions, name, password string, version int) string {
	cloudName := opts.CloudName
	if cloudName == "" {
		cloudName = "google-europe-west1"
	}
	plan := opts.Plan
	if plan == "" {
		plan = "startup-4"
	}
	sp := servicePrefix(opts)

	return fmt.Sprintf(`
resource "%s" "test" {
  project                     = "%s"
  cloud_name                  = "%s"
  plan                        = "%s"
  service_name                = "%s-%s"
  maintenance_window_dow      = "monday"
  maintenance_window_time     = "10:00:00"
  service_password_wo         = "%s"
  service_password_wo_version = %d
}
`, opts.ResourceType, ProjectName(), cloudName, plan, sp, name, password, version)
}
