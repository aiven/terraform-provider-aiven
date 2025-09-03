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

// legacyReqStringFieldsToIntegers converts topic config legacy fields (strings) to integers. See legacyStringFields.
// todo: remove in v5.0.0
func legacyReqStringFieldsToIntegers[T any](req util.RawMap, _ *T) error {
	var errM *multierror.Error
	for k := range legacyStringFields() {
		s, err := req.GetString("config", k)
		switch {
		case util.IsKeyNotFound(err):
			continue
		case err == nil:
			var i int64
			i, err = strconv.ParseInt(s, 10, 64)
			if err == nil {
				err = req.Set(i, "config", k)
			}
		}
		if err != nil {
			errM = multierror.Append(errM, fmt.Errorf("failed to cast to int %q: %w", k, err))
		}
	}
	return nil
}

// legacyRspIntegerFieldsToStrings converts legacy config fields (integers) to strings. See legacyStringFields.
// todo: remove in v5.0.0
func legacyRspIntegerFieldsToStrings[T any](rsp util.RawMap, _ *T) error {
	var errM *multierror.Error
	for k := range legacyStringFields() {
		// The API returns integers, but our schema wants strings.
		v, err := rsp.GetInt("config", k)
		switch {
		case util.IsKeyNotFound(err):
			continue
		case err == nil:
			err = rsp.Set(fmt.Sprint(v), "config", k)
		}

		if err != nil {
			errM = multierror.Append(errM, fmt.Errorf("failed to cast to string %q: %w", k, err))
		}
	}
	return errM.ErrorOrNil()
}

// legacyZeroConfig returns a map of all config keys with zero Go values.
// The Plugin Framework requires all computed+optional fields to be set in the state.
// We fill the state with zero values to:
// - satisfy this requirement when the backend doesn't return a value for a field
// - suppress the diff with SDKv2 state that always had all fields set.
// todo: remove in v5.0.0. Make all fields optional, not computed+optional.
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

// legacyRspAddMissingConfigFields fills missing config fields with zero values.
// See legacyZeroConfig.
// todo: remove in v5.0.0
func legacyRspAddMissingConfigFields[T any](rsp util.RawMap, _ *T) error {
	if !rsp.Exists("config") {
		return nil
	}

	var errM *multierror.Error
	for k, v := range legacyZeroConfig() {
		if !rsp.Exists("config", k) {
			err := rsp.Set(v, "config", k)
			if err != nil {
				errM = multierror.Append(errM, fmt.Errorf("failed to set zero value %q: %w", k, err))
			}
		}
	}
	return errM.ErrorOrNil()
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
		var errM *multierror.Error
		for k := range configFields() {
			if !rawConfig.Exists("config", k) {
				err := req.Delete("config", k)
				if err != nil {
					errM = multierror.Append(errM, fmt.Errorf("failed to remove field %q: %w", k, err))
				}
			}
		}
		return errM.ErrorOrNil()
	}
}

// legacyReqRemoveEmptyStrings removes empty strings from the request.
// - SDKv2 state previously stored empty strings instead of null values
// - Terraform has a known issue comparing null and empty strings (https://discuss.hashicorp.com/t/54523)
// TODO: Remove in v5.0.0.
func legacyReqRemoveEmptyStrings(req util.RawMap, _ *apiModel) error {
	empties := []string{
		"owner_user_group_id",
		"topic_description",
	}

	var errM *multierror.Error
	for _, key := range empties {
		if v, err := req.GetString(key); err == nil && v == "" {
			errM = multierror.Append(errM, req.Delete(key))
		}
	}
	return errM.ErrorOrNil()
}

// legacyFakeRsp fakes the Read() operation by filling missing config fields with zero values.
// This happened automatically in SDKv2 version. See legacyZeroConfig.
// todo: remove in v5.0.0. Make all config fields optional.
func legacyFakeRead(ctx context.Context, plan *tfModel, config *kafkatopic.ConfigIn) diag.Diagnostics {
	if config == nil {
		return nil
	}

	type legacyFakeRsp struct {
		Config *kafkatopic.ConfigIn `json:"config"`
	}

	fakeRsp := &legacyFakeRsp{Config: config}
	return flattenData(ctx, plan, fakeRsp, legacyRspIntegerFieldsToStrings, legacyRspAddMissingConfigFields)
}
