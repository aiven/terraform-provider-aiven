package kafkatopic

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// typedConfigValue converts TF config value to API config value.
// TF has some fields defined as strings, but they are expected to be numbers or booleans in the API.
func typedConfigValue(k string, v any) (any, error) {
	apiType, ok := apiConfigTypes()[k]
	if !ok {
		return nil, fmt.Errorf("unknown config field key %q", k)
	}

	s := fmt.Sprintf("%v", v)
	switch apiType {
	case schema.TypeString:
		return s, nil
	case schema.TypeFloat:
		return strconv.ParseFloat(s, 64)
	case schema.TypeInt:
		return strconv.ParseInt(s, 10, 64)
	case schema.TypeBool:
		return strconv.ParseBool(s)
	default:
		return v, nil
	}
}

// apiConfigTypes returns API types for Kafka topic config fields.
// https://api.aiven.io/doc/#tag/Service:_Kafka/operation/ServiceKafkaTopicCreate
func apiConfigTypes() map[string]schema.ValueType {
	return map[string]schema.ValueType{
		"cleanup_policy":                      schema.TypeString,
		"compression_type":                    schema.TypeString,
		"delete_retention_ms":                 schema.TypeInt,
		"file_delete_delay_ms":                schema.TypeInt,
		"flush_messages":                      schema.TypeInt,
		"flush_ms":                            schema.TypeInt,
		"index_interval_bytes":                schema.TypeInt,
		"local_retention_bytes":               schema.TypeInt,
		"local_retention_ms":                  schema.TypeInt,
		"max_compaction_lag_ms":               schema.TypeInt,
		"max_message_bytes":                   schema.TypeInt,
		"message_downconversion_enable":       schema.TypeBool,
		"message_format_version":              schema.TypeString,
		"message_timestamp_difference_max_ms": schema.TypeInt,
		"message_timestamp_type":              schema.TypeString,
		"min_cleanable_dirty_ratio":           schema.TypeFloat,
		"min_compaction_lag_ms":               schema.TypeInt,
		"min_insync_replicas":                 schema.TypeInt,
		"preallocate":                         schema.TypeBool,
		"remote_storage_enable":               schema.TypeBool,
		"retention_bytes":                     schema.TypeInt,
		"retention_ms":                        schema.TypeInt,
		"segment_bytes":                       schema.TypeInt,
		"segment_index_bytes":                 schema.TypeInt,
		"segment_jitter_ms":                   schema.TypeInt,
		"segment_ms":                          schema.TypeInt,
		"unclean_leader_election_enable":      schema.TypeBool,
	}
}
