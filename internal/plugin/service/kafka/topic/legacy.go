package topic

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aiven/go-client-codegen/handler/kafkatopic"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

// legacyStringFields contains fields that are strings in our schema but integers in the API.
// SDKv2 didn't distinguish between empty and zero values: `0 == 0`.
// This is likely why these fields are strings, e.g. `"" != 0`.
// todo: remove in v5.0.0. Make these fields int64.
// Terraform can convert strings to int64 automatically.
// We could probably change the field types without a major version bump.
// https://developer.hashicorp.com/terraform/language/expressions/types#type-conversion
func legacyStringFields() map[string]bool {
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

// legacyReqCastConfigLegacyFields converts topic config legacy fields (strings) to integers. See legacyStringFields.
// todo: remove in v5.0.0
func legacyReqCastConfigLegacyFields[T any](req util.RawMap, _ *T) error {
	for k := range legacyStringFields() {
		if req.Exists("config", k) {
			err := func() error {
				s, _ := req.GetString("config", k)
				i, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return err
				}
				return req.Set(i, "config", k)
			}()
			if err != nil {
				return fmt.Errorf("failed to modify request config key %q: %w", k, err)
			}
		}
	}
	return nil
}

// legacyRspCastConfigLegacyFields converts legacy config fields to strings. See legacyStringFields.
// todo: remove in v5.0.0
func legacyRspCastConfigLegacyFields(rsp util.RawMap, _ *kafkatopic.ServiceKafkaTopicGetOut) error {
	var errM *multierror.Error
	for k := range legacyStringFields() {
		// The API returns integers, but our schema has strings.
		// GetString() returns a string, and then we just set it back.
		v, err := rsp.GetString("config", k)
		if err == nil {
			errM = multierror.Append(errM, rsp.Set(v, "config", k))
		}
	}
	return errM.ErrorOrNil()
}

// legacyZeroConfig returns a map of all config keys with zero Go values.
// The Plugin Framework requires all computed+optional fields to be set in the state.
// We fill the state with zero values to:
// - satisfy this requirement when the backend doesn't return a value for a field
// - suppress the diff with SDKv2 state that always had all fields set.
// todo: remove in v5.0.0. Make all fields optional.
func legacyZeroConfig() map[string]any {
	values := make(map[string]any)
	for k, v := range configFields() {
		switch v {
		case types.StringType:
			values[k] = ""
		case types.BoolType:
			values[k] = false
		case types.Int64Type, types.Float64Type:
			values[k] = 0
		default:
			// This is fatal and must fail during development.
			panic(fmt.Sprintf("unsupported config attribute type %T for key %q", v, k))
		}
	}
	return values
}

// legacyFakeRead fakes the Read() operation by filling missing config fields with zero values.
// This happened automatically in SDKv2 version. See legacyZeroConfig.
// todo: remove in v5.0.0. Make all config fields optional.
func legacyFakeRead(ctx context.Context, plan *tfModel) diag.Diagnostics {
	if plan.Config.IsNull() {
		// The plugin framework doesn't support optional+computed:
		// https://discuss.hashicorp.com/t/optional-computed-block-handling-in-plugin-framework/56337
		return nil
	}

	// Reads the state and adds missing config keys with zero values.
	rawPlan, diags := stateRawMap(ctx, plan)
	if diags.HasError() {
		return diags
	}

	var fakeRsp map[string]any
	return flattenData(ctx, plan, &fakeRsp, func(state util.RawMap, _ *map[string]any) error {
		for k, v := range legacyZeroConfig() {
			if !rawPlan.Exists("config", k) {
				err := state.Set(v, "config", k)
				if err != nil {
					return fmt.Errorf("%s: %w", k, err)
				}
			}
		}
		return nil
	})
}

// legacyReqRemoveComputedOptionalFields removes "computed+optional" fields from the topic config POST request.
// When a Kafka topic is created, the backend assigns default configuration values.
// We set them in the state because computed fields _must_ have values.
// If these computed values are sent back in subsequent API requests, the backend will
// incorrectly mark them as user-defined values. To prevent this, we only keep configuration
// values that were explicitly set in the Terraform config (.tf files) and remove all
// "computed+optional" fields from the request.
// For the Update handler only!
// todo: remove in v5.0.0. See legacyFakeRead.
func legacyReqRemoveComputedOptionalFields(rawConfig util.RawMap) util.MapModifier[apiModel] {
	return func(req util.RawMap, in *apiModel) error {
		var err *multierror.Error
		for k := range configFields() {
			if !rawConfig.Exists("config", k) {
				err = multierror.Append(err, req.Delete("config", k))
			}
		}
		return err.ErrorOrNil()
	}
}

// keyOwnerUserGroupID
// SDKv2 didn't distinguish between empty and null values.
// Therefore, "owner_user_group_id" might be an empty string in the legacy SDKv2 state during migration.
const keyOwnerUserGroupID = "owner_user_group_id"

// legacyReqOwnerUserGroupID removes empty strings that come from the state because the API doesn't accept them.
// todo: remove in v5.0.0. Don't set empty string back in legacyRspOwnerUserGroupID.
func legacyReqOwnerUserGroupID(state *tfModel) util.MapModifier[apiModel] {
	return func(req util.RawMap, _ *apiModel) error {
		if !state.OwnerUserGroupID.IsNull() && state.OwnerUserGroupID.ValueString() == "" {
			return req.Delete(keyOwnerUserGroupID)
		}
		return nil
	}
}

// legacyRspOwnerUserGroupID copies empty strings from the old state back to the new state.
// Due to a bug in Terraform, empty strings are hidden in the plan output:
// https://discuss.hashicorp.com/t/framework-migration-test-produces-non-empty-plan/54523/12
// Every customer migrating from SDKv2 version would see a diff _without_ any change.
// We "emulate" the old behavior to suppress the diff.
// todo: remove in v5.0.0
func legacyRspOwnerUserGroupID(oldState *tfModel) util.MapModifier[kafkatopic.ServiceKafkaTopicGetOut] {
	return func(newState util.RawMap, _ *kafkatopic.ServiceKafkaTopicGetOut) error {
		owner, ok := oldState.OwnerUserGroupID.ValueString(), !oldState.OwnerUserGroupID.IsNull()
		if ok && owner == "" {
			return newState.Set(owner, keyOwnerUserGroupID)
		}
		return nil
	}
}

// legacyRspTerminationProtection fakes the field in the response to suppress diffs.
// todo: remove in v5.0.0. This field never existed in the API.
// Use instead: https://developer.hashicorp.com/terraform/tutorials/state/resource-lifecycle#prevent-resource-deletion
func legacyRspTerminationProtection[T any](rsp util.RawMap, _ *T) error {
	// This is not a real field, fakes it, so TF won't complain about state drift.
	v, err := rsp.GetBool("termination_protection")
	if err != nil {
		return err
	}
	return rsp.Set(v, "termination_protection")
}
