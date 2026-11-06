package schemautil

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ShouldNotBeEmpty(_ context.Context, _, newValue, _ any) bool {
	switch t := newValue.(type) {
	case string:
		return t != ""
	case []any:
		return len(t) != 0
	case *schema.Set:
		return t.Len() != 0
	default:
		panic(fmt.Sprintf("unexpected type: %T", t))
	}
}

func CustomizeDiffCheckUniqueTag(_ context.Context, d *schema.ResourceDiff, _ any) error {
	t := make(map[string]bool)
	for _, tag := range d.Get("tag").(*schema.Set).List() {
		tagVal := tag.(map[string]any)
		k := tagVal["key"].(string)
		if t[k] {
			return fmt.Errorf("tag keys should be unique, duplicate with the key: %s", k)
		}
		t[k] = true
	}

	return nil
}

var ErrAutoscalerConflict = errors.New("autoscaler integration is enabled, additional_disk_space cannot be set")
