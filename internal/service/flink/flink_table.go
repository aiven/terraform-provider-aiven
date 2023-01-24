package flink

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(0, 256),
		Description:  userconfig.Desc("Specifies the name of the table.").ForceNew().Build(),
	},
	"schema_sql": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The SQL statement to create the table.").ForceNew().Build(),
	},
	"integration_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The id of the service integration that is used with this table. It must have the service integration type `flink`.").Referenced().ForceNew().Build(),
	},
	"jdbc_table": {
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		ValidateFunc:  validation.StringLenBetween(0, 128),
		Description:   userconfig.Desc("Name of the jdbc table that is to be connected to this table. Valid if the service integration id refers to a mysql or postgres service.").ForceNew().Build(),
		ConflictsWith: []string{"kafka_connector_type", "kafka_topic", "kafka_key_format", "kafka_value_format", "kafka_key_fields", "kafka_value_fields_include"},
	},
	"kafka_connector_type": {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: true,
		Description: userconfig.Desc("When used as a source, upsert Kafka connectors update values that use an existing key and " +
			"delete values that are null. For sinks, the connector correspondingly writes update or delete " +
			"messages in a compacted topic. If no matching key is found, the values are added as new " +
			"entries. For more information, see the Apache Flink documentation").ForceNew().PossibleValues(schemautil.StringSliceToInterfaceSlice(getFlinkTableKafkaConnectorTypes())...).Build(),
		ValidateFunc:  validation.StringInSlice(getFlinkTableKafkaConnectorTypes(), false),
		ConflictsWith: []string{"jdbc_table", "upsert_kafka"},
	},
	"kafka_topic": {
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		ValidateFunc:  validation.StringLenBetween(1, 249),
		Description:   userconfig.Desc("Name of the kafka topic that is to be connected to this table. Valid if the service integration id refers to a kafka service.").ForceNew().Build(),
		ConflictsWith: []string{"jdbc_table"},
	},
	"kafka_key_fields": {
		Type:        schema.TypeList,
		Optional:    true,
		ForceNew:    true,
		MaxItems:    64,
		Description: userconfig.Desc("Defines an explicit list of physical columns from the table schema that configure the data type for the key format.").ForceNew().Build(),
		Elem: &schema.Schema{
			Type: schema.TypeString},
		ConflictsWith: []string{"jdbc_table", "upsert_kafka"},
	},
	"kafka_key_format": {
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		Description:   userconfig.Desc("Kafka Key Format").ForceNew().PossibleValues(schemautil.StringSliceToInterfaceSlice(getFlinkTableKafkaKeyValueFormats())...).Build(),
		ValidateFunc:  validation.StringInSlice(getFlinkTableKafkaKeyValueFormats(), false),
		ConflictsWith: []string{"jdbc_table", "upsert_kafka"},
	},
	"kafka_value_format": {
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		Description:   userconfig.Desc("Kafka Value Format").ForceNew().PossibleValues(schemautil.StringSliceToInterfaceSlice(getFlinkTableKafkaKeyValueFormats())...).Build(),
		ValidateFunc:  validation.StringInSlice(getFlinkTableKafkaKeyValueFormats(), false),
		ConflictsWith: []string{"jdbc_table", "upsert_kafka"},
	},
	"kafka_startup_mode": {
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		Description:   userconfig.Desc("Startup mode").ForceNew().PossibleValues(schemautil.StringSliceToInterfaceSlice(getFlinkTableKafkaStartupModes())...).Build(),
		ValidateFunc:  validation.StringInSlice(getFlinkTableKafkaStartupModes(), false),
		ConflictsWith: []string{"jdbc_table", "upsert_kafka"},
	},
	"like_options": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Description: userconfig.Desc("[LIKE](https://nightlies.apache.org/flink/flink-docs-master/docs/dev/table/sql/create/#like) statement for table creation.").ForceNew().Build(),
	},
	"kafka_value_fields_include": {
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(getFlinkTableKafkaValueFieldsInclude(), false),
		Description: userconfig.Desc("Controls how key columns are handled in the message value. " +
			"Select ALL to include the physical columns of the table schema in the message value. " +
			"Select EXCEPT_KEY to exclude the physical columns of the table schema from the message value. " +
			"This is the default for upsert Kafka connectors.").PossibleValues(schemautil.StringSliceToInterfaceSlice(getFlinkTableKafkaValueFieldsInclude())).ForceNew().Build(),
		ConflictsWith: []string{"jdbc_table", "upsert_kafka"},
	},
	"opensearch_index": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Description: userconfig.Desc("For an OpenSearch table, the OpenSearch index the table outputs to.").ForceNew().Build(),
	},
	"upsert_kafka": {
		Type:          schema.TypeSet,
		MaxItems:      1,
		Optional:      true,
		ForceNew:      true,
		Description:   "Kafka upsert connector configuration.",
		ConflictsWith: []string{"jdbc_table", "kafka_connector_type", "kafka_key_format", "kafka_value_format", "kafka_key_fields", "kafka_value_fields_include"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key_fields": {
					Type:        schema.TypeList,
					Optional:    true,
					ForceNew:    true,
					MaxItems:    64,
					Description: userconfig.Desc("Defines the columns from the SQL schema of the data table that are considered keys in the Kafka messages.").ForceNew().Build(),
					Elem: &schema.Schema{
						Type: schema.TypeString},
				},
				"key_format": {
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					Description:  userconfig.Desc("Sets the format that is used to convert the key part of Kafka messages.").ForceNew().PossibleValues(schemautil.StringSliceToInterfaceSlice(getFlinkTableKafkaKeyValueFormats())...).Build(),
					ValidateFunc: validation.StringInSlice(getFlinkTableKafkaKeyValueFormats(), false),
				},
				"scan_startup_mode": {
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					Description:  userconfig.Desc("Controls the startup method for the Kafka consumer that Aiven for Apache Flink is using.").ForceNew().PossibleValues(schemautil.StringSliceToInterfaceSlice(getFlinkTableKafkaStartupModes())...).Build(),
					ValidateFunc: validation.StringInSlice(getFlinkTableKafkaScanStartupModes(), false),
				},
				"topic": {
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 249),
					Description:  userconfig.Desc("Topic name").ForceNew().Build(),
				},
				"value_fields_include": {
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringInSlice(getFlinkTableKafkaValueFieldsInclude(), false),
					Description: userconfig.Desc("Controls how key columns are handled in the message value. " +
						"Select ALL to include the physical columns of the table schema in the message value. " +
						"Select EXCEPT_KEY to exclude the physical columns of the table schema from the message value. " +
						"This is the default for upsert Kafka connectors.").PossibleValues(schemautil.StringSliceToInterfaceSlice(getFlinkTableKafkaValueFieldsInclude())).ForceNew().Build(),
				},
				"value_format": {
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					Description:  userconfig.Desc("Sets the format that is used to convert the value part of Kafka messages.").ForceNew().PossibleValues(schemautil.StringSliceToInterfaceSlice(getFlinkTableKafkaKeyValueFormats())...).Build(),
					ValidateFunc: validation.StringInSlice(getFlinkTableKafkaKeyValueFormats(), false),
				},
			},
		},
	},

	// computed fields
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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema:             aivenFlinkTableSchema,
		CustomizeDiff:      customdiff.If(schemautil.ResourceShouldExist, resourceFlinkTableCustomizeDiff),
		DeprecationMessage: schemautil.DeprecationMessage,
	}
}

func resourceFlinkTableRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, tableID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.FlinkTables.Get(project, serviceName, aiven.GetFlinkTableRequest{TableId: tableID})
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
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
	integrationID := d.Get("integration_id").(string)

	// connector options
	jdbcTable := d.Get("jdbc_table").(string)
	kafkaConnectorType := d.Get("kafka_connector_type").(string)
	kafkaTopic := d.Get("kafka_topic").(string)
	kafkaKeyFields := schemautil.FlattenToString(d.Get("kafka_key_fields").([]interface{}))
	kafkaKeyFormat := d.Get("kafka_key_format").(string)
	kafkaValueFormat := d.Get("kafka_value_format").(string)
	kafkaStartupMode := d.Get("kafka_startup_mode").(string)
	likeOptions := d.Get("like_options").(string)
	kafkaValueFieldsInclude := d.Get("kafka_value_fields_include").(string)
	openSearchIndex := d.Get("opensearch_index").(string)

	createRequest := aiven.CreateFlinkTableRequest{
		Name:                    tableName,
		SchemaSQL:               schemaSQL,
		IntegrationId:           integrationID,
		JDBCTable:               jdbcTable,
		KafkaConnectorType:      kafkaConnectorType,
		KafkaTopic:              kafkaTopic,
		KafkaKeyFields:          kafkaKeyFields,
		KafkaKeyFormat:          kafkaKeyFormat,
		KafkaValueFormat:        kafkaValueFormat,
		KafkaStartupMode:        kafkaStartupMode,
		LikeOptions:             likeOptions,
		KafkaValueFieldsInclude: kafkaValueFieldsInclude,
		OpenSearchIndex:         openSearchIndex,
		UpsertKafka:             readUpsertKafkaFromSchema(d),
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

	project, serviceName, tableID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.FlinkTables.Delete(
		project,
		serviceName,
		aiven.DeleteFlinkTableRequest{
			TableId: tableID,
		})
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}
	return nil
}

func resourceFlinkTableCustomizeDiff(_ context.Context, d *schema.ResourceDiff, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName := d.Get("project").(string), d.Get("service_name").(string)

	_, err := client.FlinkTables.Validate(projectName, serviceName, aiven.ValidateFlinkTableRequest{
		Name:                    d.Get("table_name").(string),
		SchemaSQL:               d.Get("schema_sql").(string),
		IntegrationId:           d.Get("integration_id").(string),
		JDBCTable:               d.Get("jdbc_table").(string),
		KafkaConnectorType:      d.Get("kafka_connector_type").(string),
		KafkaTopic:              d.Get("kafka_topic").(string),
		KafkaKeyFields:          schemautil.FlattenToString(d.Get("kafka_key_fields").([]interface{})),
		KafkaKeyFormat:          d.Get("kafka_key_format").(string),
		KafkaValueFormat:        d.Get("kafka_value_format").(string),
		KafkaStartupMode:        d.Get("kafka_startup_mode").(string),
		LikeOptions:             d.Get("like_options").(string),
		KafkaValueFieldsInclude: d.Get("kafka_value_fields_include").(string),
		OpenSearchIndex:         d.Get("opensearch_index").(string),
		UpsertKafka:             readUpsertKafkaFromSchema(d),
	})

	return err
}

func readUpsertKafkaFromSchema(d schemautil.ResourceStateOrResourceDiff) *aiven.FlinkTableUpsertKafka {
	set := d.Get("upsert_kafka").(*schema.Set).List()
	if len(set) == 0 {
		return nil
	}

	r := aiven.FlinkTableUpsertKafka{}
	for _, v := range set {
		cv := v.(map[string]interface{})

		r.KeyFields = schemautil.FlattenToString(cv["key_fields"].([]interface{}))
		r.KeyFormat = cv["key_format"].(string)
		r.ScanStartupMode = cv["scan_startup_mode"].(string)
		r.Topic = cv["topic"].(string)
		r.ValueFieldsInclude = cv["value_fields_include"].(string)
		r.ValueFormat = cv["value_format"].(string)
	}

	return &r
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
		formatDebeziumJSON          = "debezium-json"
		formatJSON                  = "json"
	)
	return []string{formatAvro, formatAvroConfluent, formatDebeziumAvroConfluent, formatDebeziumJSON, formatJSON}
}

func getFlinkTableKafkaValueFieldsInclude() []string {
	const (
		all       = "ALL"
		exceptKey = "EXCEPT_KEY"
	)

	return []string{all, exceptKey}
}

func getFlinkTableKafkaScanStartupModes() []string {
	const (
		startupModeEarliestOffset = "earliest-offset"
		startupModeLatestOffset   = "latest-offset"
	)
	return []string{startupModeEarliestOffset, startupModeLatestOffset}
}
