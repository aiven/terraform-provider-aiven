package custompluginfile

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	sweep.AddTestSweepers("aiven_kafka_connect_custom_plugin_file", &resource.Sweeper{
		Name: "aiven_kafka_connect_custom_plugin_file",
		F: func(_ string) error {
			client, err := sweep.SharedGenClient()
			if err != nil {
				return err
			}

			organizations, err := client.AccountList(ctx)
			if common.IsCritical(err) {
				return fmt.Errorf("error retrieving a list of organizations: %w", err)
			}

			for _, o := range organizations {
				if !strings.HasPrefix(o.AccountName, sweep.DefaultPrefix) {
					continue
				}

				files, err := client.KafkaConnectCustomPluginFileList(ctx, o.OrganizationId)
				if common.IsCritical(err) {
					return fmt.Errorf("error retrieving a list of kafka connect custom plugin files for organization %s: %w", o.OrganizationId, err)
				}

				for _, f := range files {
					if !strings.HasPrefix(f.PluginName, sweep.DefaultPrefix) {
						continue
					}

					if err = client.KafkaConnectCustomPluginFileDelete(ctx, o.OrganizationId, f.PluginFileId); common.IsCritical(err) {
						return fmt.Errorf("error deleting kafka connect custom plugin file %s: %w", f.PluginFileId, err)
					}
				}
			}

			return nil
		},
		Dependencies: []string{"aiven_organization"},
	})
}
