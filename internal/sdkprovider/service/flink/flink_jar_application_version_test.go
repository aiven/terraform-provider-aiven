package flink_test

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// TestAccAivenFlinkJarApplicationVersion_basic
// This test requires a jar file to run.
func TestAccAivenFlinkJarApplicationVersion_basic(t *testing.T) {
	acc.SkipIfNotBeta(t)

	jarFile := os.Getenv("AIVEN_TEST_FLINK_JAR_FILE")
	if jarFile == "" {
		remove, tmpFile, err := createMinimalJar()
		require.NoError(t, err)
		jarFile = tmpFile
		defer remove()
	}

	project := os.Getenv("AIVEN_PROJECT_NAME")
	resourceNameApp := "aiven_flink_jar_application.app"
	resourceNameVersion := "aiven_flink_jar_application_version.version"
	resourceNameDeployment := "aiven_flink_jar_application_deployment.deployment"
	serviceName := fmt.Sprintf("test-acc-flink-%s", acc.RandStr())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenFlinkJarApplicationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlinkJarApplicationVersionResource(project, serviceName, jarFile),
				Check: resource.ComposeTestCheckFunc(
					// aiven_flink_jar_application
					resource.TestCheckResourceAttr(resourceNameApp, "project", project),
					resource.TestCheckResourceAttr(resourceNameApp, "service_name", serviceName),

					// aiven_flink_jar_application_version
					resource.TestCheckResourceAttr(resourceNameVersion, "project", project),
					resource.TestCheckResourceAttr(resourceNameVersion, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceNameVersion, "source", jarFile),
					resource.TestCheckResourceAttrSet(resourceNameVersion, "application_id"),
					resource.TestCheckResourceAttrSet(resourceNameVersion, "source_checksum"),
					resource.TestCheckResourceAttr(resourceNameVersion, "file_info.0.file_status", "READY"),

					// aiven_flink_jar_application_deployment
					resource.TestCheckResourceAttr(resourceNameDeployment, "project", project),
					resource.TestCheckResourceAttr(resourceNameDeployment, "service_name", serviceName),
					resource.TestCheckResourceAttrSet(resourceNameDeployment, "application_id"),
					resource.TestCheckResourceAttrSet(resourceNameDeployment, "version_id"),
				),
			},
		},
	})
}

func testAccCheckAivenFlinkJarApplicationVersionDestroy(s *terraform.State) error {
	client, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_flink_jar_application_version" {
			continue
		}

		project, serviceName, applicationID, version, err := schemautil.SplitResourceID4(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = client.ServiceFlinkGetJarApplicationVersion(ctx, project, serviceName, applicationID, version)
		if avngen.IsNotFound(err) {
			return nil
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("flink jar application version (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccFlinkJarApplicationVersionResource(project, serviceName, exampleJar string) string {
	return fmt.Sprintf(`
resource "aiven_flink" "service" {
  project                 = %[1]q
  service_name            = %[2]q
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "04:00:00"

  flink_user_config {
    custom_code = true
  }
}

resource "aiven_flink_jar_application" "app" {
  project      = %[1]q
  service_name = %[2]q
  name         = "my-app-jar"

  depends_on = [aiven_flink.service]
}

resource "aiven_flink_jar_application_version" "version" {
  project        = %[1]q
  service_name   = %[2]q
  application_id = aiven_flink_jar_application.app.application_id
  source         = %[3]q
}

resource "aiven_flink_jar_application_deployment" "deployment" {
  project        = %[1]q
  service_name   = %[2]q
  application_id = aiven_flink_jar_application_version.version.application_id
  version_id     = aiven_flink_jar_application_version.version.application_version_id
}
`, project, serviceName, exampleJar)
}

// createMinimalJar creates a JAR file.
// It doesn't work but should be okay for the test.
func createMinimalJar() (func(), string, error) {
	// Create the JAR file
	file, err := os.CreateTemp("", "temp-*.jar")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Create a new zip writer
	zw := zip.NewWriter(file)
	defer zw.Close()

	// Add META-INF directory
	dirHeader := &zip.FileHeader{
		Name:     "META-INF/",
		Method:   zip.Store, // entries should use STORE
		Modified: time.Now(),
	}
	dirHeader.SetMode(0o755)
	_, err = zw.CreateHeader(dirHeader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create META-INF directory: %w", err)
	}

	// Add MANIFEST.MF file
	manifestHeader := &zip.FileHeader{
		Name:     "META-INF/MANIFEST.MF",
		Method:   zip.Deflate, // use compression for files
		Modified: time.Now(),
	}
	manifestHeader.SetMode(0o644)

	mf, err := zw.CreateHeader(manifestHeader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create manifest file: %w", err)
	}

	// Write minimal manifest content
	manifest := "Manifest-Version: 1.0\r\nCreated-By: test-jar-go\r\n\r\n"
	_, err = io.WriteString(mf, manifest)
	if err != nil {
		return nil, "", fmt.Errorf("failed to write manifest content: %w", err)
	}

	return func() { os.Remove(file.Name()) }, file.Name(), nil
}
