package custompluginfile_test

import (
	"context"
	"fmt"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenKafkaConnectCustomPluginFile(t *testing.T) {
	acc.SkipIfNotBeta(t)

	orgID := acc.OrganizationID()

	const resourceName = "aiven_kafka_connect_custom_plugin_file.foo"

	t.Run("base", func(t *testing.T) {
		pluginName := acc.RandName("plugin")

		// Test-owned copies so the test can edit them.
		jarFile := acc.CopyFile(t, acc.KafkaConnectPluginJar(t), "plugin.jar")
		jarRenamed := acc.CopyFile(t, jarFile, "renamed.jar")

		config := testAccCustomPluginFile(orgID, pluginName, "1.0.0", "description", jarFile)
		configDescription := testAccCustomPluginFile(orgID, pluginName, "1.0.0", "updated description", jarFile)
		configRenamed := testAccCustomPluginFile(orgID, pluginName, "1.0.0", "updated description", jarRenamed)

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
						resource.TestCheckResourceAttr(resourceName, "service_type", "kafka_connect"),
						resource.TestCheckResourceAttr(resourceName, "content_type", "application/java-archive"),
						// Upload is verified by the time create returns.
						resource.TestCheckResourceAttr(resourceName, "file_status", "READY"),
						resource.TestCheckResourceAttrSet(resourceName, "file_sha256"),
						resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
						resource.TestCheckResourceAttrSet(resourceName, "plugin_file_id"),
						resource.TestCheckResourceAttrSet(resourceName, "plugin_id"),
						resource.TestCheckResourceAttrSet(resourceName, "created_at"),
						resource.TestCheckResourceAttrSet(resourceName, "created_by"),
						acc.StoreAttr(resourceName, "plugin_file_id", &fileID),
						acc.StoreAttr(resourceName, "file_sha256", &checksum),
					),
				},
				{
					// The file description is updated in place: the file stays.
					Config: configDescription,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "file_description", "updated description"),
						resource.TestCheckResourceAttrPtr(resourceName, "plugin_file_id", &fileID),
						resource.TestCheckResourceAttrPtr(resourceName, "file_sha256", &checksum),
					),
				},
				{
					// The same file under a new path uploads nothing, so the file record stays.
					Config: configRenamed,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "source", jarRenamed),
						resource.TestCheckResourceAttrPtr(resourceName, "plugin_file_id", &fileID),
						resource.TestCheckResourceAttrPtr(resourceName, "file_sha256", &checksum),
					),
				},
				{
					// Import must not plan for replacement; the framework replans and fails on a non-empty plan.
					Config:            configRenamed,
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
					// source is local; the API can't import it.
					ImportStateVerifyIgnore: []string{"source", "timeouts", "service_type", "content_type"},
				},
				{
					// Edited content at the same path uploads as a new file.
					PreConfig: func() { acc.AppendToFile(t, jarRenamed, "aiven") },
					Config:    configRenamed,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "file_status", "READY"),
						acc.CheckAttrDiffers(resourceName, "plugin_file_id", &fileID),
						acc.CheckAttrDiffers(resourceName, "file_sha256", &checksum),
					),
				},
			},
		})
	})

	t.Run("forceNew", func(t *testing.T) {
		pluginName := acc.RandName("plugin")
		jarFile := acc.CopyFile(t, acc.KafkaConnectPluginJar(t), "plugin.jar")

		var fileID string
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenKafkaConnectCustomPluginFileDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccCustomPluginFile(orgID, pluginName, "1.0.0", "description", jarFile),
					Check:  acc.StoreAttr(resourceName, "plugin_file_id", &fileID),
				},
				{
					Config: testAccCustomPluginFile(orgID, pluginName, "1.0.1", "description", jarFile),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "plugin_version", "1.0.1"),
						acc.CheckAttrDiffers(resourceName, "plugin_file_id", &fileID),
					),
				},
			},
		})
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

func testAccCustomPluginFile(orgID, pluginName, version, fileDescription, source string) string {
	return fmt.Sprintf(`
resource "aiven_kafka_connect_custom_plugin_file" "foo" {
  organization_id  = %[1]q
  plugin_name      = %[2]q
  plugin_version   = %[3]q
  service_type     = "kafka_connect"
  content_type     = "application/java-archive"
  file_description = %[4]q
  source           = %[5]q
}
`, orgID, pluginName, version, fileDescription, source)
}
