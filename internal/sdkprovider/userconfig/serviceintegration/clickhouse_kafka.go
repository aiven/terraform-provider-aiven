// Code generated by user config generator. DO NOT EDIT.

package serviceintegration

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/diff"
)

func clickhouseKafkaUserConfig() *schema.Schema {
	return &schema.Schema{
		Description:      "ClickhouseKafka user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later",
		DiffSuppressFunc: diff.SuppressUnchanged,
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{"tables": {
			Description: "Tables to create",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"auto_offset_reset": {
					Description:  "Enum: `beginning`, `earliest`, `end`, `largest`, `latest`, `smallest`. Action to take when there is no initial offset in offset store or the desired offset is out of range. Default: `earliest`.",
					Optional:     true,
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"beginning", "earliest", "end", "largest", "latest", "smallest"}, false),
				},
				"columns": {
					Description: "Table columns",
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"name": {
							Description: "Column name. Example: `key`.",
							Required:    true,
							Type:        schema.TypeString,
						},
						"type": {
							Description: "Column type. Example: `UInt64`.",
							Required:    true,
							Type:        schema.TypeString,
						},
					}},
					MaxItems: 100,
					Required: true,
					Type:     schema.TypeList,
				},
				"data_format": {
					Description:  "Enum: `Avro`, `AvroConfluent`, `CSV`, `JSONAsString`, `JSONCompactEachRow`, `JSONCompactStringsEachRow`, `JSONEachRow`, `JSONStringsEachRow`, `MsgPack`, `Parquet`, `RawBLOB`, `TSKV`, `TSV`, `TabSeparated`. Message data format. Default: `JSONEachRow`.",
					Required:     true,
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"Avro", "AvroConfluent", "CSV", "JSONAsString", "JSONCompactEachRow", "JSONCompactStringsEachRow", "JSONEachRow", "JSONStringsEachRow", "MsgPack", "Parquet", "RawBLOB", "TSKV", "TSV", "TabSeparated"}, false),
				},
				"date_time_input_format": {
					Description:  "Enum: `basic`, `best_effort`, `best_effort_us`. Method to read DateTime from text input formats. Default: `basic`.",
					Optional:     true,
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"basic", "best_effort", "best_effort_us"}, false),
				},
				"group_name": {
					Description: "Kafka consumers group. Default: `clickhouse`.",
					Required:    true,
					Type:        schema.TypeString,
				},
				"handle_error_mode": {
					Description:  "Enum: `default`, `stream`. How to handle errors for Kafka engine. Default: `default`.",
					Optional:     true,
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"default", "stream"}, false),
				},
				"max_block_size": {
					Description: "Number of row collected by poll(s) for flushing data from Kafka. Default: `0`.",
					Optional:    true,
					Type:        schema.TypeInt,
				},
				"max_rows_per_message": {
					Description: "The maximum number of rows produced in one kafka message for row-based formats. Default: `1`.",
					Optional:    true,
					Type:        schema.TypeInt,
				},
				"name": {
					Description: "Name of the table. Example: `events`.",
					Required:    true,
					Type:        schema.TypeString,
				},
				"num_consumers": {
					Description: "The number of consumers per table per replica. Default: `1`.",
					Optional:    true,
					Type:        schema.TypeInt,
				},
				"poll_max_batch_size": {
					Description: "Maximum amount of messages to be polled in a single Kafka poll. Default: `0`.",
					Optional:    true,
					Type:        schema.TypeInt,
				},
				"poll_max_timeout_ms": {
					Description: "Timeout in milliseconds for a single poll from Kafka. Takes the value of the stream_flush_interval_ms server setting by default (500ms). Default: `0`.",
					Optional:    true,
					Type:        schema.TypeInt,
				},
				"skip_broken_messages": {
					Description: "Skip at least this number of broken messages from Kafka topic per block. Default: `0`.",
					Optional:    true,
					Type:        schema.TypeInt,
				},
				"thread_per_consumer": {
					Description: "Provide an independent thread for each consumer. All consumers run in the same thread by default. Default: `false`.",
					Optional:    true,
					Type:        schema.TypeBool,
				},
				"topics": {
					Description: "Kafka topics",
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{"name": {
						Description: "Name of the topic. Example: `topic_name`.",
						Required:    true,
						Type:        schema.TypeString,
					}}},
					MaxItems: 100,
					Required: true,
					Type:     schema.TypeList,
				},
			}},
			MaxItems: 400,
			Optional: true,
			Type:     schema.TypeList,
		}}},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	}
}
