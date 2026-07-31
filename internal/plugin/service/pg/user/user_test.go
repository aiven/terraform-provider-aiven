package user_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver for the out-of-band DB rotation test
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// TestAccAivenPGUser_basic tests PG user CRUD operations.
// Note: data source checks intentionally skip the password attribute.
// The data source reads from the API directly in the same test step as the resource creation/update
// may receive a stale empty value due to API eventual consistency.
func TestAccAivenPGUser_basic(t *testing.T) {
	projectName := acc.ProjectName()
	serviceName := acc.RandName("pg")
	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		serviceName,
		acc.WithServiceType("pg"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
	)

	client, err := acc.GetTestGenAivenClient()
	require.NoError(t, err)

	t.Run("create user without password", func(t *testing.T) {
		resourceName := "aiven_pg_user.foo"
		userName := acc.RandName("user")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenPGUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccPGUserWithoutPassword(projectName, serviceName, userName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "username", userName),
						resource.TestCheckResourceAttrSet(resourceName, "password"),
						acc.TestAccPasswordHasGeneratedPassword(resourceName),
					),
				},
			},
		})
	})

	t.Run("password transitions", func(t *testing.T) {
		resourceName := "aiven_pg_user.foo"
		userName := acc.RandName("user")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenPGUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					// create with custom password
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccPGUserWithPassword(projectName, serviceName, userName, "Test$1234"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "username", userName),
						acc.TestAccPasswordHasCustomPassword(resourceName, "Test$1234"),
						schemautil.TestAccCheckAivenServiceUserAttributes(resourceName),
						resource.TestCheckResourceAttr("data.aiven_pg_user.user", "project", projectName),
						resource.TestCheckResourceAttr("data.aiven_pg_user.user", "service_name", serviceName),
						resource.TestCheckResourceAttr("data.aiven_pg_user.user", "username", userName),
						resource.TestCheckResourceAttr("data.aiven_pg_user.user", "type", "normal"),
					),
				},
				{
					// transition to auto-generated password
					Config: testAccPGUserWithoutPassword(projectName, serviceName, userName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "username", userName),
						acc.TestAccPasswordHasGeneratedPassword(resourceName),
					),
				},
				{
					// transition to custom password
					Config: testAccPGUserWithPassword(projectName, serviceName, userName, "Custom$Pass456"),
					Check: resource.ComposeTestCheckFunc(
						acc.TestAccPasswordHasCustomPassword(resourceName, "Custom$Pass456"),
					),
				},
				{
					// transition to write-only password
					Config: testAccPGUserWriteOnly(projectName, serviceName, userName, 1),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "username", userName),
						acc.TestAccPasswordHasWOPassword(resourceName),
						func(state *terraform.State) error {
							rsp, err := client.ServiceUserGet(t.Context(), projectName, serviceName, userName)
							require.NoError(t, err)
							require.Equal(t, "WriteOnlyPass$1", rsp.Password)
							return nil
						},
					),
				},
				{
					// rotate write-only password
					Config: testAccPGUserWriteOnly(projectName, serviceName, userName, 2),
					Check: resource.ComposeTestCheckFunc(
						acc.TestAccPasswordHasWOPassword(resourceName),
						func(state *terraform.State) error {
							rsp, err := client.ServiceUserGet(t.Context(), projectName, serviceName, userName)
							require.NoError(t, err)
							require.Equal(t, "WriteOnlyPass$2", rsp.Password)
							return nil
						},
					),
				},
				{
					// back to auto-generated
					Config: testAccPGUserWithoutPassword(projectName, serviceName, userName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(resourceName, "password"),
						acc.TestAccPasswordHasGeneratedPassword(resourceName),
					),
				},
				{
					// back to custom password to stabilize state before import
					Config: testAccPGUserWithPassword(projectName, serviceName, userName, "Import$Pass789"),
					Check: resource.ComposeTestCheckFunc(
						acc.TestAccPasswordHasCustomPassword(resourceName, "Import$Pass789"),
					),
				},
				{
					// import with a stable custom password
					Config:            testAccPGUserWithPassword(projectName, serviceName, userName, "Import$Pass789"),
					ResourceName:      resourceName,
					ImportStateId:     util.ComposeID(projectName, serviceName, userName),
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("pg_allow_replication", func(t *testing.T) {
		resourceName := "aiven_pg_user.foo"
		userName := acc.RandName("user")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenPGUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					// create with replication enabled
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccPGUserReplication(projectName, serviceName, userName, "Test$1234", true),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "username", userName),
						resource.TestCheckResourceAttr(resourceName, "password", "Test$1234"),
						resource.TestCheckResourceAttr(resourceName, "pg_allow_replication", "true"),
						schemautil.TestAccCheckAivenServiceUserAttributes(resourceName),
						resource.TestCheckResourceAttr("data.aiven_pg_user.user", "project", projectName),
						resource.TestCheckResourceAttr("data.aiven_pg_user.user", "service_name", serviceName),
						resource.TestCheckResourceAttr("data.aiven_pg_user.user", "username", userName),
						resource.TestCheckResourceAttr("data.aiven_pg_user.user", "type", "normal"),
					),
				},
				{
					// disable replication
					Config: testAccPGUserReplication(projectName, serviceName, userName, "Test$1234", false),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "password", "Test$1234"),
						resource.TestCheckResourceAttr(resourceName, "pg_allow_replication", "false"),
					),
				},
				{
					// re-enable replication
					Config: testAccPGUserReplication(projectName, serviceName, userName, "Test$1234", true),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "password", "Test$1234"),
						resource.TestCheckResourceAttr(resourceName, "pg_allow_replication", "true"),
					),
				},
			},
		})
	})

	t.Run("remove missing", func(t *testing.T) {
		resourceName := "aiven_pg_user.foo"
		userName := acc.RandName("user")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenPGUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccPGUserWithoutPassword(projectName, serviceName, userName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s", projectName, serviceName, userName)),
					),
				},
				{
					// Delete user externally, verify plan detects missing
					Config: testAccPGUserWithoutPassword(projectName, serviceName, userName),
					PreConfig: func() {
						err := client.ServiceUserDelete(t.Context(), projectName, serviceName, userName)
						require.NoError(t, err)

						// Wait for deletion to propagate across API nodes
						require.Eventually(t, func() bool {
							_, err := client.ServiceUserGet(t.Context(), projectName, serviceName, userName)
							return avngen.IsNotFound(err)
						}, 30*time.Second, time.Second, "delete did not propagate")
					},
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
				},
				{
					// Recreate
					Config: testAccPGUserWithoutPassword(projectName, serviceName, userName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s", projectName, serviceName, userName)),
					),
				},
			},
		})
	})

	t.Run("bulk creation", func(t *testing.T) {
		resourceName := "aiven_pg_user.foo.0"
		rName := acc.RandStr()
		userName := fmt.Sprintf("user-%s-1", rName)

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenPGUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccPGUserBulk(projectName, serviceName, rName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "username", userName),
						resource.TestCheckResourceAttrSet(resourceName, "password"),
						schemautil.TestAccCheckAivenServiceUserAttributes(resourceName),
						resource.TestCheckResourceAttr("data.aiven_pg_user.user", "project", projectName),
						resource.TestCheckResourceAttr("data.aiven_pg_user.user", "service_name", serviceName),
						resource.TestCheckResourceAttr("data.aiven_pg_user.user", "username", userName),
						resource.TestCheckResourceAttr("data.aiven_pg_user.user", "type", "normal"),
					),
				},
			},
		})
	})

	t.Run("password in template interpolation", func(t *testing.T) {
		resourceName := "aiven_pg_user.foo"
		userName := acc.RandName("user")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenPGUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccPGUserTemplateInterpolation(projectName, serviceName, userName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "username", userName),
						resource.TestCheckResourceAttrSet(resourceName, "password"),
					),
				},
			},
		})
	})
}

// TestAccAivenPGUser_OutOfBandPasswordRotation pins the CURRENT behavior when a service
// user's password is rotated outside of Terraform.
// With the behavior changes, these subtests must change deliberately.
func TestAccAivenPGUser_OutOfBandPasswordRotation(t *testing.T) {
	t.Skip("Skipping out-of-band password rotation test")

	projectName := acc.ProjectName()
	serviceName := acc.RandName("pg")
	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		serviceName,
		acc.WithServiceType("pg"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
	)

	client, err := acc.GetTestGenAivenClient()
	require.NoError(t, err)

	// apiPassword fetches the user's password from the Aiven API
	apiPassword := func(userName string) (string, error) {
		rsp, err := client.ServiceUserGet(t.Context(), projectName, serviceName, userName)
		if err != nil {
			return "", err
		}
		return rsp.Password, nil
	}

	// rotatePassword resets the user's credentials to newPassword via the Aiven API,
	// and waits until the API reports the new password.
	rotatePassword := func(userName, newPassword string) {
		_, err := client.ServiceUserCredentialsModify(t.Context(), projectName, serviceName, userName,
			&service.ServiceUserCredentialsModifyIn{
				Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
				NewPassword: &newPassword,
			})
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			password, err := apiPassword(userName)
			return err == nil && password == newPassword
		}, 3*time.Minute, 2*time.Second, "credentials reset did not propagate")
	}

	// requireAPIPassword returns a check that waits until the API reports wantPassword.
	requireAPIPassword := func(userName, wantPassword, msg string) resource.TestCheckFunc {
		return func(*terraform.State) error {
			require.Eventually(t, func() bool {
				password, err := apiPassword(userName)
				return err == nil && password == wantPassword
			}, 3*time.Minute, 2*time.Second, msg)
			return nil
		}
	}

	// A plain (in-config) password is rotated via the Aiven API behind Terraform's back:
	// the plan stays empty, state keeps the old password, and only an
	// explicit password change in the config reconverges the API with state.
	t.Run("plain password rotated via api", func(t *testing.T) {
		resourceName := "aiven_pg_user.foo"
		userName := acc.RandName("user")

		const (
			originalPassword = "Original$Pass1"
			rotatedPassword  = "Rotated$Pass2"
			healedPassword   = "Healed$Pass3"
		)

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenPGUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccPGUserNoDatasource(projectName, serviceName, userName, originalPassword),
					Check:  acc.TestAccPasswordHasCustomPassword(resourceName, originalPassword),
				},
				{
					// Rotate the password via the API behind Terraform's back.
					// Pinned behavior: the plan stays empty because PasswordFlatten keeps
					// the state password over the rotated one the API now returns.
					PreConfig: func() {
						rotatePassword(userName, rotatedPassword)
					},
					Config:   testAccPGUserNoDatasource(projectName, serviceName, userName, originalPassword),
					PlanOnly: true,
				},
				{
					// State keeps the original password, the API holds the rotated one.
					Config: testAccPGUserNoDatasource(projectName, serviceName, userName, originalPassword),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "password", originalPassword),
						requireAPIPassword(userName, rotatedPassword, "the API must still hold the out-of-band password"),
					),
				},
				{
					// An explicit config change is the only path that resets credentials.
					Config: testAccPGUserNoDatasource(projectName, serviceName, userName, healedPassword),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "password", healedPassword),
						requireAPIPassword(userName, healedPassword, "password reset did not propagate to the API"),
					),
				},
			},
		})
	})

	// A write-only password is rotated via the Aiven API behind Terraform's back:
	// the plan stays empty because nothing about the password is stored in state
	// and only a password_wo_version bump re-applies the configured password.
	t.Run("write-only password rotated via api", func(t *testing.T) {
		resourceName := "aiven_pg_user.foo"
		userName := acc.RandName("user")

		// testAccPGUserWriteOnly derives the password from the version: WriteOnlyPass$<version>.
		const (
			woPassword1     = "WriteOnlyPass$1"
			rotatedPassword = "Rotated$Pass2"
			woPassword2     = "WriteOnlyPass$2"
		)

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenPGUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccPGUserWriteOnly(projectName, serviceName, userName, 1),
					Check: resource.ComposeTestCheckFunc(
						acc.TestAccPasswordHasWOPassword(resourceName),
						requireAPIPassword(userName, woPassword1, "write-only password was not applied"),
					),
				},
				{
					// Rotate the password via the API behind Terraform's back.
					// Unlike the plain-password subtests, this is the intended write-only
					// convention, not a side effect.
					PreConfig: func() {
						rotatePassword(userName, rotatedPassword)
					},
					Config:   testAccPGUserWriteOnly(projectName, serviceName, userName, 1),
					PlanOnly: true,
				},
				{
					// State still has no password and the unchanged version, while the API
					// holds the rotated one.
					Config: testAccPGUserWriteOnly(projectName, serviceName, userName, 1),
					Check: resource.ComposeTestCheckFunc(
						acc.TestAccPasswordHasWOPassword(resourceName),
						resource.TestCheckResourceAttr(resourceName, "password_wo_version", "1"),
						requireAPIPassword(userName, rotatedPassword, "the API must still hold the out-of-band password"),
					),
				},
				{
					// Bumping password_wo_version is the explicit "overwrite now" trigger:
					// it resets credentials and reconverges the API with the config.
					Config: testAccPGUserWriteOnly(projectName, serviceName, userName, 2),
					Check: resource.ComposeTestCheckFunc(
						acc.TestAccPasswordHasWOPassword(resourceName),
						resource.TestCheckResourceAttr(resourceName, "password_wo_version", "2"),
						requireAPIPassword(userName, woPassword2, "version bump did not re-apply the write-only password"),
					),
				},
			},
		})
	})

	// A plain (in-config) password is rotated directly in the database, the plan stays empty.
	t.Run("plain password rotated in the database", func(t *testing.T) {
		resourceName := "aiven_pg_user.foo"
		userName := acc.RandName("user")

		const (
			originalPassword = "Original$Pass1"
			rotatedPassword  = "RotatedInDb$2"
			healedPassword   = "Healed$Pass3"
		)

		// pgExec runs the query on the service's defaultdb, connecting as userName with the given password.
		pgExec := func(password, query string) error {
			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()

			s, err := client.ServiceGet(ctx, projectName, serviceName)
			if err != nil {
				return err
			}

			u, err := url.Parse(s.ServiceUri)
			if err != nil {
				return err
			}
			u.User = url.UserPassword(userName, password)

			db, err := sql.Open("pgx", u.String())
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()

			_, err = db.ExecContext(ctx, query)
			return err
		}

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenPGUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccPGUserNoDatasource(projectName, serviceName, userName, originalPassword),
					Check:  acc.TestAccPasswordHasCustomPassword(resourceName, originalPassword),
				},
				{
					// Rotate the password directly in the database, behind both Terraform
					// and the Aiven API. Pinned behavior: the plan stays empty — the API keeps
					// reporting the password it set (or blanks it after noticing the hash
					// mismatch), and PasswordFlatten keeps the state password either way.
					PreConfig: func() {
						// Waits until the original credentials work against the database.
						require.Eventually(t, func() bool {
							return pgExec(originalPassword, "SELECT 1") == nil
						}, 3*time.Minute, 5*time.Second, "created credentials never became usable in the database")

						// A role can always change its own password, no admin privileges needed.
						require.NoError(t, pgExec(originalPassword,
							fmt.Sprintf("ALTER ROLE CURRENT_USER WITH PASSWORD '%s'", rotatedPassword)))

						require.NoError(t, pgExec(rotatedPassword, "SELECT 1"), "rotated password must work in the database")
						require.Error(t, pgExec(originalPassword, "SELECT 1"), "original password must no longer work in the database")
					},
					Config:   testAccPGUserNoDatasource(projectName, serviceName, userName, originalPassword),
					PlanOnly: true,
				},
				{
					// State keeps the original password even though the API never learns the
					// DB-side one: it returns the password it originally set until it notices
					// the hash mismatch in the database, and an empty password after that.
					// Either way, state is now actively wrong about the database.
					Config: testAccPGUserNoDatasource(projectName, serviceName, userName, originalPassword),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "password", originalPassword),
						func(*terraform.State) error {
							password, err := apiPassword(userName)
							require.NoError(t, err)
							require.Contains(t, []string{originalPassword, ""}, password,
								"the API must never learn the DB-side password")
							return nil
						},
					),
				},
				{
					// Changing the password in the config resets credentials and overwrites
					// the DB-side password: database, API, and state reconverge.
					Config: testAccPGUserNoDatasource(projectName, serviceName, userName, healedPassword),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "password", healedPassword),
						func(*terraform.State) error {
							require.Eventually(t, func() bool {
								return pgExec(healedPassword, "SELECT 1") == nil
							}, 3*time.Minute, 5*time.Second, "password reset did not propagate to the database")

							require.Error(t, pgExec(rotatedPassword, "SELECT 1"), "DB-side password must have been overwritten")
							return nil
						},
					),
				},
			},
		})
	})
}

func testAccCheckAivenPGUserResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("error instantiating client: %w", err)
	}

	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_pg_user" {
			continue
		}

		projectName, serviceName, username, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = c.ServiceUserGet(ctx, projectName, serviceName, username)
		if err != nil && !avngen.IsNotFound(err) {
			return fmt.Errorf("error checking if user was destroyed: %w", err)
		}

		if err == nil {
			return fmt.Errorf("pg user (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccPGUserWithoutPassword(projectName, serviceName, userName string) string {
	return fmt.Sprintf(`
resource "aiven_pg_user" "foo" {
  project      = %[1]q
  service_name = %[2]q
  username     = %[3]q
}
`, projectName, serviceName, userName)
}

func testAccPGUserWithPassword(projectName, serviceName, userName, password string) string {
	return fmt.Sprintf(`
resource "aiven_pg_user" "foo" {
  project      = %[1]q
  service_name = %[2]q
  username     = %[3]q
  password     = %[4]q
}

data "aiven_pg_user" "user" {
  service_name = aiven_pg_user.foo.service_name
  project      = aiven_pg_user.foo.project
  username     = aiven_pg_user.foo.username
}`, projectName, serviceName, userName, password)
}

func testAccPGUserNoDatasource(projectName, serviceName, userName, password string) string {
	return fmt.Sprintf(`
resource "aiven_pg_user" "foo" {
  project      = %[1]q
  service_name = %[2]q
  username     = %[3]q
  password     = %[4]q
}
`, projectName, serviceName, userName, password)
}

func testAccPGUserWriteOnly(projectName, serviceName, userName string, version int) string {
	return fmt.Sprintf(`
resource "aiven_pg_user" "foo" {
  project             = %[1]q
  service_name        = %[2]q
  username            = %[3]q
  password_wo         = "WriteOnlyPass$%[4]d"
  password_wo_version = %[4]d
}
`, projectName, serviceName, userName, version)
}

func testAccPGUserReplication(projectName, serviceName, userName, password string, allowReplication bool) string {
	return fmt.Sprintf(`
resource "aiven_pg_user" "foo" {
  project              = %[1]q
  service_name         = %[2]q
  username             = %[3]q
  password             = %[4]q
  pg_allow_replication = %[5]t
}

data "aiven_pg_user" "user" {
  service_name = aiven_pg_user.foo.service_name
  project      = aiven_pg_user.foo.project
  username     = aiven_pg_user.foo.username
}`, projectName, serviceName, userName, password, allowReplication)
}

func testAccPGUserBulk(projectName, serviceName, rName string) string {
	return fmt.Sprintf(`
resource "aiven_pg_user" "foo" {
  count        = 42
  project      = %[1]q
  service_name = %[2]q
  username     = "user-%[3]s-${count.index + 1}"
}

data "aiven_pg_user" "user" {
  project      = %[1]q
  service_name = %[2]q
  username     = aiven_pg_user.foo.0.username

  depends_on = [aiven_pg_user.foo]
}`, projectName, serviceName, rName)
}

func testAccPGUserTemplateInterpolation(projectName, serviceName, userName string) string {
	return fmt.Sprintf(`
resource "aiven_pg_user" "foo" {
  project      = %[1]q
  service_name = %[2]q
  username     = %[3]q
}

output "use-template-interpolation" {
  sensitive = true
  value     = "${aiven_pg_user.foo.password}/testing"
}
`, projectName, serviceName, userName)
}
