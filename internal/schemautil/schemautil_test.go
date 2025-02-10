package schemautil

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/mocks"
)

func Test_splitResourceID(t *testing.T) {
	type args struct {
		resourceID string
		n          int
	}
	tests := []struct {
		name      string
		args      args
		want      []string
		wantError error
	}{
		{
			"basic",
			args{
				resourceID: "test/identifier",
				n:          2,
			},
			[]string{"test", "identifier"},
			nil,
		},
		{
			"invalid",
			args{
				resourceID: "invalid",
				n:          2,
			},
			nil,
			errors.New("invalid resource id: invalid"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotError := SplitResourceID(tt.args.resourceID, tt.args.n)
			if tt.wantError != nil {
				if !reflect.DeepEqual(gotError, tt.wantError) {
					t.Errorf("splitResourceID() gotError = %v, want %v", gotError, tt.wantError)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitResourceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_PointerValueOrDefault(t *testing.T) {
	var foo *string
	bar := "bar"
	assert.Equal(t, PointerValueOrDefault(foo, "default"), "default")
	assert.Equal(t, PointerValueOrDefault(&bar, "default"), "bar")
}

func TestResourceDataSetAddForceNew(t *testing.T) {
	s := map[string]*schema.Schema{
		"set": {
			Elem:     &schema.Schema{Type: schema.TypeString},
			Optional: true,
			Type:     schema.TypeSet,
		},
		"object": {
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"string": {Optional: true, Type: schema.TypeString},
			}},
			Optional: true,
			Type:     schema.TypeList,
		},
		"set_of_objects": {
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"description": {Optional: true, Type: schema.TypeString},
			}},
			Optional: true,
			Type:     schema.TypeSet,
		},
		"optional_force_new": {
			Type:     schema.TypeString,
			ForceNew: true,
			Optional: true, // ignored
		},
		"force_new": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
	}

	fieldSet := []any{"a", "b"}
	fieldObject := map[string]any{"string": "a"}
	fieldSetOfObjects := []any{
		map[string]any{"description": "a"},
	}
	dto := map[string]any{
		"set":       fieldSet,
		"object":    fieldObject,
		"to_rename": fieldSetOfObjects, // has different name in schema
		"to_ignore": "will be ignored", // not in schema
	}

	// Missing force_new field
	d := mocks.NewMockResourceData(t)
	err := ResourceDataSet(d, dto, s)
	require.Error(t, err, fmt.Errorf("%w: %q", errMissingForceNew, "force_new"))

	// Expects to set all fields
	d.EXPECT().Set("set", fieldSet).Return(nil)
	d.EXPECT().Set("object", []any{fieldObject}).Return(nil)        // is wrapped into a list
	d.EXPECT().Set("set_of_objects", fieldSetOfObjects).Return(nil) // renamed
	d.EXPECT().Set("force_new", "hey!").Return(nil)                 // set force_new

	err = ResourceDataSet(
		d, dto, s,
		AddForceNew("force_new", "hey!"),
		RenameAlias("to_rename", "set_of_objects"),
	)
	require.NoError(t, err)
}
