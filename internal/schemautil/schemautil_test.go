package schemautil

import (
	"errors"
	"reflect"
	"testing"
)

func Test_validateDurationString(t *testing.T) {
	type args struct {
		v interface{}
		k string
	}
	tests := []struct {
		name       string
		args       args
		wantWs     []string
		wantErrors bool
	}{
		{
			"basic",
			args{
				v: "2m",
				k: "",
			},
			nil,
			false,
		},
		{
			"wrong-duration",
			args{
				v: "123qweert",
				k: "",
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWs, gotErrors := ValidateDurationString(tt.args.v, tt.args.k)
			if !reflect.DeepEqual(gotWs, tt.wantWs) {
				t.Errorf("validateDurationString() gotWs = %v, want %v", gotWs, tt.wantWs)
			}
			if !(tt.wantErrors == (len(gotErrors) > 0)) {
				t.Errorf("validateDurationString() gotErrors = %v", gotErrors)
			}
		})
	}
}

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
