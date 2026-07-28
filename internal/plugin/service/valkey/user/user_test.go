package user_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenValkeyUser(t *testing.T) {
	projectName := acc.ProjectName()
	serviceName := acc.RandName("valkey")

	// All subtests share a single Valkey service to avoid provisioning one per case.
	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		serviceName,
		acc.WithServiceType("valkey"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
	)

	client, err := acc.GetTestGenAivenClient()
	require.NoError(t, err)

	t.Run("basic", func(t *testing.T) {
		resourceName := "aiven_valkey_user.foo"
		userName := acc.RandName("user")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenValkeyUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					// create with optional password and ACLs
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccValkeyUserResource(projectName, serviceName, userName, "acc-custom-Test$1234"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "username", userName),

						// Password - using optional password
						resource.TestCheckResourceAttr(resourceName, "password", "acc-custom-Test$1234"),
						resource.TestCheckNoResourceAttr(resourceName, "password_wo"),
						resource.TestCheckNoResourceAttr(resourceName, "password_wo_version"),

						resource.TestCheckResourceAttrSet(resourceName, "type"),

						// ACLs
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_commands.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_commands.0", "+set"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_keys.#", "2"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_categories.#", "2"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_channels.#", "1"),

						schemautil.TestAccCheckAivenServiceUserAttributes(resourceName),

						// Data source: check non-password attributes only.
						resource.TestCheckResourceAttr("data.aiven_valkey_user.user", "username", userName),
						resource.TestCheckResourceAttr("data.aiven_valkey_user.user", "project", projectName),
						resource.TestCheckResourceAttr("data.aiven_valkey_user.user", "service_name", serviceName),
						resource.TestCheckResourceAttr("data.aiven_valkey_user.user", "valkey_acl_commands.#", "1"),
					),
				},
				{
					// update optional password
					Config: testAccValkeyUserResource(projectName, serviceName, userName, "acc-custom-UpdatedP@ss5678"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "username", userName),
						resource.TestCheckResourceAttr(resourceName, "password", "acc-custom-UpdatedP@ss5678"),
						resource.TestCheckNoResourceAttr(resourceName, "password_wo"),
						schemautil.TestAccCheckAivenServiceUserAttributes(resourceName),
					),
				},
				{
					// update ACLs in place (no recreation)
					Config: testAccValkeyUserUpdatedACL(projectName, serviceName, userName, "acc-custom-UpdatedP@ss5678"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "username", userName),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_commands.#", "2"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_commands.0", "+set"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_commands.1", "+get"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_keys.#", "3"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_keys.2", "new_key*"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_categories.#", "3"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_categories.2", "+@read"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_channels.#", "2"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_channels.1", "notifications"),
					),
				},
				{
					// import an existing user (composite ID project/service_name/username)
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					// transition to auto-generated password
					Config: testAccValkeyUserAutoGenerated(projectName, serviceName, userName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "username", userName),
						resource.TestCheckResourceAttrSet(resourceName, "password"),
						resource.TestCheckNoResourceAttr(resourceName, "password_wo"),
					),
				},
				{
					// transition to write-only password
					Config: testAccValkeyUserWriteOnly(projectName, serviceName, userName, 1),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "username", userName),

						// Password - using write-only password (password cleared, version set)
						resource.TestCheckNoResourceAttr(resourceName, "password"),
						resource.TestCheckResourceAttr(resourceName, "password_wo_version", "1"),

						// Checks if password_wo was actually set in Aiven.
						checkBackendPassword(t, client, projectName, serviceName, userName,
							func(p string) bool { return p == "WriteOnlyPass$1" },
							"expected password_wo to be set in Aiven"),
					),
				},
				{
					// rotate write-only password
					Config: testAccValkeyUserWriteOnly(projectName, serviceName, userName, 2),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "username", userName),

						// Password - rotated write-only password (password cleared, version incremented)
						resource.TestCheckNoResourceAttr(resourceName, "password"),
						resource.TestCheckResourceAttr(resourceName, "password_wo_version", "2"),

						// Checks if password_wo was actually updated in Aiven.
						checkBackendPassword(t, client, projectName, serviceName, userName,
							func(p string) bool { return p == "WriteOnlyPass$2" },
							"expected password_wo to be updated in Aiven"),
					),
				},
				{
					// transition back to auto-generated password
					Config: testAccValkeyUserAutoGenerated(projectName, serviceName, userName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "username", userName),
						resource.TestCheckResourceAttrSet(resourceName, "password"),
						resource.TestCheckNoResourceAttr(resourceName, "password_wo"),

						// Checks if password_wo was actually removed in Aiven and regenerated.
						checkBackendPassword(t, client, projectName, serviceName, userName,
							func(p string) bool { return p != "WriteOnlyPass$2" && p != "" },
							"expected password to be regenerated by Aiven"),
					),
				},
				{
					// This resource has RemoveMissing set to true, so it should be recreated if it's missing (deleted).
					Config: testAccValkeyUserAutoGenerated(projectName, serviceName, userName),
					PreConfig: func() {
						err := client.ServiceUserDelete(t.Context(), projectName, serviceName, userName)
						require.NoError(t, err)
					},
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
					Check: resource.ComposeTestCheckFunc(
						// Doesn't have ID as it will be recreated
						resource.TestCheckNoResourceAttr(resourceName, "id"),
					),
				},
				{
					// Resource is recreated after being applied
					Config: testAccValkeyUserAutoGenerated(projectName, serviceName, userName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s", projectName, serviceName, userName)),
						func(*terraform.State) error {
							user, err := client.ServiceUserGet(t.Context(), projectName, serviceName, userName)
							require.NoError(t, err, "retrieving valkey user after recreation")
							require.Equal(t, userName, user.Username)
							return nil
						},
					),
				},
			},
		})
	})

	// Proves ACL removal works: emptied ACLs can be dropped from the configuration without
	// drift, while an ACL that goes missing in Aiven is still reported as drift.
	t.Run("acl_removal", func(t *testing.T) {
		resourceName := "aiven_valkey_user.foo"
		userName := acc.RandName("user")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenValkeyUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					// create with ACLs
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccValkeyUserACL(projectName, serviceName, userName, valkeyUserBaseACL),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_commands.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_keys.#", "2"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_categories.#", "2"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_channels.#", "1"),
					),
				},
				{
					// valkey_acl_keys is drained in Aiven, so the plan drifts on that key alone
					PreConfig: func() {
						_, err := client.ServiceUserCredentialsModify(t.Context(), projectName, serviceName, userName,
							&service.ServiceUserCredentialsModifyIn{
								Operation: service.ServiceUserCredentialsModifyOperationTypeSetAccessControl,
								AccessControl: &service.AccessControlIn{
									// The other ACLs are resent as is, so only valkey_acl_keys can drift.
									ValkeyAclCategories: &[]string{"-@all", "+@admin"},
									ValkeyAclChannels:   &[]string{"test"},
									ValkeyAclCommands:   &[]string{"+set"},
									ValkeyAclKeys:       &[]string{},
								},
							})
						require.NoError(t, err, "draining valkey_acl_keys")
					},
					Config: testAccValkeyUserACL(projectName, serviceName, userName, valkeyUserBaseACL),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
							acc.ExpectOnlyAttributesChanged(resourceName, "valkey_acl_keys"),
						},
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_keys.#", "2"),
					),
				},
				{
					// empty lists clear the ACLs and are kept as configured, not turned into null
					Config: testAccValkeyUserACL(projectName, serviceName, userName, valkeyUserEmptyACL),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_commands.#", "0"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_keys.#", "0"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_categories.#", "0"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_channels.#", "0"),
					),
				},
				{
					// the emptied ACLs are stable
					Config:             testAccValkeyUserACL(projectName, serviceName, userName, valkeyUserEmptyACL),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
				{
					// dropping the ACLs from the configuration removes them from the state
					Config: testAccValkeyUserACL(projectName, serviceName, userName, ""),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckNoResourceAttr(resourceName, "valkey_acl_commands.#"),
						resource.TestCheckNoResourceAttr(resourceName, "valkey_acl_keys.#"),
						resource.TestCheckNoResourceAttr(resourceName, "valkey_acl_categories.#"),
						resource.TestCheckNoResourceAttr(resourceName, "valkey_acl_channels.#"),
					),
				},
				{
					// the removed ACLs don't come back as empty lists
					Config:             testAccValkeyUserACL(projectName, serviceName, userName, ""),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// Proves ACLs created as empty lists keep the configured shape: Terraform tells an empty
	// list apart from a null one, so "[]" stays "[]" in the state and doesn't drift after.
	t.Run("acl_empty_on_create", func(t *testing.T) {
		resourceName := "aiven_valkey_user.foo"
		userName := acc.RandName("user")
		config := fmt.Sprintf(`
resource "aiven_valkey_user" "foo" {
  project      = %[1]q
  service_name = %[2]q
  username     = %[3]q

  valkey_acl_commands   = []
  valkey_acl_keys       = []
  valkey_acl_categories = []
  valkey_acl_channels   = []
}`, projectName, serviceName, userName)

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenValkeyUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s", projectName, serviceName, userName)),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_commands.#", "0"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_keys.#", "0"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_categories.#", "0"),
						resource.TestCheckResourceAttr(resourceName, "valkey_acl_channels.#", "0"),
					),
				},
				{
					// the empty ACLs don't drift into null on refresh
					Config:             config,
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// Verifies that state created by the previous SDK-based provider version is
	// compatible with the Plugin Framework version.
	t.Run("backward_compat", func(t *testing.T) {
		userName := acc.RandName("user")
		config := fmt.Sprintf(`
resource "aiven_valkey_user" "foo" {
  project      = %[1]q
  service_name = %[2]q
  username     = %[3]q

  valkey_acl_commands   = ["+set"]
  valkey_acl_keys       = ["prefix*", "another_key"]
  valkey_acl_categories = ["-@all", "+@admin"]
  valkey_acl_channels   = ["test"]
}

data "aiven_valkey_user" "user" {
  project      = aiven_valkey_user.foo.project
  service_name = aiven_valkey_user.foo.service_name
  username     = aiven_valkey_user.foo.username
}`, projectName, serviceName, userName)

		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() { acc.TestAccPreCheck(t) },
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				TFConfig: config,
				PreConfig: func() {
					require.NoError(t, <-serviceIsReady)
				},
				OldProviderVersion: "4.47.0",
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aiven_valkey_user.foo", "id"),
					resource.TestCheckResourceAttr("aiven_valkey_user.foo", "username", userName),
					resource.TestCheckResourceAttrSet("aiven_valkey_user.foo", "password"),
					resource.TestCheckResourceAttrSet("aiven_valkey_user.foo", "type"),
					resource.TestCheckResourceAttr("aiven_valkey_user.foo", "valkey_acl_commands.#", "1"),
					resource.TestCheckResourceAttr("aiven_valkey_user.foo", "valkey_acl_keys.#", "2"),

					resource.TestCheckResourceAttr("data.aiven_valkey_user.user", "username", userName),
					resource.TestCheckResourceAttrSet("data.aiven_valkey_user.user", "password"),
				),
			}),
		})
	})
}

// checkBackendPassword returns a TestCheckFunc that polls the Aiven backend until the
// stored service user password satisfies pred.
// The backend is eventually consistent after a credentials reset.
func checkBackendPassword(
	t *testing.T,
	client avngen.Client,
	projectName, serviceName, userName string,
	pred func(password string) bool,
	msg string,
) resource.TestCheckFunc {
	return func(*terraform.State) error {
		require.Eventually(t, func() bool {
			rsp, err := client.ServiceUserGet(t.Context(), projectName, serviceName, userName)
			return err == nil && pred(rsp.Password)
		}, 2*time.Minute, time.Second, msg)
		return nil
	}
}

func testAccCheckAivenValkeyUserResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("error instantiating client: %w", err)
	}

	ctx := context.Background()

	// loop through the resources in state, verifying each aiven_valkey_user is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_valkey_user" {
			continue
		}

		projectName, serviceName, username, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = c.ServiceUserGet(ctx, projectName, serviceName, username)
		if err == nil {
			return fmt.Errorf("valkey user (%s) still exists", rs.Primary.ID)
		}

		if !avngen.IsNotFound(err) {
			return fmt.Errorf("error checking if user was destroyed: %w", err)
		}
	}

	return nil
}

const valkeyUserBaseACL = `
  valkey_acl_commands   = ["+set"]
  valkey_acl_keys       = ["prefix*", "another_key"]
  valkey_acl_categories = ["-@all", "+@admin"]
  valkey_acl_channels   = ["test"]`

const valkeyUserEmptyACL = `
  valkey_acl_commands   = []
  valkey_acl_keys       = []
  valkey_acl_categories = []
  valkey_acl_channels   = []`

// testAccValkeyUserACL renders a user with the given ACL block and no password,
// so the ACLs are the only thing under test.
func testAccValkeyUserACL(projectName, serviceName, userName, acl string) string {
	return fmt.Sprintf(`
resource "aiven_valkey_user" "foo" {
  project      = %[1]q
  service_name = %[2]q
  username     = %[3]q
%[4]s
}`, projectName, serviceName, userName, acl)
}

func testAccValkeyUserResource(projectName, serviceName, userName, password string) string {
	return fmt.Sprintf(`
resource "aiven_valkey_user" "foo" {
  project      = %[1]q
  service_name = %[2]q
  username     = %[3]q
  password     = %[4]q
%[5]s
}

data "aiven_valkey_user" "user" {
  service_name = aiven_valkey_user.foo.service_name
  project      = aiven_valkey_user.foo.project
  username     = aiven_valkey_user.foo.username
}`, projectName, serviceName, userName, password, valkeyUserBaseACL)
}

func testAccValkeyUserAutoGenerated(projectName, serviceName, userName string) string {
	return testAccValkeyUserACL(projectName, serviceName, userName, valkeyUserBaseACL)
}

func testAccValkeyUserWriteOnly(projectName, serviceName, userName string, version int) string {
	return fmt.Sprintf(`
resource "aiven_valkey_user" "foo" {
  project             = %[1]q
  service_name        = %[2]q
  username            = %[3]q
  password_wo         = "WriteOnlyPass$%[4]d"
  password_wo_version = %[4]d
%[5]s
}`, projectName, serviceName, userName, version, valkeyUserBaseACL)
}

func testAccValkeyUserUpdatedACL(projectName, serviceName, userName, password string) string {
	return fmt.Sprintf(`
resource "aiven_valkey_user" "foo" {
  project      = %[1]q
  service_name = %[2]q
  username     = %[3]q
  password     = %[4]q

  valkey_acl_commands   = ["+set", "+get"]
  valkey_acl_keys       = ["prefix*", "another_key", "new_key*"]
  valkey_acl_categories = ["-@all", "+@admin", "+@read"]
  valkey_acl_channels   = ["test", "notifications"]
}

data "aiven_valkey_user" "user" {
  service_name = aiven_valkey_user.foo.service_name
  project      = aiven_valkey_user.foo.project
  username     = aiven_valkey_user.foo.username
}`, projectName, serviceName, userName, password)
}
