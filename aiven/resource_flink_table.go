package aiven

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenFlinkTableSchema = map[string]*schema.Schema{
	"project": {
		Type:        schema.TypeString,
		Description: "Project to link the Flink Table to",
		Required:    true,
		ForceNew:    true,
	},
	"service_name": {
		Type:        schema.TypeString,
		Description: "Service to link the Flink Table to",
		Required:    true,
		ForceNew:    true,
	},
	"table_id": {
		Type:        schema.TypeString,
		Description: "Table id of the flink table",
		Computed:    true,
	},
	"table_name": {
		Type:        schema.TypeString,
		Description: "Table name of the flink table",
		Required:    true,
		ForceNew:    true,
	},
	"integration_id": {
		Type:        schema.TypeString,
		Description: "Id of the service integration this table applies to",
		Required:    true,
		ForceNew:    true,
	},
	"jdbc_table": {
		Type:        schema.TypeString,
		Description: "Table name in Database, used as a source/sink. (PG integration only)",
		Optional:    true,
		ForceNew:    true,
	},
	"kafka_topic": {
		Type:        schema.TypeString,
		Description: "Topic name, used as a source/sink. (Kafka integration only)",
		Optional:    true,
		ForceNew:    true,
	},
	"like_options": {
		Type:        schema.TypeString,
		Description: "Clause can be used to create a table based on a definition of an existing table",
		Optional:    true,
		ForceNew:    true,
	},
	"partitioned_by": {
		Type:        schema.TypeString,
		Description: "A column from a schema, table will be partitioned by",
		Optional:    true,
		ForceNew:    true,
	},
	"schema_sql": {
		Type:        schema.TypeString,
		Description: "Source/Sink table schema",
		Required:    true,
		ForceNew:    true,
	},
}

func resourceFlinkTable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFlinkTableCreate,
		ReadContext:   resourceFlinkTableRead,
		DeleteContext: resourceFlinkTableDelete,
		Schema:        aivenFlinkTableSchema,
	}
}

func resourceFlinkTableRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, tableId := splitResourceID3(d.Id())

	r, err := client.FlinkTables.Get(project, serviceName, aiven.GetFlinkTableRequest{TableId: tableId})
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
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

	return nil
}

func resourceFlinkTableCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	integrationId := d.Get("integration_id").(string)
	jdbcTable := d.Get("jdbc_table").(string)
	kafkaTopic := d.Get("kafka_topic").(string)
	likeOptions := d.Get("like_options").(string)
	tableName := d.Get("table_name").(string)
	partitionedBy := d.Get("partitioned_by").(string)
	schemaSQL := d.Get("schema_sql").(string)

	createRequest := aiven.CreateFlinkTableRequest{
		IntegrationId: integrationId,
		JDBCTable:     jdbcTable,
		KafkaTopic:    kafkaTopic,
		LikeOptions:   likeOptions,
		Name:          tableName,
		PartitionedBy: partitionedBy,
		SchemaSQL:     schemaSQL,
	}

	r, err := client.FlinkTables.Create(project, serviceName, createRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildResourceID(project, serviceName, r.TableId))

	return resourceFlinkTableRead(ctx, d, m)
}

func resourceFlinkTableDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, tableId := splitResourceID3(d.Id())

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
