package topic

import (
	"context"
	"fmt"
	"math"
	"strconv"

	aiven "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkatopic"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
// 2. In the SDKv2 implementation, there was a state drift between `terraform apply` and `terraform refresh` operations:
// -
// - During `refresh`, the config block was fully read (see https://github.com/aiven/terraform-provider-aiven/blob/d80b4818409594f4569d7ece61c1fe069e4192fb/internal/sdkprovider/service/kafkatopic/kafka_topic.go#L575)
// - "Unknown" fields were suppressed but remained in state (see https://github.com/aiven/terraform-provider-aiven/blob/d80b4818409594f4569d7ece61c1fe069e4192fb/internal/sdkprovider/service/kafkatopic/kafka_topic.go#L244C32-L244C59)
// This new implementation removes any fields not explicitly set by the user from the API response before they reach Terraform.
// While technically a breaking change, this provides more consistent user experience by eliminating state drift.
//
// 3. Fields "partitions" and "replication" are actually optional+computed: when values are not provided, the BE sets defaults.
// However, our provider doesn't read the topic after creation to speed up the things.
// Probably, this is the reason why these fields were made required in the first place in SDKv2.
//
// 4. The "config" block is implemented as a NestedBlock for backward compatibility with SDKv2.
// Technically all the fields in the config are computed+optional, there are default values
// for each field (see https://api.aiven.io/doc/#tag/Service:_Kafka/operation/ServiceKafkaTopicGet).
// However, the plugin framework doesn't support computed blocks (see
// https://github.com/hashicorp/terraform-plugin-framework/issues/883), making this incompatible
// with the current solution. Luckily, since we don't read the resource, we don't set computed values anyway.
// The fields are marked as optional only. Instead, missing fields in the plan are
// removed from the API response to suppress unwanted diff and provide consistent UX between
// create, update (do not read at all) and refresh (reads resource from the API) operations.
//
// 5. There are legacy integer fields that are implemented as strings (see legacyFields()). Though Terraform
// converts types automatically, some operations like equality checks might not do that.
// For backward compatibility we keep these fields as strings.
// See https://developer.hashicorp.com/terraform/language/expressions/types#type-conversion
//
// 6. When a Kafka service is powered off, all topics are deleted automatically.
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
	isResource bool // See usage in Read()
}

func (vw *view) ResValidateConfig(ctx context.Context, config *tfModel) diag.Diagnostics {
	return lo.Flatten([]diag.Diagnostics{
		vw.validateTopicConfig(ctx, config),
		vw.validateAlreadyExists(ctx, config),
	})
}

func (vw *view) Create(ctx context.Context, plan *tfModel) diag.Diagnostics {
	var req kafkatopic.ServiceKafkaTopicCreateIn
	diags := expandData(ctx, plan, nil, &req, modifyRequest)
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

	// We do not call a Kafka Topic read here to speed up the performance.
	plan.SetID(plan.Project.ValueString(), plan.ServiceName.ValueString(), plan.TopicName.ValueString())
	return nil
}

func (vw *view) Update(ctx context.Context, plan, state *tfModel) diag.Diagnostics {
	var req kafkatopic.ServiceKafkaTopicUpdateIn
	diags := expandData(ctx, plan, state, &req, modifyRequest)
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

	// We do not call a Kafka Topic read here to speed up the performance.
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
		if vw.isResource && aiven.IsNotFound(err) && !state.ID.IsNull() {
			// When a Kafka service is powered off, all topics are deleted automatically.
			// See: https://aiven.io/docs/platform/concepts/service-power-cycle#power-off-a-service
			//
			// This handles two cases:
			// 1. Auto-recreates the topic when the service is powered back on
			// 2. Skips recreation during `terraform import` since the topic doesn't exist (checks ID)
			//
			// Note: Create() doesn't read the resource state after creation.
			return vw.Create(ctx, state)
		}

		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	modifyResponse, diags := responseModifier(ctx, state)
	if diags.HasError() {
		return diags
	}

	return flattenData(ctx, state, rsp, modifyResponse)
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

// modifyRequest Turns topic config legacy fields (strings) to integers.
func modifyRequest[T any](req map[string]any, _ *T) error {
	// Request always has "config" field.
	config, _ := req["config"].(map[string]any)
	return walkConfig(config)
}

// responseModifier
// responseModifier modifies the Kafka topic response in three ways:
// 1. Converts topic config legacy fields from integers to strings for Terraform compatibility
// 2. Removes any config fields not explicitly set by the user to prevent unwanted diffs
// 3. Converts the "partitions" field from a list of partition objects to a simple integer count
func responseModifier(ctx context.Context, state *tfModel) (util.MapModifier[kafkatopic.ServiceKafkaTopicGetOut], diag.Diagnostics) {
	// Extracts plan values from the config to keep them in the state.
	// The rest are removed to suppress the diff.
	userConfig, diags := stateTopicConfig(ctx, state, errmsg.SummaryErrorReadingResource)
	if diags.HasError() {
		return nil, diags
	}

	return func(rsp map[string]any, in *kafkatopic.ServiceKafkaTopicGetOut) error {
		// Response has "partitions" field as a list of objects, not an integer.
		// https://api.aiven.io/doc/#tag/Service:_Kafka/operation/ServiceKafkaTopicGet
		rsp["partitions"] = len(in.Partitions)

		// Erases the config if user hasn't set it.
		// TF doesn't support computed+optional blocks.
		// https://github.com/hashicorp/terraform-plugin-framework/issues/883
		// We must remove it to suppress the diff, even an empty object counts.
		delete(rsp, "config")
		if userConfig == nil {
			return nil
		}

		// We cant iterate over struct fields, turns into a map of same structure.
		var mapCfg mapConfig
		err := schemautil.Remarshal(in, &mapCfg)
		if err != nil {
			return err
		}

		cfg := make(map[string]any)
		for k, v := range mapCfg.Config {
			if _, ok := userConfig[k]; ok {
				cfg[k] = v.Value
			}
		}

		rsp["config"] = cfg
		return walkConfig(cfg)
	}, nil
}

// mapConfig Implements ServiceKafkaTopicGet response in a map to iterate over keys.
// https://api.aiven.io/doc/#tag/Service:_Kafka/operation/ServiceKafkaTopicGet
type mapConfig struct {
	Config map[string]struct {
		Value any `json:"value"`
	} `json:"config"`
}

// walkConfig walks over config values and converts legacy fields.
func walkConfig(config map[string]any) error {
	var allErr error
	legacy := legacyFields()
	for k, v := range config {
		if !legacy[k] {
			continue
		}

		val, err := convertLegacy(v)
		if err != nil {
			allErr = multierror.Append(allErr, fmt.Errorf("config field %q=%v, error: %w", k, v, err))
		}
		config[k] = val
	}
	return allErr
}

// convertLegacy converts string <-> integer values.
func convertLegacy(v any) (any, error) {
	switch t := v.(type) {
	case float64:
		// json.Marshal has turned integers into float64. Want string.
		// Check bounds before converting
		if t < float64(math.MinInt) || t > float64(math.MaxInt) {
			// This should never happen but just in case.
			return nil, fmt.Errorf("value %v out of int range", t)
		}
		return strconv.FormatInt(int64(t), 10), nil
	case string:
		// Want integer, but it's a string in the state.
		return strconv.ParseInt(t, 10, 64)
	}
	return nil, fmt.Errorf("unknown type %T", v)
}

// legacyFields these fields are string in our schema,
// but they are integers in the API.
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

// stateTopicConfig turns the config block into a map.
func stateTopicConfig(ctx context.Context, state *tfModel, summaryErr string) (map[string]any, diag.Diagnostics) {
	if state.Config.IsNull() {
		return nil, nil
	}

	config, diags := util.ExpandSingleNested(ctx, expandConfig, state.Config)
	if diags.HasError() {
		return nil, diags
	}

	var userConfig map[string]any
	err := schemautil.Remarshal(config, &userConfig)
	if err != nil {
		diags.AddError(summaryErr, err.Error())
		return nil, diags
	}

	return userConfig, nil
}
