package flink

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenFlinkTableSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"table_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: schemautil.Complex("Specifies the name of the table.").ForceNew().Build(),
	},
	"schema_sql": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: schemautil.Complex("The SQL statement to create the table.").ForceNew().Build(),
	},
	"integration_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: schemautil.Complex("The id of the service integration that is used with this table. It must have the service integration type `flink`.").Referenced().ForceNew().Build(),
	},
	"jdbc_table": {
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		Description:   schemautil.Complex("Name of the jdbc table that is to be connected to this table. Valid if the service integration id refers to a mysql or postgres service.").ForceNew().Build(),
		ConflictsWith: []string{"kafka_connector_type", "kafka_topic", "kafka_key_format", "kafka_value_format", "kafka_key_fields"},
	},
	"kafka_connector_type": {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: true,
		Description: schemautil.Complex("When used as a source, upsert Kafka connectors update values that use an existing key and " +
			"delete values that are null. For sinks, the connector correspondingly writes update or delete " +
			"messages in a compacted topic. If no matching key is found, the values are added as new " +
			"entries. For more information, see the Apache Flink documentation").ForceNew().PossibleValues(schemautil.StringSliceToInterfaceSlice(getFlinkTableKafkaConnectorTypes())...).Build(),
		ValidateFunc:  validation.StringInSlice(getFlinkTableKafkaConnectorTypes(), false),
		ConflictsWith: []string{"jdbc_table"},
	},
	"kafka_topic": {
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		Description:   schemautil.Complex("Name of the kafka topic that is to be connected to this table. Valid if the service integration id refers to a kafka service.").ForceNew().Build(),
		ConflictsWith: []string{"jdbc_table"},
	},
	"kafka_key_fields": {
		Type:        schema.TypeList,
		Optional:    true,
		ForceNew:    true,
		Description: schemautil.Complex("Defines an explicit list of physical columns from the table schema that configure the data type for the key format.").ForceNew().Build(),
		Elem: &schema.Schema{
			Type: schema.TypeString},
		ConflictsWith: []string{"jdbc_table"},
	},
	"kafka_key_format": {
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		Description:   schemautil.Complex("Kafka Key Format").ForceNew().PossibleValues(schemautil.StringSliceToInterfaceSlice(getFlinkTableKafkaKeyValueFormats())...).Build(),
		ValidateFunc:  validation.StringInSlice(getFlinkTableKafkaKeyValueFormats(), false),
		ConflictsWith: []string{"jdbc_table"},
	},
	"kafka_value_format": {
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		Description:   schemautil.Complex("Kafka Value Format").ForceNew().PossibleValues(schemautil.StringSliceToInterfaceSlice(getFlinkTableKafkaKeyValueFormats())...).Build(),
		ValidateFunc:  validation.StringInSlice(getFlinkTableKafkaKeyValueFormats(), false),
		ConflictsWith: []string{"jdbc_table"},
	},
	"kafka_startup_mode": {
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		Description:   schemautil.Complex("Startup mode").ForceNew().PossibleValues(schemautil.StringSliceToInterfaceSlice(getFlinkTableKafkaStartupModes())...).Build(),
		ValidateFunc:  validation.StringInSlice(getFlinkTableKafkaStartupModes(), false),
		ConflictsWith: []string{"jdbc_table"},
	},
	"like_options": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Description: schemautil.Complex("[LIKE](https://nightlies.apache.org/flink/flink-docs-master/docs/dev/table/sql/create/#like) statement for table creation.").ForceNew().Build(),
	},
	"table_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The Table ID of the flink table in the flink service.",
	},
}

func ResourceFlinkTable() *schema.Resource {
	return &schema.Resource{
		Description:   "The Flink Table resource allows the creation and management of Aiven Tables.",
		CreateContext: resourceFlinkTableCreate,
		ReadContext:   resourceFlinkTableRead,
		DeleteContext: resourceFlinkTableDelete,
		Schema:        aivenFlinkTableSchema,
	}
}

func resourceFlinkTableRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, tableId := schemautil.SplitResourceID3(d.Id())

	r, err := client.FlinkTables.Get(project, serviceName, aiven.GetFlinkTableRequest{TableId: tableId})
	if err != nil {
		return diag.FromErr(service.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("project", project); err != nil {
		return diag.Errorf("error setting Flink Tables `project` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.Errorf("error setting Flink Tables `service_name` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("integration_id", r.IntegrationId); err != nil {
		return diag.Errorf("error setting Flink Tables `integration_id` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("table_id", r.TableId); err != nil {
		return diag.Errorf("error setting Flink Tables `table_id` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("table_name", r.TableName); err != nil {
		return diag.Errorf("error setting Flink Tables `table_name` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("schema_sql", r.SchemaSQL); err != nil {
		return diag.Errorf("error setting Flink Tables `schema_sql` for resource %s: %s", d.Id(), err)
	}

	return nil
}

func resourceFlinkTableCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	tableName := d.Get("table_name").(string)
	schemaSQL := d.Get("schema_sql").(string)
	integrationId := d.Get("integration_id").(string)

	// connector options
	jdbcTable := d.Get("jdbc_table").(string)
	kafkaConnectorType := d.Get("kafka_connector_type").(string)
	kafkaTopic := d.Get("kafka_topic").(string)
	kafkaKeyFields := schemautil.FlattenToString(d.Get("kafka_key_fields").([]interface{}))
	kafkaKeyFormat := d.Get("kafka_key_format").(string)
	kafkaValueFormat := d.Get("kafka_value_format").(string)
	kafkaStartupMode := d.Get("kafka_startup_mode").(string)
	likeOptions := d.Get("like_options").(string)

	createRequest := aiven.CreateFlinkTableRequest{
		Name:               tableName,
		SchemaSQL:          schemaSQL,
		IntegrationId:      integrationId,
		JDBCTable:          jdbcTable,
		KafkaConnectorType: kafkaConnectorType,
		KafkaTopic:         kafkaTopic,
		KafkaKeyFields:     kafkaKeyFields,
		KafkaKeyFormat:     kafkaKeyFormat,
		KafkaValueFormat:   kafkaValueFormat,
		KafkaStartupMode:   kafkaStartupMode,
		LikeOptions:        likeOptions,
	}

	r, err := client.FlinkTables.Create(project, serviceName, createRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, r.TableId))

	return resourceFlinkTableRead(ctx, d, m)
}

func resourceFlinkTableDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, tableId := schemautil.SplitResourceID3(d.Id())

	err := client.FlinkTables.Delete(
		project,
		serviceName,
		aiven.DeleteFlinkTableRequest{
			TableId: tableId,
		})
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}
	return nil
}

func getFlinkTableKafkaConnectorTypes() []string {
	const (
		connectorTypeKafka       = "kafka"
		connectorTypeUpsertKafka = "upsert-kafka"
	)
	return []string{connectorTypeKafka, connectorTypeUpsertKafka}
}

func getFlinkTableKafkaStartupModes() []string {
	const (
		startupModeEarliestOffset = "earliest-offset"
		startupModeLatestOffset   = "latest-offset"
		startupModeGroupOffsets   = "group-offsets"
		startupModeTimestamp      = "timestamp"
	)
	return []string{startupModeEarliestOffset, startupModeLatestOffset, startupModeGroupOffsets, startupModeTimestamp}
}

func getFlinkTableKafkaKeyValueFormats() []string {
	const (
		formatAvro                  = "avro"
		formatAvroConfluent         = "avro-confluent"
		formatDebeziumAvroConfluent = "debezium-avro-confluent"
		formatDebeziumJson          = "debezium-json"
		formatJson                  = "json"
	)
	return []string{formatAvro, formatAvroConfluent, formatDebeziumAvroConfluent, formatDebeziumJson, formatJson}
}
