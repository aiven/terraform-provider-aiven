package schemautil

import (
	"errors"
	"reflect"
	"testing"
	"time"

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
	assert.Equal(t, "default", PointerValueOrDefault(foo, "default"))
	assert.Equal(t, "bar", PointerValueOrDefault(&bar, "default"))
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
	require.ErrorIs(t, err, errMissingForceNew)
	require.ErrorContains(t, err, "force_new")

	// Expects to set all fields
	d.EXPECT().Set("set", fieldSet).Return(nil)
	d.EXPECT().Set("object", []any{fieldObject}).Return(nil)        // is wrapped into a list
	d.EXPECT().Set("set_of_objects", fieldSetOfObjects).Return(nil) // renamed
	d.EXPECT().Set("force_new", "hey!").Return(nil)                 // set force_new

	err = ResourceDataSet(
		d, dto, s,
		SetForceNew("force_new", "hey!"),
		RenameAlias("to_rename", "set_of_objects"),
	)
	require.NoError(t, err)
}

func TestValidateMaintenanceWindowTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		value        any
		key          string
		expectErrors bool
	}{
		{
			name:         "valid time 00:00:00",
			value:        "00:00:00",
			expectErrors: false,
		},
		{
			name:         "valid time 23:59:59",
			value:        "23:59:59",
			expectErrors: false,
		},
		{
			name:         "valid time 10:30:45",
			value:        "10:30:45",
			expectErrors: false,
		},
		{
			name:         "valid time 01:02:03",
			value:        "01:02:03",
			expectErrors: false,
		},
		{
			name:         "valid time 09:00:00",
			value:        "09:00:00",
			expectErrors: false,
		},
		{
			name:         "invalid single digit hour",
			value:        "9:00:00",
			expectErrors: true,
		},
		{
			name:         "invalid type - not string",
			value:        123,
			expectErrors: true,
		},
		{
			name:         "invalid format - missing seconds",
			value:        "10:30",
			expectErrors: true,
		},
		{
			name:         "invalid format - extra digits",
			value:        "10:30:45:99",
			expectErrors: true,
		},
		{
			name:         "invalid hour - too high",
			value:        "24:00:00",
			expectErrors: true,
		},
		{
			name:         "invalid hour - negative",
			value:        "-1:00:00",
			expectErrors: true,
		},
		{
			name:         "invalid minutes - too high",
			value:        "10:60:00",
			expectErrors: true,
		},
		{
			name:         "invalid seconds - too high",
			value:        "10:30:60",
			expectErrors: true,
		},
		{
			name:         "invalid format - random string",
			value:        "invalid-time",
			expectErrors: true,
		},
		{
			name:         "invalid format - empty string",
			value:        "",
			expectErrors: true,
		},
		{
			name:         "invalid format - only numbers",
			value:        "123456",
			expectErrors: true,
		},
		{
			name:         "invalid format - wrong separators",
			value:        "10-30-00",
			expectErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := ValidateMaintenanceWindowTime(tt.value, "maintenance_window_time")

			if tt.expectErrors {
				assert.NotEmpty(t, errs, "expected validation errors but got none")
			} else {
				assert.Empty(t, errs, "expected no validation errors but got: %v", errs)
			}
		})
	}
}

func FuzzValidateMaintenanceWindowTime(f *testing.F) {
	f.Add("00:00:00")
	f.Add("23:59:59")
	f.Add("09:30:45")
	f.Add("11:11:11")
	f.Add("9:00:00")
	f.Add("24:00:00")
	f.Add("10:60:00")
	f.Add("10:00")
	f.Add("invalid")

	f.Fuzz(func(t *testing.T, input string) {
		_, errs := ValidateMaintenanceWindowTime(input, "test_field")

		// if no errors are returned, the input MUST be a valid parsable time
		if len(errs) == 0 {
			parsed, err := time.Parse("15:04:05", input)
			require.NoError(t, err, "input %q was considered valid by function but failed standard library parsing", input)
			assert.Equal(t, input, parsed.Format("15:04:05"), "input %q was considered valid by function but changed when parsed by stdlib", input)
		}
	})
}
