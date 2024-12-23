package kafka

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
)

func aivenKafkaConnectSchema() map[string]*schema.Schema {
	return schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeKafkaConnect)
}

func ResourceKafkaConnect() *schema.Resource {
	return &schema.Resource{
		Description: `
Creates and manages an [Aiven for Apache Kafka® Connect](https://aiven.io/docs/products/kafka/kafka-connect) service.
Kafka Connect lets you integrate an Aiven for Apache Kafka® service with external data sources using connectors.

To set up and integrate Kafka Connect:
1. Create a Kafka service in the same Aiven project using the ` + "`aiven_kafka`" + ` resource.
2. Create topics for importing and exporting data using ` + "`aiven_kafka_topic`" + `.
3. Create the Kafka Connect service.
4. Use the ` + "`aiven_service_integration`" + ` resource to integrate the Kafka and Kafka Connect services.
5. Add source and sink connectors using ` + "`aiven_kafka_connector`" + ` resource.
`,
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeKafkaConnect),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: schemautil.CustomizeDiffGenericService(schemautil.ServiceTypeKafkaConnect),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema:         aivenKafkaConnectSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.KafkaConnect(),
	}
}
