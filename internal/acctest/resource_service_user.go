package acctest

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// TestAccServiceUserHasCustomPassword validates that service user has a custom password equal to expectedPassword.
func TestAccServiceUserHasCustomPassword(resourceName, expectedPassword string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		// validate local state
		localPassword := rs.Primary.Attributes["password"]
		if localPassword != expectedPassword {
			return fmt.Errorf("local state: password mismatch: got %q, want %q", localPassword, expectedPassword)
		}

		if rs.Primary.Attributes["password_wo"] != "" {
			return fmt.Errorf("local state: password_wo should be empty for custom password mode, got %q", rs.Primary.Attributes["password_wo"])
		}

		if rs.Primary.Attributes["password_wo_version"] != "" && rs.Primary.Attributes["password_wo_version"] != "0" {
			return fmt.Errorf("local state: password_wo_version should be empty or 0 for custom password mode, got %q", rs.Primary.Attributes["password_wo_version"])
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

// TestAccServiceUserHasWOPassword validates that service user has a write-only password.
func TestAccServiceUserHasWOPassword(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		// validate local state
		localPassword := rs.Primary.Attributes["password"]
		if localPassword != "" {
			return fmt.Errorf("local state: password should be empty for write-only mode, got %q", localPassword)
		}

		woVersion := rs.Primary.Attributes["password_wo_version"]
		if woVersion == "" || woVersion == "0" {
			return fmt.Errorf("local state: password_wo_version should be set and non-zero for write-only mode, got %q", woVersion)
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

// TestAccServiceUserHasGeneratedPassword validates that the service user has an auto-generated password.
// Optionally validates password has specified prefix (e.g., "AVN").
func TestAccServiceUserHasGeneratedPassword(resourceName string, expectedPrefix ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		// validate local state
		localPassword := rs.Primary.Attributes["password"]
		if localPassword == "" {
			return fmt.Errorf("local state: password should be set for generated password mode")
		}

		if rs.Primary.Attributes["password_wo"] != "" {
			return fmt.Errorf("local state: password_wo should be empty for generated password mode, got %q", rs.Primary.Attributes["password_wo"])
		}

		woVersion := rs.Primary.Attributes["password_wo_version"]
		if woVersion != "" && woVersion != "0" {
			return fmt.Errorf("local state: password_wo_version should be empty or 0 for generated password mode, got %q", woVersion)
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

		// prefix validation
		if len(expectedPrefix) > 0 && expectedPrefix[0] != "" {
			if !strings.HasPrefix(remotePassword, expectedPrefix[0]) {
				return fmt.Errorf("remote state: password should start with %q for generated password mode, got %q", expectedPrefix[0], remotePassword)
			}
		}

		return nil
	}
}

// fetchRemotePassword retrieves the service user password from the API
func fetchRemotePassword(rs *terraform.ResourceState) (string, error) {
	projectName, serviceName, username, err := schemautil.SplitResourceID3(rs.Primary.ID)
	if err != nil {
		return "", fmt.Errorf("error parsing resource ID: %w", err)
	}

	genClient, err := GetTestGenAivenClient()
	if err != nil {
		return "", fmt.Errorf("error getting generated client: %w", err)
	}

	user, err := genClient.ServiceUserGet(context.Background(), projectName, serviceName, username)
	if err != nil {
		return "", fmt.Errorf("error getting user from API: %w", err)
	}

	return user.Password, nil
}
