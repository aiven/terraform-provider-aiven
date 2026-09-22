package custompluginfile_test

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

func TestAccAivenKafkaConnectCustomPluginFile(t *testing.T) {
	acc.SkipIfNotBeta(t)

	const resourceName = "aiven_kafka_connect_custom_plugin_file.foo"

	pluginName := acc.RandName("plugin")

	// Copies the test owns, so it can edit the file content.
	jarFile := copyFile(t, pluginJar(t), "plugin.jar")
	jarRenamed := copyFile(t, jarFile, "renamed.jar")

	config := testAccCustomPluginFile(pluginName, "1.0.0", "description", jarFile)
	configDescription := testAccCustomPluginFile(pluginName, "1.0.0", "updated description", jarFile)
	configRenamed := testAccCustomPluginFile(pluginName, "1.0.0", "updated description", jarRenamed)

	var fileID, checksum string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenKafkaConnectCustomPluginFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "plugin_name", pluginName),
					resource.TestCheckResourceAttr(resourceName, "plugin_version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "source", jarFile),
					// Defaults.
					resource.TestCheckResourceAttr(resourceName, "service_type", "kafka_connect"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "application/java-archive"),
					// The upload is verified by the time the create returns.
					resource.TestCheckResourceAttr(resourceName, "file_status", "READY"),
					resource.TestCheckResourceAttrPair(resourceName, "source_checksum", resourceName, "file_sha256"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "plugin_file_id"),
					resource.TestCheckResourceAttrSet(resourceName, "plugin_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					storeAttr(resourceName, "plugin_file_id", &fileID),
					storeAttr(resourceName, "source_checksum", &checksum),
				),
			},
			{
				// The file description is updated in place: the file stays.
				Config: configDescription,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "file_description", "updated description"),
					resource.TestCheckResourceAttrPtr(resourceName, "plugin_file_id", &fileID),
					resource.TestCheckResourceAttrPtr(resourceName, "source_checksum", &checksum),
				),
			},
			{
				// The same file under a new path uploads nothing, so the file record stays.
				Config: configRenamed,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "source", jarRenamed),
					resource.TestCheckResourceAttrPtr(resourceName, "plugin_file_id", &fileID),
					resource.TestCheckResourceAttrPtr(resourceName, "source_checksum", &checksum),
				),
			},
			{
				// The imported file must not be planned for replacement: the second step of the
				// test framework plans again after the import, and fails on a non-empty plan.
				Config:            configRenamed,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// The file path is local, so the API has nothing to import it from.
				ImportStateVerifyIgnore: []string{"source", "timeouts"},
			},
			{
				// Edited content at a known path can only be uploaded as a new file.
				PreConfig: func() { appendToFile(t, jarRenamed) },
				Config:    configRenamed,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "file_status", "READY"),
					resource.TestCheckResourceAttrPair(resourceName, "source_checksum", resourceName, "file_sha256"),
					checkAttrDiffers(resourceName, "plugin_file_id", &fileID),
					checkAttrDiffers(resourceName, "source_checksum", &checksum),
				),
			},
		},
	})
}

// TestAccAivenKafkaConnectCustomPluginFile_forceNew proves that a changed version replaces the file.
func TestAccAivenKafkaConnectCustomPluginFile_forceNew(t *testing.T) {
	acc.SkipIfNotBeta(t)

	const resourceName = "aiven_kafka_connect_custom_plugin_file.foo"

	pluginName := acc.RandName("plugin")
	jarFile := copyFile(t, pluginJar(t), "plugin.jar")

	var fileID string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenKafkaConnectCustomPluginFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginFile(pluginName, "1.0.0", "description", jarFile),
				Check:  storeAttr(resourceName, "plugin_file_id", &fileID),
			},
			{
				Config: testAccCustomPluginFile(pluginName, "1.0.1", "description", jarFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "plugin_version", "1.0.1"),
					checkAttrDiffers(resourceName, "plugin_file_id", &fileID),
				),
			},
		},
	})
}

func testAccCheckAivenKafkaConnectCustomPluginFileDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_kafka_connect_custom_plugin_file" {
			continue
		}

		organizationID, fileID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = c.KafkaConnectCustomPluginFileGet(ctx, organizationID, fileID)
		if avngen.IsNotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("kafka connect custom plugin file %s still exists", rs.Primary.ID)
	}

	return nil
}

// pluginJar returns a minimal Kafka Connect sink connector. Verification only reaches READY
// when the jar contains a discoverable connector class, so an empty jar won't do. The jar is
// built by testdata/build.sh, or the AIVEN_TEST_KAFKA_CONNECT_PLUGIN_FILE env var points to another.
func pluginJar(t *testing.T) string {
	t.Helper()

	path := os.Getenv("AIVEN_TEST_KAFKA_CONNECT_PLUGIN_FILE")
	if path == "" {
		path = filepath.Join("testdata", "noop-sink-connector.jar")
	}

	if _, err := os.Stat(path); err != nil {
		t.Skipf("plugin jar %q is missing, run testdata/build.sh: %v", path, err)
	}

	return path
}

// copyFile copies the file into the test's own directory under the given name.
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

func testAccCustomPluginFile(pluginName, version, fileDescription, source string) string {
	return fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = %[1]q
}

resource "aiven_kafka_connect_custom_plugin_file" "foo" {
  organization_id  = data.aiven_organization.foo.id
  plugin_name      = %[2]q
  plugin_version   = %[3]q
  file_description = %[4]q
  source           = %[5]q
}
`, acc.OrganizationName(), pluginName, version, fileDescription, source)
}
