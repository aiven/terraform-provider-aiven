package typeupgrader

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestMap is a test for Map.
func TestMap(t *testing.T) {
	type args struct {
		m     map[string]interface{}
		rules map[string]string
	}

	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				m: map[string]interface{}{
					"bool": "true",
					"int":  "1",
				},
				rules: map[string]string{},
			},
			want: map[string]interface{}{
				"bool": "true",
				"int":  "1",
			},
		},
		{
			name: "bool",
			args: args{
				m: map[string]interface{}{
					"bool": "true",
					"int":  "1",
				},
				rules: map[string]string{
					"bool": "bool",
				},
			},
			want: map[string]interface{}{
				"bool": true,
				"int":  "1",
			},
		},
		{
			name: "int",
			args: args{
				m: map[string]interface{}{
					"bool": "true",
					"int":  "1",
				},
				rules: map[string]string{
					"int": "int",
				},
			},
			want: map[string]interface{}{
				"bool": "true",
				"int":  1,
			},
		},
		{
			name: "bool and int",
			args: args{
				m: map[string]interface{}{
					"bool": "true",
					"int":  "1",
				},
				rules: map[string]string{
					"bool": "bool",
					"int":  "int",
				},
			},
			want: map[string]interface{}{
				"bool": true,
				"int":  1,
			},
		},
		{
			name: "complex map",
			args: args{
				m: map[string]interface{}{
					"bool": "true",
					"int":  "1",
					"map": []interface{}{
						map[string]interface{}{
							"bool": "true",
							"int":  "1",
						},
					},
				},
				rules: map[string]string{
					"bool": "bool",
					"int":  "int",
				},
			},
			want: map[string]interface{}{
				"bool": true,
				"int":  1,
				"map": []interface{}{
					map[string]interface{}{
						"bool": "true",
						"int":  "1",
					},
				},
			},
		},
		{
			name: "bool and int with error",
			args: args{
				m: map[string]interface{}{
					"bool": "true",
					"int":  "foo",
				},
				rules: map[string]string{
					"bool": "bool",
					"int":  "int",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Map(tt.args.m, tt.args.rules)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpgradeMap() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && !cmp.Equal(tt.args.m, tt.want) {
				t.Errorf(cmp.Diff(tt.want, tt.args.m))
			}
		})
	}
}

// TestSlice is a test for Slice.
func TestSlice(t *testing.T) {
	type args struct {
		s []interface{}
		t string
	}

	tests := []struct {
		name    string
		args    args
		want    []interface{}
		wantErr bool
	}{
		{
			name: "int",
			args: args{
				s: []interface{}{"1", "3", "3", "7"},
				t: "int",
			},
			want: []interface{}{1, 3, 3, 7},
		},
		{
			name: "int with error",
			args: args{
				s: []interface{}{"1", "foo", "3", "7"},
				t: "int",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Slice(tt.args.s, tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpgradeSlice() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && !cmp.Equal(tt.args.s, tt.want) {
				t.Errorf(cmp.Diff(tt.want, tt.args.s))
			}
		})
	}
}
