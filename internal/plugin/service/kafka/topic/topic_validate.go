package topic

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/kafkatopicrepository"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

// validateAlreadyExists Conflict validation for new topics on `terraform plan` stage
func (vw *view) validateAlreadyExists(ctx context.Context, config *tfModel) diag.Diagnostics {
	if config.ID.IsNull() {
		return nil
	}

	exists, err := kafkatopicrepository.New(vw.Client).Exists(
		ctx,
		config.Project.ValueString(),
		config.ServiceName.ValueString(),
		config.TopicName.ValueString(),
	)

	if exists {
		err = fmt.Errorf("topic already exists")
	}

	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(errmsg.SummaryInvalidConfiguration, err.Error())
		return diags
	}
	return nil
}

// validateTopicConfig validates Kafka topic configuration values by ensuring that local retention settings
// do not exceed their corresponding global retention settings.
// For example, local_retention_ms >= retention_ms.
// While this validation is also performed on the backend, we check it here to provide faster feedback
// to users during the planning phase.
func (vw *view) validateTopicConfig(ctx context.Context, config *tfModel) diag.Diagnostics {
	if config.Config.IsNull() {
		return nil
	}

	rawConfig, diags := stateTopicConfig(ctx, config, errmsg.SummaryInvalidConfiguration)
	if rawConfig == nil || diags.HasError() {
		return diags
	}

	const prefix = "local_"
	for localKey := range emptyConfig() {
		if !strings.HasPrefix(localKey, prefix) {
			continue
		}

		// There is a validation ensuring that local_* keys come with their remote counterparts.
		remoteKey := strings.TrimPrefix(localKey, prefix)
		localVal, _ := rawConfig.GetString("config", localKey)
		remoteVal, _ := rawConfig.GetString("config", remoteKey)
		if greaterLocalNumber(localVal, remoteVal) {
			diags.AddError(
				errmsg.SummaryInvalidConfiguration,
				fmt.Sprintf("%s cannot be greater than %s", localKey, remoteKey),
			)
		}
	}
	return diags
}

func greaterLocalNumber(local, remote string) bool {
	aInt, aErr := strconv.ParseInt(local, 10, 64)
	if aErr == nil {
		bInt, _ := strconv.ParseInt(remote, 10, 64)
		return aInt > bInt
	}

	aFloat, aErr := strconv.ParseFloat(local, 64)
	if aErr == nil {
		bFloat, _ := strconv.ParseFloat(remote, 64)
		return aFloat > bFloat
	}
	return false
}

// stateTopicConfig turns the config block into a util.RawMap.
func stateTopicConfig(ctx context.Context, state *tfModel, summaryErr string) (util.RawMap, diag.Diagnostics) {
	if state.Config.IsNull() {
		return nil, nil
	}

	config, diags := util.ExpandSingleNested(ctx, expandConfig, state.Config)
	if diags.HasError() {
		return nil, diags
	}

	b, err := json.Marshal(config)
	if err != nil {
		diags.AddError(summaryErr, err.Error())
		return nil, diags
	}

	return util.NewRawMap(b), nil
}
