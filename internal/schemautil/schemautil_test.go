package schemautil

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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
