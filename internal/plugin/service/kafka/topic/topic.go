package topic

import (
	"context"
	"fmt"
	"strconv"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkatopic"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/kafkatopicrepository"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// Resource limitations and notes:
//
// 1. ServiceKafkaTopicCreate and ServiceKafkaTopicUpdate do not return the resource back
// (see https://api.aiven.io/doc/#tag/Service:_Kafka/operation/ServiceKafkaTopicCreate).
// For performance optimization, we avoid calling Read() after create/update operations.
// The Read() process in kafkatopicrepository is a heavy operation that fetches values directly
// from Kafka, potentially taking up to a minute per topic.
// This behavior is ported from the SDKv2 implementation.
// https://github.com/aiven/terraform-provider-aiven/blob/d80b4818409594f4569d7ece61c1fe069e4192fb/internal/sdkprovider/service/kafkatopic/kafka_topic.go#L366
//
// 2. State Drift Behavior:
//   - During `terraform apply`: Response is not read (see note 1) to optimize performance.
//     Instead, SDKv2 sets all fields in state to zero values ("", 0, false) since it cannot
//     distinguish between zero and missing values.
//   - During `terraform refresh`: The config block is fully read with actual values from Kafka
//     (see SDKv2 implementation at internal/sdkprovider/service/kafkatopic/kafka_topic.go#L575)
//
//   Maintains backwards compatibility by pre-filling the config block with zero values:
//   - Ensures all computed fields have values as required by Plugin Framework
//   - Works around Plugin Framework's lack of computed block support
//     (tracked in github.com/hashicorp/terraform-plugin-framework/issues/883)
//
// 3. Fields "partitions" and "replication" are actually optional+computed: when values are not provided, the BE sets defaults.
// However, our provider doesn't read the topic after creation to speed up the things.
// Probably, this is the reason why these fields were made required in the first place in SDKv2.
//
// 4. There are legacy integer fields that are implemented as strings (see legacyFields()). Though Terraform
// converts types automatically, some operations like equality checks might not do that.
// For backward compatibility, we keep these fields as strings.
// See https://developer.hashicorp.com/terraform/language/expressions/types#type-conversion
//
// 5. When a Kafka service is powered off and it doesn't have backups, all topics are deleted.
// See: https://aiven.io/docs/platform/concepts/service-power-cycle#power-off-a-service
// The Read() method recreates missing topics after service restarts, avoiding manual state cleanup (`terraform state rm`).

func NewResource() resource.Resource {
	return adapter.NewResource(aivenName, &view{isResource: true}, newResourceSchema, newResourceModel, composeID())
}

func NewDatasource() datasource.DataSource {
	return adapter.NewDatasource(aivenName, new(view), newDatasourceSchema, newDatasourceModel)
}

var _ adapter.ResValidateConfig[tfModel] = (*view)(nil)

type view struct {
	adapter.View
	isResource bool // See read()
}

func (vw *view) ResValidateConfig(ctx context.Context, config *tfModel) diag.Diagnostics {
	return lo.Flatten([]diag.Diagnostics{
		vw.validateTopicConfig(ctx, config),
		vw.validateAlreadyExists(ctx, config),
	})
}

func (vw *view) Create(ctx context.Context, plan *tfModel) diag.Diagnostics {
	var req kafkatopic.ServiceKafkaTopicCreateIn
	diags := expandData(ctx, plan, nil, &req, modifyRequestConfig)
	if diags.HasError() {
		return diags
	}

	err := kafkatopicrepository.New(vw.Client).Create(
		ctx,
		plan.Project.ValueString(),
		plan.ServiceName.ValueString(),
		&req,
	)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return diags
	}

	// Doesn't call Read() after create to avoid heavy reading.
	plan.SetID(plan.Project.ValueString(), plan.ServiceName.ValueString(), plan.TopicName.ValueString())
	return nil
}

func (vw *view) Update(ctx context.Context, plan, state *tfModel) diag.Diagnostics {
	var req kafkatopic.ServiceKafkaTopicUpdateIn
	diags := expandData(ctx, plan, state, &req, modifyRequestConfig)
	if diags.HasError() {
		return diags
	}

	err := kafkatopicrepository.New(vw.Client).Update(
		ctx,
		plan.Project.ValueString(),
		plan.ServiceName.ValueString(),
		plan.TopicName.ValueString(),
		&req,
	)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorUpdatingResource, err.Error())
		return diags
	}

	// Doesn't call Read() after create to avoid heavy reading.
	return nil
}

func (vw *view) Read(ctx context.Context, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := kafkatopicrepository.New(vw.Client).Read(
		ctx,
		state.Project.ValueString(),
		state.ServiceName.ValueString(),
		state.TopicName.ValueString(),
	)
	if err != nil {
		if vw.isResource && avngen.IsNotFound(err) && !state.ID.IsNull() {
			// When a Kafka service (without backups) is powered off, all topics and their configurations are deleted.
			// See:
			// - https://aiven.io/docs/platform/concepts/service-power-cycle#power-off-a-service
			// - https://aiven.io/docs/products/kafka/concepts/configuration-backup#how-backups-work
			// - https://github.com/aiven/terraform-provider-aiven/issues/1004
			//
			// This handles two cases:
			// 1. Auto-recreates the topic when the service is powered back on
			// 2. Skips recreation during `terraform import` since the topic doesn't exist (checks ID)
			return vw.Create(ctx, state)
		}

		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	return flattenData(ctx, state, rsp, modifyResponsePartitions, modifyResponseConfig)
}

func (vw *view) Delete(ctx context.Context, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	err := kafkatopicrepository.New(vw.Client).Delete(
		ctx,
		state.Project.ValueString(),
		state.ServiceName.ValueString(),
		state.TopicName.ValueString(),
	)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}

	return nil
}

func modifyRequestConfig[T any](req util.RawMap, _ *T) error {
	// Turns topic config legacy fields (strings) to integers.
	for k := range legacyFields() {
		err := modifyRequestConfigKey(req, k)
		if err != nil {
			return fmt.Errorf("failed to modify request config key %q: %w", k, err)
		}
	}
	return nil
}

func modifyRequestConfigKey(req util.RawMap, key string) error {
	v, err := req.GetString("config", key)
	if err != nil {
		if util.IsKeyNotFound(err) {
			return nil
		}
		return err
	}

	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return err
	}
	return req.Set(i, "config", key)
}

func modifyResponsePartitions(rsp util.RawMap, in *kafkatopic.ServiceKafkaTopicGetOut) error {
	// Response has "partitions" field as a list of objects, not an integer.
	// https://api.aiven.io/doc/#tag/Service:_Kafka/operation/ServiceKafkaTopicGet
	return rsp.Set(len(in.Partitions), "partitions")
}

func modifyResponseConfig(rsp util.RawMap, in *kafkatopic.ServiceKafkaTopicGetOut) error {
	// Converts legacy config fields to strings.
	// We cant iterate over struct fields, turns into a map of same structure.
	var mapCfg mapConfig
	err := schemautil.Remarshal(in, &mapCfg)
	if err != nil {
		return err
	}

	// Creates an empty config with zero values and fills it with values from the response.
	// This serves two purposes:
	//
	// 1. Backwards compatibility with SDKv2:
	//    - SDKv2 didn't distinguish between empty and null values
	//    - All fields as a side-effect had values in the state, even if not set
	//    - Setting zeros here prevents unnecessary diffs when user migrates from the SDKv2 version.
	//    - Real values will be populated on next `terraform refresh`
	//
	// 2. Handling missing response fields:
	//    - Terraform requires all Computed fields to have a value
	//    - If a field is missing in the API response (e.g. due to a backend issue or API change), Terraform would fail with an error
	//    - Default zero values ensure all required fields are populated
	config := emptyConfig()
	legacy := legacyFields()
	for k, v := range mapCfg.Config {
		if legacy[k] {
			// Legacy v.Value is json.Number which is string, as we need.
			config[k] = fmt.Sprint(v.Value)
		} else {
			config[k] = v.Value
		}
	}

	for k, v := range config {
		err = rsp.Set(v, "config", k)
		if err != nil {
			return fmt.Errorf("failed to set legacy config field %s=%v: %w", k, v, err)
		}
	}

	return nil
}

// mapConfig Implements ServiceKafkaTopicGet response in a map to iterate over keys.
// https://api.aiven.io/doc/#tag/Service:_Kafka/operation/ServiceKafkaTopicGet
type mapConfig struct {
	Config map[string]struct {
		Value any `json:"value"`
	} `json:"config"`
}

// legacyFields these fields are string in our schema,
// but they are integers in the API.
// The SDKv2 didn't distinguish between empty and zero values: `0 == 0`.
// Probably, this is why these fields are strings, e.g. `"" != 0`.
func legacyFields() map[string]bool {
	return map[string]bool{
		"delete_retention_ms":                 true,
		"file_delete_delay_ms":                true,
		"flush_messages":                      true,
		"flush_ms":                            true,
		"index_interval_bytes":                true,
		"local_retention_bytes":               true,
		"local_retention_ms":                  true,
		"max_compaction_lag_ms":               true,
		"max_message_bytes":                   true,
		"message_timestamp_difference_max_ms": true,
		"min_compaction_lag_ms":               true,
		"min_insync_replicas":                 true,
		"retention_bytes":                     true,
		"retention_ms":                        true,
		"segment_bytes":                       true,
		"segment_index_bytes":                 true,
		"segment_jitter_ms":                   true,
		"segment_ms":                          true,
	}
}

// emptyConfig returns a map of all config keys with zero Go values.
func emptyConfig() map[string]any {
	atts := newResourceSchema(context.Background()).GetBlocks()["config"].GetNestedObject().GetAttributes()
	values := make(map[string]any, len(atts))
	for k, v := range atts {
		switch v.GetType() {
		case types.StringType:
			values[k] = ""
		case types.BoolType:
			values[k] = false
		case types.Int64Type, types.Float64Type:
			values[k] = 0
		default:
			// This is fatal and must fail during development.
			panic(fmt.Sprintf("unsupported config attribute type %T for key %q", v.GetType(), k))
		}
	}
	return values
}
