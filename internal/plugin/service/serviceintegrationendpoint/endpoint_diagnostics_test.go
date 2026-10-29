package serviceintegrationendpoint_test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func testAccCheckEndpointUserConfig(resourceName string, expected map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id, err := endpointResourceID(s, resourceName)
		if err != nil {
			return err
		}
		return checkEndpointAPIUserConfig(id, expected)
	}
}

func endpointResourceID(s *terraform.State, resourceName string) (string, error) {
	if s == nil {
		return "", fmt.Errorf("state is unavailable for %s", resourceName)
	}
	resourceState, ok := s.RootModule().Resources[resourceName]
	if !ok || resourceState.Primary == nil || resourceState.Primary.ID == "" {
		return "", fmt.Errorf("resource %s has no ID in state", resourceName)
	}
	return resourceState.Primary.ID, nil
}

func checkEndpointAPIUserConfig(resourceID string, expected map[string]string) error {
	project, endpointID, err := schemautil.SplitResourceID2(resourceID)
	if err != nil {
		return err
	}
	client, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}
	endpoint, err := client.ServiceIntegrationEndpointGet(context.Background(), project, endpointID, [2]string{"include_secrets", "true"})
	if err != nil {
		return fmt.Errorf("endpoint API GET (include_secrets=true): %w", err)
	}
	mismatches := make([]error, 0, len(expected))
	for _, key := range slices.Sorted(maps.Keys(expected)) {
		value := endpointMapValue(endpoint.UserConfig, key)
		if value == expected[key] {
			continue
		}
		mismatches = append(mismatches, fmt.Errorf(
			"endpoint API GET (include_secrets=true): user_config[%q]: got %s; expected configured string (value omitted)",
			key, endpointValueDescription(value),
		))
	}
	return errors.Join(mismatches...)
}

// endpointPlanDiagnostics logs both refresh drift and the proposed actions. It does not
// change the test outcome; the following plan checks still enforce the expected behavior.
type endpointPlanDiagnostics struct {
	t            *testing.T
	resourceName string
}

func (d endpointPlanDiagnostics) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, _ *plancheck.CheckPlanResponse) {
	d.t.Helper()
	if req.Plan == nil {
		d.t.Log("endpoint plan diagnostics: plan is unavailable")
		return
	}
	for _, phase := range []struct {
		name    string
		changes []*tfjson.ResourceChange
	}{
		{name: "refresh drift", changes: req.Plan.ResourceDrift},
		{name: "planned change", changes: req.Plan.ResourceChanges},
	} {
		for _, change := range phase.changes {
			if change == nil || change.Address != d.resourceName || change.Change == nil {
				continue
			}
			d.t.Logf("%s: %s actions=%v", phase.name, change.Address, change.Change.Actions)
			logEndpointPlanChanges(d.t, change.Address, change.Change.Before, change.Change.After, change.Change.AfterUnknown)
		}
	}
}

// Values are compared before rendering, so even two different nonempty strings are
// reported as a change. Their contents are never printed, regardless of sensitivity flags.
func logEndpointPlanChanges(t *testing.T, path string, before, after, unknown any) {
	t.Helper()
	if unknown == true {
		t.Logf("%s: %s -> unknown", path, endpointValueDescription(before))
		return
	}
	beforeMap, beforeObject := before.(map[string]any)
	afterMap, afterObject := after.(map[string]any)
	unknownMap, _ := unknown.(map[string]any)
	if beforeObject || afterObject {
		keys := make(map[string]struct{})
		for _, object := range []map[string]any{beforeMap, afterMap, unknownMap} {
			for key := range object {
				keys[key] = struct{}{}
			}
		}
		if len(keys) > 0 {
			for _, key := range slices.Sorted(maps.Keys(keys)) {
				logEndpointPlanChanges(t, path+"."+key, endpointMapValue(beforeMap, key), endpointMapValue(afterMap, key), unknownMap[key])
			}
			return
		}
	}
	beforeList, beforeArray := before.([]any)
	afterList, afterArray := after.([]any)
	unknownList, _ := unknown.([]any)
	if beforeArray || afterArray {
		if beforeArray != afterArray || len(beforeList) != len(afterList) {
			t.Logf("%s: %s -> %s", path, endpointValueDescription(before), endpointValueDescription(after))
		}
		for i := range max(len(beforeList), len(afterList), len(unknownList)) {
			var itemUnknown any
			if i < len(unknownList) {
				itemUnknown = unknownList[i]
			}
			logEndpointPlanChanges(t, fmt.Sprintf("%s[%d]", path, i), endpointListValue(beforeList, i), endpointListValue(afterList, i), itemUnknown)
		}
		return
	}
	if !reflect.DeepEqual(before, after) {
		t.Logf("%s: %s -> %s", path, endpointValueDescription(before), endpointValueDescription(after))
	}
}

type endpointAbsentValue struct{}

func endpointMapValue(values map[string]any, key string) any {
	if value, ok := values[key]; ok {
		return value
	}
	return endpointAbsentValue{}
}

func endpointListValue(values []any, index int) any {
	if index < len(values) {
		return values[index]
	}
	return endpointAbsentValue{}
}

func endpointValueDescription(value any) string {
	switch v := value.(type) {
	case endpointAbsentValue:
		return "absent"
	case nil:
		return "null"
	case string:
		switch {
		case v == "":
			return `"" (empty string)`
		case strings.Contains(v, "<redacted>") || strings.Trim(v, "*") == "":
			return "redacted string"
		default:
			return "non-empty string (value omitted)"
		}
	case []any:
		return fmt.Sprintf("list (%d elements)", len(v))
	case map[string]any:
		return fmt.Sprintf("object (%d keys)", len(v))
	default:
		return fmt.Sprintf("%T (value omitted)", value)
	}
}
