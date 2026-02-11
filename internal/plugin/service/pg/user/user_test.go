package user_test

import (
	"context"
	"fmt"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

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
						schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_pg_user.user"),
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
					// import
					Config: fmt.Sprintf(`
resource "aiven_pg_user" "foo" {
  project      = %q
  service_name = %q
  username     = %q
}
`, projectName, serviceName, userName),
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
						schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_pg_user.user"),
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
					Config: testAccPGUserBulk(projectName, serviceName, "BulkPass$123"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "username", "user-1"),
						acc.TestAccPasswordHasCustomPassword(resourceName, "BulkPass$123"),
						schemautil.TestAccCheckAivenServiceUserAttributes("data.aiven_pg_user.user"),
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

	t.Run("backward compatibility", func(t *testing.T) {
		userName := acc.RandName("user")
		config := fmt.Sprintf(`
resource "aiven_pg_user" "test" {
  project      = %[1]q
  service_name = %[2]q
  username     = %[3]q
}

data "aiven_pg_user" "test" {
  project      = aiven_pg_user.test.project
  service_name = aiven_pg_user.test.service_name
  username     = aiven_pg_user.test.username
}`, projectName, serviceName, userName)

		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() { acc.TestAccPreCheck(t) },
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				TFConfig: config,
				PreConfig: func() {
					require.NoError(t, <-serviceIsReady)
				},
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aiven_pg_user.test", "id"),
					resource.TestCheckResourceAttr("aiven_pg_user.test", "username", userName),
					resource.TestCheckResourceAttrSet("aiven_pg_user.test", "password"),
					resource.TestCheckResourceAttr("aiven_pg_user.test", "type", "normal"),

					resource.TestCheckResourceAttr("data.aiven_pg_user.test", "username", userName),
					resource.TestCheckResourceAttrSet("data.aiven_pg_user.test", "password"),
				),
			}),
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

func testAccPGUserBulk(projectName, serviceName, password string) string {
	return fmt.Sprintf(`
resource "aiven_pg_user" "foo" {
  count        = 42
  project      = %[1]q
  service_name = %[2]q
  username     = "user-${count.index + 1}"
  password     = %[3]q
}

data "aiven_pg_user" "user" {
  project      = %[1]q
  service_name = %[2]q
  username     = aiven_pg_user.foo.0.username

  depends_on = [aiven_pg_user.foo]
}`, projectName, serviceName, password)
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
