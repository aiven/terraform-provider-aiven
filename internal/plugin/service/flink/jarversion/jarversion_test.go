package jarversion_test

import (
	"context"
	"fmt"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenFlinkJarApplicationVersion(t *testing.T) {
	acc.SkipIfNotBeta(t)

	const resourceName = "aiven_flink_jar_application_version.foo"

	projectName := acc.ProjectName()
	jarFile := acc.FlinkJarFile(t)

	serviceName := acc.RandName("flink")
	serviceIsReady := acc.CreateTestService(
		t,
		projectName,
		serviceName,
		acc.WithServiceType("flink"),
		acc.WithPlan("business-4"),
		acc.WithCloud("google-europe-west1"),
		// Jar applications are only available when the service accepts custom code.
		acc.WithUserConfig(map[string]any{"custom_code": true}),
	)

	// Prove that state written by the last published provider (which stored a top-level
	// source_checksum attribute) upgrades to the current schema (file_sha256 replaces it)
	// without a diff: step 2 of BackwardCompatibilitySteps replans with the current provider
	// and asserts the plan is empty.
	t.Run("backward compatibility test", func(t *testing.T) {
		appName := acc.RandName("compat")

		// The old provider computes source_checksum itself from the file bytes, so a stable path
		// is enough; nothing in the config references file_sha256.
		jarCopy := acc.CopyFile(t, jarFile, "app.jar")
		config := testAccFlinkJarApplicationVersion(projectName, serviceName, appName, jarCopy)

		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() { acc.TestAccPreCheck(t) },
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				PreConfig:          func() { require.NoError(t, <-serviceIsReady) },
				TFConfig:           config,
				OldProviderVersion: "4.63.0",
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "source", jarCopy),
					resource.TestCheckResourceAttrSet(resourceName, "application_version_id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			}),
		})
	})

	t.Run("base test", func(t *testing.T) {
		appName := acc.RandName("basic")

		// A copy the test owns, so it can rename and edit the jar file.
		jarCopy := acc.CopyFile(t, jarFile, "app.jar")
		jarRenamed := acc.CopyFile(t, jarFile, "renamed.jar")

		config := testAccFlinkJarApplicationVersion(projectName, serviceName, appName, jarCopy)
		configRenamed := testAccFlinkJarApplicationVersion(projectName, serviceName, appName, jarRenamed)
		configTrackedSource := testAccFlinkJarApplicationVersionTrackedSource(
			projectName, serviceName, appName, jarRenamed,
		)
		configUnknownSource := testAccFlinkJarApplicationVersionUnknownSource(
			projectName, serviceName, appName, jarRenamed,
		)

		var versionID, checksum string
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenFlinkJarApplicationVersionDestroy,
			Steps: []resource.TestStep{
				{
					PreConfig: func() { require.NoError(t, <-serviceIsReady) },
					Config:    config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "source", jarCopy),
						// The upload is complete by the time the create returns.
						resource.TestCheckResourceAttr(resourceName, "file_info.0.file_status", "READY"),
						// The top-level file_sha256 mirrors the nested one.
						resource.TestCheckResourceAttrPair(
							resourceName, "file_sha256",
							resourceName, "file_info.0.file_sha256",
						),
						resource.TestCheckResourceAttrSet(resourceName, "application_version_id"),
						resource.TestCheckResourceAttrSet(resourceName, "version"),
						resource.TestCheckResourceAttrSet(resourceName, "created_at"),
						resource.TestCheckResourceAttrSet(resourceName, "created_by"),
						acc.StoreAttr(resourceName, "application_version_id", &versionID),
						acc.StoreAttr(resourceName, "file_sha256", &checksum),
					),
				},
				{
					Config:            config,
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
					// The jar file path is local, so the API has nothing to import it from.
					ImportStateVerifyIgnore: []string{"source"},
				},
				{
					// The same jar under a new path uploads nothing, so the version stays.
					Config: configRenamed,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "source", jarRenamed),
						resource.TestCheckResourceAttrPtr(resourceName, "application_version_id", &versionID),
						resource.TestCheckResourceAttrPtr(resourceName, "file_sha256", &checksum),
					),
				},
				{
					// Edited content at a known path can only be uploaded to a new version.
					// terraform_data starts tracking the file here, but source remains a literal.
					PreConfig: func() { acc.AppendToFile(t, jarRenamed, "aiven") },
					Config:    configTrackedSource,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "source", jarRenamed),
						resource.TestCheckResourceAttr(resourceName, "file_info.0.file_status", "READY"),
						resource.TestCheckResourceAttrSet(resourceName, "file_sha256"),
						acc.CheckAttrDiffers(resourceName, "application_version_id", &versionID),
						acc.CheckAttrDiffers(resourceName, "file_sha256", &checksum),
						acc.StoreAttr(resourceName, "application_version_id", &versionID),
						acc.StoreAttr(resourceName, "file_sha256", &checksum),
					),
				},
				{
					// Editing the file now replaces terraform_data, so its output and therefore
					// source are unknown in the initial plan. Replacement must be planned before
					// the path resolves during apply.
					PreConfig: func() { acc.AppendToFile(t, jarRenamed, "aiven") },
					Config:    configUnknownSource,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "source", jarRenamed),
						resource.TestCheckResourceAttr(resourceName, "file_info.0.file_status", "READY"),
						resource.TestCheckResourceAttrSet(resourceName, "file_sha256"),
						acc.CheckAttrDiffers(resourceName, "application_version_id", &versionID),
						acc.CheckAttrDiffers(resourceName, "file_sha256", &checksum),
					),
				},
			},
		})
	})
}

func testAccCheckAivenFlinkJarApplicationVersionDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_flink_jar_application_version" {
			continue
		}

		projectName, serviceName, applicationID, versionID, err := schemautil.SplitResourceID4(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = c.ServiceFlinkGetJarApplicationVersion(ctx, projectName, serviceName, applicationID, versionID)
		if avngen.IsNotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("flink jar application version %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccFlinkJarApplicationVersion(projectName, serviceName, appName, jarFile string) string {
	return fmt.Sprintf(`
resource "aiven_flink_jar_application" "foo" {
  project      = %[1]q
  service_name = %[2]q
  name         = %[3]q
}

resource "aiven_flink_jar_application_version" "foo" {
  project        = %[1]q
  service_name   = %[2]q
  application_id = aiven_flink_jar_application.foo.application_id
  source         = %[4]q
}
`, projectName, serviceName, appName, jarFile)
}

func testAccFlinkJarApplicationVersionUnknownSource(projectName, serviceName, appName, jarFile string) string {
	return testAccFlinkJarApplicationVersionTerraformDataSource(
		projectName,
		serviceName,
		appName,
		jarFile,
		"terraform_data.jar_source.output",
	)
}

func testAccFlinkJarApplicationVersionTrackedSource(projectName, serviceName, appName, jarFile string) string {
	return testAccFlinkJarApplicationVersionTerraformDataSource(
		projectName,
		serviceName,
		appName,
		jarFile,
		fmt.Sprintf("%q", jarFile),
	)
}

func testAccFlinkJarApplicationVersionTerraformDataSource(
	projectName, serviceName, appName, jarFile, source string,
) string {
	return fmt.Sprintf(`
resource "terraform_data" "jar_source" {
  input            = %[4]q
  triggers_replace = filesha256(%[4]q)
}

resource "aiven_flink_jar_application" "foo" {
  project      = %[1]q
  service_name = %[2]q
  name         = %[3]q
}

resource "aiven_flink_jar_application_version" "foo" {
  project        = %[1]q
  service_name   = %[2]q
  application_id = aiven_flink_jar_application.foo.application_id
  source         = %[5]s
}
`, projectName, serviceName, appName, jarFile, source)
}
