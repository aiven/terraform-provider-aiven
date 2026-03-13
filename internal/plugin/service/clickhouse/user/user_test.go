package user_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	ch "github.com/aiven/go-client-codegen/handler/clickhouse"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenClickHouseUser(t *testing.T) {
	projectName := acc.ProjectName()
	serviceName := acc.RandName("clickhouse")
	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		serviceName,
		acc.WithServiceType("clickhouse"),
		acc.WithPlan("startup-8"),
		acc.WithCloud("google-europe-west1"),
	)

	client, err := acc.GetTestGenAivenClient()
	require.NoError(t, err)

	t.Run("create user without password", func(t *testing.T) {
		resourceName := "aiven_clickhouse_user.foo"
		userName := acc.RandName("user")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenClickHouseUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccClickHouseUserWithoutPassword(projectName, serviceName, userName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "username", userName),
						resource.TestCheckResourceAttrSet(resourceName, "uuid"),
						testAccCheckClickHouseUserHasGeneratedPassword(resourceName),
					),
				},
			},
		})
	})

	t.Run("basic lifecycle", func(t *testing.T) {
		resourceName := "aiven_clickhouse_user.foo"
		userName := acc.RandName("user")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenClickHouseUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					// Start with an explicit password so the later transitions begin from a stable state.
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccClickHouseUserWithPassword(projectName, serviceName, userName, "Test$1234"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "username", userName),
						testAccCheckClickHouseUserHasCustomPassword(resourceName, "Test$1234"),
						resource.TestCheckResourceAttr("data.aiven_clickhouse_user.by_username", "username", userName),
						resource.TestCheckResourceAttrPair("data.aiven_clickhouse_user.by_uuid", "uuid", resourceName, "uuid"),
						resource.TestCheckResourceAttrPair("data.aiven_clickhouse_user.by_uuid", "username", resourceName, "username"),
					),
				},
				{
					// Remove the password from config and check that the existing user can return to a generated password.
					Config: testAccClickHouseUserWithoutPassword(projectName, serviceName, userName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckClickHouseUserHasGeneratedPassword(resourceName),
					),
				},
				{
					// Set an explicit password again to cover the path back from a generated password.
					Config: testAccClickHouseUserWithPassword(projectName, serviceName, userName, "Custom$Pass456"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckClickHouseUserHasCustomPassword(resourceName, "Custom$Pass456"),
					),
				},
				{
					// Import from a stable state to check that the runtime ID, which includes the server-side UUID, is enough.
					// The password is ignored here because ClickHouse import cannot reconstruct it from the API.
					Config:                  testAccClickHouseUserWithPassword(projectName, serviceName, userName, "Custom$Pass456"),
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"password"},
					ImportStateIdFunc: func(state *terraform.State) (string, error) {
						rs, ok := state.RootModule().Resources[resourceName]
						if !ok {
							return "", fmt.Errorf("expected resource %q to be present in state", resourceName)
						}
						return rs.Primary.ID, nil
					},
				},
				{
					// Switch to write-only mode and check that the password is no longer stored in state.
					Config: testAccClickHouseUserWriteOnly(projectName, serviceName, userName, "WriteOnlyPass$1", 1),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckClickHouseUserHasWOPassword(resourceName),
					),
				},
				{
					// Rotate the write-only password to confirm that changing only the version starts another update.
					Config: testAccClickHouseUserWriteOnly(projectName, serviceName, userName, "WriteOnlyPass$2", 2),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckClickHouseUserHasWOPassword(resourceName),
					),
				},
				{
					// Return to a generated password and check that write-only markers are cleaned from state.
					Config: testAccClickHouseUserWithoutPassword(projectName, serviceName, userName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckClickHouseUserHasGeneratedPassword(resourceName),
					),
				},
			},
		})
	})

	t.Run("remove missing", func(t *testing.T) {
		resourceName := "aiven_clickhouse_user.foo"
		userName := acc.RandName("user")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenClickHouseUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					// Create the user first so the test starts from a normal managed state.
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccClickHouseUserWithoutPasswordResourceOnly(projectName, serviceName, userName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(resourceName, "id"),
					),
				},
				{
					// Delete the user outside Terraform and check that the next plan notices the resource is gone.
					Config: testAccClickHouseUserWithoutPasswordResourceOnly(projectName, serviceName, userName),
					PreConfig: func() {
						user, err := findClickHouseUserByName(t.Context(), client, projectName, serviceName, userName)
						require.NoError(t, err)

						err = client.ServiceClickHouseUserDelete(t.Context(), projectName, serviceName, user.Uuid)
						require.NoError(t, err)
					},
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
				},
				{
					// Apply the same config again and make sure Terraform recreates the missing user.
					Config: testAccClickHouseUserWithoutPasswordResourceOnly(projectName, serviceName, userName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(resourceName, "id"),
						func(state *terraform.State) error {
							_, err := findClickHouseUserByName(t.Context(), client, projectName, serviceName, userName)
							return err
						},
					),
				},
			},
		})
	})

	t.Run("backward compatibility", func(t *testing.T) {
		userName := acc.RandName("user")
		resourceName := "aiven_clickhouse_user.test"

		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() { acc.TestAccPreCheck(t) },
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				// Create the user with the old provider and check that the migrated resource can read
				// the same state without losing important fields.
				PreConfig: func() {
					require.NoError(t, <-serviceIsReady)
				},
				TFConfig: testAccClickHouseUserBackwardCompatibility(projectName, serviceName, userName),
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "username", userName),
					resource.TestCheckResourceAttrSet(resourceName, "password"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttrSet(resourceName, "required"),
					resource.TestCheckResourceAttr("data.aiven_clickhouse_user.test", "username", userName),
					resource.TestCheckResourceAttrSet("data.aiven_clickhouse_user.test", "uuid"),
				),
				OldProviderVersion: "4.47.0",
			}),
		})
	})
}

func testAccCheckAivenClickHouseUserResourceDestroy(s *terraform.State) error {
	client, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_clickhouse_user" {
			continue
		}

		project, serviceName, uuid, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		users, err := client.ServiceClickHouseUserList(context.Background(), project, serviceName)
		if err != nil {
			return err
		}

		if slices.ContainsFunc(users, func(user ch.UserOut) bool { return user.Uuid == uuid }) {
			return fmt.Errorf("clickhouse user %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func findClickHouseUserByName(ctx context.Context, client avngen.Client, project, serviceName, userName string) (*ch.UserOut, error) {
	users, err := client.ServiceClickHouseUserList(ctx, project, serviceName)
	if err != nil {
		return nil, err
	}

	for i := range users {
		if users[i].Name == userName {
			return &users[i], nil
		}
	}

	return nil, fmt.Errorf("clickhouse user %q not found in service %q", userName, serviceName)
}

func testAccCheckClickHouseUserHasGeneratedPassword(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		localPassword := rs.Primary.Attributes["password"]
		if localPassword == "" {
			return fmt.Errorf("local state: password should be set for generated password mode")
		}

		woVersion := rs.Primary.Attributes["password_wo_version"]
		if woVersion != "" && woVersion != "0" {
			return fmt.Errorf("local state: password_wo_version should be empty or 0 for generated password mode, got %q", woVersion)
		}

		return nil
	}
}

func testAccCheckClickHouseUserHasCustomPassword(resourceName, expectedPassword string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		localPassword := rs.Primary.Attributes["password"]
		if localPassword != expectedPassword {
			return fmt.Errorf("local state: password mismatch: got %q, want %q", localPassword, expectedPassword)
		}
		if rs.Primary.Attributes["password_wo"] != "" {
			return fmt.Errorf("local state: password_wo should be empty for custom password mode, got %q", rs.Primary.Attributes["password_wo"])
		}

		woVersion := rs.Primary.Attributes["password_wo_version"]
		if woVersion != "" && woVersion != "0" {
			return fmt.Errorf("local state: password_wo_version should be empty or 0 for custom password mode, got %q", woVersion)
		}

		return nil
	}
}

func testAccCheckClickHouseUserHasWOPassword(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.Attributes["password"] != "" {
			return fmt.Errorf("local state: password should be empty for write-only mode, got %q", rs.Primary.Attributes["password"])
		}

		woVersion := rs.Primary.Attributes["password_wo_version"]
		if woVersion == "" || woVersion == "0" {
			return fmt.Errorf("local state: password_wo_version should be set and non-zero for write-only mode, got %q", woVersion)
		}

		return nil
	}
}

func testAccClickHouseUserWithPassword(project, serviceName, userName, password string) string {
	return fmt.Sprintf(`
resource "aiven_clickhouse_user" "foo" {
  project      = %[1]q
  service_name = %[2]q
  username     = %[3]q
  password     = %[4]q
}

data "aiven_clickhouse_user" "by_username" {
  project      = aiven_clickhouse_user.foo.project
  service_name = aiven_clickhouse_user.foo.service_name
  username     = aiven_clickhouse_user.foo.username

  depends_on = [aiven_clickhouse_user.foo]
}

data "aiven_clickhouse_user" "by_uuid" {
  project      = aiven_clickhouse_user.foo.project
  service_name = aiven_clickhouse_user.foo.service_name
  uuid         = aiven_clickhouse_user.foo.uuid

  depends_on = [aiven_clickhouse_user.foo]
}
`, project, serviceName, userName, password)
}

func testAccClickHouseUserWithoutPassword(project, serviceName, userName string) string {
	return fmt.Sprintf(`
resource "aiven_clickhouse_user" "foo" {
  project      = %[1]q
  service_name = %[2]q
  username     = %[3]q
}

data "aiven_clickhouse_user" "by_username" {
  project      = aiven_clickhouse_user.foo.project
  service_name = aiven_clickhouse_user.foo.service_name
  username     = aiven_clickhouse_user.foo.username

  depends_on = [aiven_clickhouse_user.foo]
}

data "aiven_clickhouse_user" "by_uuid" {
  project      = aiven_clickhouse_user.foo.project
  service_name = aiven_clickhouse_user.foo.service_name
  uuid         = aiven_clickhouse_user.foo.uuid

  depends_on = [aiven_clickhouse_user.foo]
}
`, project, serviceName, userName)
}

func testAccClickHouseUserWriteOnly(project, serviceName, userName, password string, version int) string {
	return fmt.Sprintf(`
resource "aiven_clickhouse_user" "foo" {
  project             = %[1]q
  service_name        = %[2]q
  username            = %[3]q
  password_wo         = %[4]q
  password_wo_version = %[5]d
}

data "aiven_clickhouse_user" "by_username" {
  project      = aiven_clickhouse_user.foo.project
  service_name = aiven_clickhouse_user.foo.service_name
  username     = aiven_clickhouse_user.foo.username

  depends_on = [aiven_clickhouse_user.foo]
}

data "aiven_clickhouse_user" "by_uuid" {
  project      = aiven_clickhouse_user.foo.project
  service_name = aiven_clickhouse_user.foo.service_name
  uuid         = aiven_clickhouse_user.foo.uuid

  depends_on = [aiven_clickhouse_user.foo]
}
`, project, serviceName, userName, password, version)
}

func testAccClickHouseUserWithoutPasswordResourceOnly(project, serviceName, userName string) string {
	return fmt.Sprintf(`
resource "aiven_clickhouse_user" "foo" {
  project      = %[1]q
  service_name = %[2]q
  username     = %[3]q
}
`, project, serviceName, userName)
}

func testAccClickHouseUserBackwardCompatibility(project, serviceName, userName string) string {
	return fmt.Sprintf(`
resource "aiven_clickhouse_user" "test" {
  project      = %[1]q
  service_name = %[2]q
  username     = %[3]q
}

data "aiven_clickhouse_user" "test" {
  project      = aiven_clickhouse_user.test.project
  service_name = aiven_clickhouse_user.test.service_name
  username     = aiven_clickhouse_user.test.username

  depends_on = [aiven_clickhouse_user.test]
}
`, project, serviceName, userName)
}
