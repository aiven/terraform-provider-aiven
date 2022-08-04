package vpc

import (
	"errors"
	"reflect"
	"testing"
)

func Test_validateVPCID(t *testing.T) {
	type args struct {
		i interface{}
		k string
	}

	tests := []struct {
		name         string
		args         args
		wantWarnings []string
		wantErrors   []error
	}{
		{
			name:         "basic",
			args:         args{"project/vpc", "vpc_id"},
			wantWarnings: nil,
			wantErrors:   nil,
		},
		{
			name:         "wrong id",
			args:         args{"wrong", "vpc_id"},
			wantWarnings: nil,
			wantErrors:   []error{errors.New("invalid vpc_id, expected <project_name>/<vpc_id>")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWarnings, gotErrors := validateVPCID(tt.args.i, tt.args.k)
			if !reflect.DeepEqual(gotWarnings, tt.wantWarnings) {
				t.Errorf("validateVPCID() gotWarnings = %v, want %v", gotWarnings, tt.wantWarnings)
			}
			if !reflect.DeepEqual(gotErrors, tt.wantErrors) {
				t.Errorf("validateVPCID() gotErrors = %v, want %v", gotErrors, tt.wantErrors)
			}
		})
	}
}
