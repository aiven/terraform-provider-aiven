package topic

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/kafkatopicrepository"
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

	// Turns TF value into RawMap
	raw, diags := stateRawMap(ctx, config)
	if diags.HasError() {
		return diags
	}

	// Converts legacy strings into integers
	// todo: use integers in v5.0.0
	_ = legacyReqStringFieldsToIntegers(raw, config)

	// Compares type int only, because float comparison is tricky.
	const prefix = "local_"
	for lKey := range configFields() {
		if strings.HasPrefix(lKey, prefix) {
			rKey := strings.TrimPrefix(lKey, prefix)
			l, lErr := raw.GetInt("config", lKey)
			r, rErr := raw.GetInt("config", rKey)
			if l > r && lErr == nil && rErr == nil {
				diags.AddError(
					errmsg.SummaryInvalidConfiguration,
					fmt.Sprintf("%s cannot be greater than %s", lKey, rKey),
				)
			}
		}
	}
	return diags
}
