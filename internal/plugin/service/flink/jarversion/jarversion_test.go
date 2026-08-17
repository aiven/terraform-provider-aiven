package jarversion_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

	// The new provider must read the SDKv2 state as is: the second step's empty plan proves it.
	t.Run("backward compatibility test", func(t *testing.T) {
		config := testAccFlinkJarApplicationVersion(
			projectName, serviceName, acc.RandName("compat"), jarFile,
		)
		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() { acc.TestAccPreCheck(t) },
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				PreConfig:          func() { require.NoError(t, <-serviceIsReady) },
				TFConfig:           config,
				OldProviderVersion: "4.61.0",
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "source", jarFile),
					resource.TestCheckResourceAttr(resourceName, "file_info.0.file_status", "READY"),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
					resource.TestCheckResourceAttrSet(resourceName, "application_version_id"),
					resource.TestCheckResourceAttrSet(resourceName, "source_checksum"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
				),
			}),
		})
	})

	t.Run("base test", func(t *testing.T) {
		appName := acc.RandName("basic")

		// A copy the test owns, so it can rename and edit the jar file.
		jarCopy := copyFile(t, jarFile, "app.jar")
		jarRenamed := copyFile(t, jarFile, "renamed.jar")

		config := testAccFlinkJarApplicationVersion(projectName, serviceName, appName, jarCopy)
		configRenamed := testAccFlinkJarApplicationVersion(projectName, serviceName, appName, jarRenamed)

		var versionID string
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
						resource.TestCheckResourceAttrPair(
							resourceName, "source_checksum",
							resourceName, "file_info.0.file_sha256",
						),
						resource.TestCheckResourceAttrSet(resourceName, "application_version_id"),
						resource.TestCheckResourceAttrSet(resourceName, "version"),
						resource.TestCheckResourceAttrSet(resourceName, "created_at"),
						resource.TestCheckResourceAttrSet(resourceName, "created_by"),
						storeAttr(resourceName, "application_version_id", &versionID),
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
					),
				},
				{
					// Edited content can only be uploaded to a new version.
					PreConfig: func() { appendToFile(t, jarRenamed) },
					Config:    configRenamed,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "file_info.0.file_status", "READY"),
						checkAttrDiffers(resourceName, "application_version_id", &versionID),
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

// copyFile copies the jar file into the test's own directory under the given name.
func copyFile(t *testing.T, source, name string) string {
	t.Helper()

	b, err := os.ReadFile(source) //nolint:gosec // The path comes from the test setup.
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, b, 0o600))
	return path
}

// appendToFile changes the file content, and with it its checksum.
func appendToFile(t *testing.T, path string) {
	t.Helper()

	file, err := os.OpenFile(filepath.Clean(path), os.O_APPEND|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	defer file.Close()

	_, err = file.WriteString("aiven")
	require.NoError(t, err)
}

func storeAttr(resourceName, key string, target *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		value, err := attrValue(s, resourceName, key)
		if err != nil {
			return err
		}

		*target = value
		return nil
	}
}

func checkAttrDiffers(resourceName, key string, previous *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		value, err := attrValue(s, resourceName, key)
		if err != nil {
			return err
		}

		if value == *previous {
			return fmt.Errorf("expected %s.%s to change, got %q", resourceName, key, value)
		}

		return nil
	}
}

func attrValue(s *terraform.State, resourceName, key string) (string, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return "", fmt.Errorf("resource %q not found in state", resourceName)
	}

	value := rs.Primary.Attributes[key]
	if value == "" {
		return "", fmt.Errorf("attribute %q of %q is empty", key, resourceName)
	}

	return value, nil
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
