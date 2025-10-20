package topic

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkatopic"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

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
// 2. State Drift Behavior since SDKv2:
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
// 3. Fields "partitions" and "replication" are actually computed+optional: when values are not provided, the BE sets defaults.
// However, our provider doesn't read the topic after creation to speed up the things.
// Probably, this is the reason why these fields were made required in the first place in SDKv2.
//
// 4. There are legacy integer fields that are implemented as strings (see legacyStringFields). Though Terraform
// converts types automatically, some operations like equality checks might not do that.
// For backward compatibility, we keep these fields as strings.
// See https://developer.hashicorp.com/terraform/language/expressions/types#type-conversion
//
// 5. When a Kafka service is powered off and it doesn't have backups, all topics are deleted.
// See: https://aiven.io/docs/platform/concepts/service-power-cycle#power-off-a-service
// The Read() method recreates missing topics after service restarts, avoiding manual state cleanup (`terraform state rm`).
//
// 6. Unlike SDKv2, removing the "config" block now shows in plan output due to Plugin Framework's type system
// (see https://github.com/hashicorp/terraform-plugin-framework/issues/1030#issuecomment-2322378726).
// While technically a breaking change, we must accept it as implementing custom types would require extensive
// modifications to utility functions and the code generation.

func NewResource() resource.Resource {
	return adapter.NewResource(aivenName, &view{isResource: true}, newResourceSchema, newResourceModel, composeID())
}

func NewDatasource() datasource.DataSource {
	return adapter.NewDatasource(aivenName, new(view), newDatasourceSchema, newDatasourceModel)
}

var (
	_ adapter.ResValidateConfig[tfModel] = (*view)(nil)
	_ adapter.ResModifyPlan[tfModel]     = (*view)(nil)
)

type view struct {
	adapter.View
	isResource bool // See read()
}

func (vw *view) ResValidateConfig(ctx context.Context, config *tfModel) diag.Diagnostics {
	return util.MergeSlices(
		vw.validateTopicConfig(ctx, config),
		vw.validateAlreadyExists(ctx, config),
	)
}

func (vw *view) ResModifyPlan(ctx context.Context, plan, state, config *tfModel) diag.Diagnostics {
	// todo: remove in v5.0.0. Termination protection is a virtual field
	if plan.ID.IsNull() && state.TerminationProtection.ValueBool() {
		var diags diag.Diagnostics
		diags.AddError(
			errmsg.SummaryErrorDeletingResource,
			fmt.Sprintf("Termination protection is enabled, cannot delete topic %q", state.TopicName.ValueString()),
		)
		return diags
	}

	return nil
}

func (vw *view) Create(ctx context.Context, plan *tfModel) diag.Diagnostics {
	var req kafkatopic.ServiceKafkaTopicCreateIn
	diags := expandData(ctx, plan, nil, &req, legacyReqStringFieldsToIntegers)
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
	return util.MergeSlices(diags, legacyFakeRead(ctx, plan, req.Config))
}

func (vw *view) Update(ctx context.Context, plan, state, config *tfModel) diag.Diagnostics {
	rawConfig, diags := stateRawMap(ctx, config)
	if diags.HasError() {
		return diags
	}

	var req kafkatopic.ServiceKafkaTopicUpdateIn
	diags = expandData(
		ctx, plan, state, &req,
		legacyReqRemoveEmptyStrings,
		legacyReqStringFieldsToIntegers,
		legacyReqRemoveComputedOptionalFields(rawConfig),
	)
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
	return util.MergeSlices(diags, legacyFakeRead(ctx, plan, req.Config))
}

func (vw *view) Read(ctx context.Context, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := kafkatopicrepository.New(vw.Client).Read(
		ctx,
		state.Project.ValueString(),
		state.ServiceName.ValueString(),
		state.TopicName.ValueString(),
	)

	if vw.isResource && avngen.IsNotFound(err) && !state.ID.IsNull() {
		// When a Kafka service (without backups) is powered off, all topics and their configurations are deleted.
		// See:
		// - https://aiven.io/docs/platform/concepts/service-power-cycle#power-off-a-service
		// - https://aiven.io/docs/products/kafka/concepts/configuration-backup#how-backups-work
		// - https://github.com/aiven/terraform-provider-aiven/issues/1004
		//
		// This handles two cases:
		// 1. Auto-recreates the topic when the service is powered back on
		// 2. Doesn't create the topic during `terraform import` (checks the ID)
		return util.MergeSlices(diags, vw.Create(ctx, state))
	}

	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	return util.MergeSlices(diags, flattenData(
		ctx, state, rsp,
		lenRspPartitions,
		flattenRspConfigValues(state),
		legacyRspIntegerFieldsToStrings,
		legacyRspAddMissingConfigFields,
	))
}

const ErrTerminationProtectionDelete = "cannot delete the resource because termination_protection is enabled"

func (vw *view) Delete(ctx context.Context, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if state.TerminationProtection.ValueBool() {
		diags.AddError(
			errmsg.SummaryErrorDeletingResource,
			ErrTerminationProtectionDelete,
		)
		return diags
	}

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

	return diags
}

// lenRspPartitions converts "partitions" field from a list of objects to an integer.
// Response has "partitions" field as a list of objects, not an integer.
// https://api.aiven.io/doc/#tag/Service:_Kafka/operation/ServiceKafkaTopicGet
func lenRspPartitions(rsp util.RawMap, in *kafkatopic.ServiceKafkaTopicGetOut) error {
	return rsp.Set(len(in.Partitions), "partitions")
}

// flattenRspConfigValues "flattens" API response values (a list of objects) into a map.
func flattenRspConfigValues(state *tfModel) util.MapModifier[kafkatopic.ServiceKafkaTopicGetOut] {
	return func(rsp util.RawMap, in *kafkatopic.ServiceKafkaTopicGetOut) error {
		if state.Config.IsNull() {
			// The plugin framework doesn't support computed+optional:
			// https://github.com/hashicorp/terraform-plugin-framework/issues/883
			// When the config is not planned/in the state, we remove the whole block.
			// todo: remove in v5.0.0, make it an attribute.
			_ = rsp.Delete("config")
			return nil
		}

		// mapConfig Implements ServiceKafkaTopicGet response in a map to iterate over keys.
		// The response is not just a key-value map, but a map of objects with "value" field:
		// https://api.aiven.io/doc/#tag/Service:_Kafka/operation/ServiceKafkaTopicGet
		var mapCfg struct {
			Config map[string]struct {
				Value any `json:"value"`
			} `json:"config"`
		}

		err := schemautil.Remarshal(in, &mapCfg)
		if err != nil {
			return err
		}

		config := make(map[string]any, len(mapCfg.Config))
		for k, v := range mapCfg.Config {
			config[k] = v.Value
		}
		return rsp.Set(config, "config")
	}
}

// configFields returns a map of all config keys to iterate over.
func configFields() map[string]attr.Type {
	return attrsConfig().AttributeTypes()
}

// stateRawMap turns state into a RawMap.
func stateRawMap(ctx context.Context, state *tfModel) (util.RawMap, diag.Diagnostics) {
	var out map[string]any
	diags := expandData(ctx, state, nil, &out)
	if diags.HasError() {
		return nil, diags
	}

	m, err := util.MarshalRawMap(&out)
	if err != nil {
		diags.AddError(
			"Marshalling Error",
			fmt.Sprintf("failed to marshal state: %s", err.Error()),
		)
		return nil, diags
	}
	return m, nil
}
