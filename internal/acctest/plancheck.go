package acctest

import (
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

type attributeChangeCheck struct {
	resourceAddr string
	attrs        []string
}

func ExpectOnlyAttributesChanged(resourceAddr string, attrs ...string) plancheck.PlanCheck {
	return &attributeChangeCheck{
		resourceAddr: resourceAddr,
		attrs:        attrs,
	}
}

func (c *attributeChangeCheck) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	var targetResource *tfjson.ResourceChange

	// Find our resource in the changes
	for _, rc := range req.Plan.ResourceChanges {
		if rc.Address == c.resourceAddr {
			targetResource = rc
			break
		}
	}

	if targetResource == nil {
		resp.Error = fmt.Errorf("resource %s not found in plan", c.resourceAddr)
		return
	}

	if targetResource.Change == nil {
		resp.Error = fmt.Errorf("no changes found for resource %s", c.resourceAddr)
		return
	}

	before, ok := targetResource.Change.Before.(map[string]interface{})
	if !ok {
		resp.Error = fmt.Errorf("before state for resource %s is not a map", c.resourceAddr)
		return
	}

	after, ok := targetResource.Change.After.(map[string]interface{})
	if !ok {
		resp.Error = fmt.Errorf("after state for resource %s is not a map", c.resourceAddr)
		return
	}

	// Create a set of expected changes
	expectedChanges := make(map[string]struct{})
	for _, attr := range c.attrs {
		expectedChanges[attr] = struct{}{}
	}

	// Check all attributes in the after state
	for key, afterValue := range after {
		beforeValue, existsInBefore := before[key]

		// If value changed
		if !existsInBefore || !cmp.Equal(beforeValue, afterValue) {
			// Check if this change was expected
			if _, expected := expectedChanges[key]; !expected {
				if afterValue == nil && beforeValue != nil {
					continue
				}

				diff := cmp.Diff(beforeValue, afterValue)
				resp.Error = fmt.Errorf(
					"unexpected change in attribute %q for resource %s:\n%s",
					key,
					c.resourceAddr,
					diff,
				)
				return
			}
			// Remove from expected changes as we found it
			delete(expectedChanges, key)
		}
	}

	for key := range expectedChanges {
		if beforeValue, exists := before[key]; exists {
			if _, existsInAfter := after[key]; !existsInAfter && beforeValue != nil {
				// If the value existed in before but not in after, consider it changed
				delete(expectedChanges, key)
			}
		}
	}

	// Check if all expected changes were found
	if len(expectedChanges) > 0 {
		remaining := make([]string, 0, len(expectedChanges))
		for attr := range expectedChanges {
			remaining = append(remaining, attr)
		}
		resp.Error = fmt.Errorf(
			"expected changes in attributes %v for resource %s were not found in plan",
			remaining,
			c.resourceAddr,
		)
	}
}
