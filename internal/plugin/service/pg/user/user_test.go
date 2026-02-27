package user_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenPGUserResourceDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						require.NoError(t, <-serviceIsReady)
					},
					Config: testAccPGUserBulk(projectName, serviceName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "username", "user-1"),
						resource.TestCheckResourceAttrSet(resourceName, "password"),
						schemautil.TestAccCheckAivenServiceUserAttributes(resourceName),
						resource.TestCheckResourceAttr("data.aiven_pg_user.user", "project", projectName),
						resource.TestCheckResourceAttr("data.aiven_pg_user.user", "service_name", serviceName),
						resource.TestCheckResourceAttr("data.aiven_pg_user.user", "username", "user-1"),
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

func testAccPGUserBulk(projectName, serviceName string) string {
	return fmt.Sprintf(`
resource "aiven_pg_user" "foo" {
  count        = 42
  project      = %[1]q
  service_name = %[2]q
  username     = "user-${count.index + 1}"
}

data "aiven_pg_user" "user" {
  project      = %[1]q
  service_name = %[2]q
  username     = aiven_pg_user.foo.0.username

  depends_on = [aiven_pg_user.foo]
}`, projectName, serviceName)
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
