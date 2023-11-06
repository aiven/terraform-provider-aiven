package typeupgrader

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestMap is a test for Map.
func TestMap(t *testing.T) {
	type args struct {
		mapInput  map[string]any
		typeRules map[string]string
	}

	tests := []struct {
		name    string
		args    args
		want    map[string]any
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				mapInput: map[string]any{
					"bool": "true",
					"int":  "1",
				},
				typeRules: map[string]string{},
			},
			want: map[string]any{
				"bool": "true",
				"int":  "1",
			},
		},
		{
			name: "bool",
			args: args{
				mapInput: map[string]any{
					"bool": "true",
					"int":  "1",
				},
				typeRules: map[string]string{
					"bool": "bool",
				},
			},
			want: map[string]any{
				"bool": true,
				"int":  "1",
			},
		},
		{
			name: "int",
			args: args{
				mapInput: map[string]any{
					"bool": "true",
					"int":  "1",
				},
				typeRules: map[string]string{
					"int": "int",
				},
			},
			want: map[string]any{
				"bool": "true",
				"int":  1,
			},
		},
		{
			name: "bool and int",
			args: args{
				mapInput: map[string]any{
					"bool": "true",
					"int":  "1",
				},
				typeRules: map[string]string{
					"bool": "bool",
					"int":  "int",
				},
			},
			want: map[string]any{
				"bool": true,
				"int":  1,
			},
		},
		{
			name: "complex map",
			args: args{
				mapInput: map[string]any{
					"bool": "true",
					"int":  "1",
					"map": []any{
						map[string]any{
							"bool": "true",
							"int":  "1",
						},
					},
				},
				typeRules: map[string]string{
					"bool": "bool",
					"int":  "int",
				},
			},
			want: map[string]any{
				"bool": true,
				"int":  1,
				"map": []any{
					map[string]any{
						"bool": "true",
						"int":  "1",
					},
				},
			},
		},
		{
			name: "bool and int with error",
			args: args{
				mapInput: map[string]any{
					"bool": "true",
					"int":  "foo",
				},
				typeRules: map[string]string{
					"bool": "bool",
					"int":  "int",
				},
			},
			wantErr: true,
		},
		{
			name: "unknown type",
			args: args{
				mapInput: map[string]any{
					"bool": "true",
					"int":  "1",
				},
				typeRules: map[string]string{
					"bool": "bool",
					"int":  "foo",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Map(tt.args.mapInput, tt.args.typeRules)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpgradeMap() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && !cmp.Equal(tt.args.mapInput, tt.want) {
				t.Errorf(cmp.Diff(tt.want, tt.args.mapInput))
			}
		})
	}
}

// TestSlice is a test for Slice.
func TestSlice(t *testing.T) {
	type args struct {
		sliceInput []any
		typeInput  string
	}

	tests := []struct {
		name    string
		args    args
		want    []any
		wantErr bool
	}{
		{
			name: "int",
			args: args{
				sliceInput: []any{"1", "3", "3", "7"},
				typeInput:  "int",
			},
			want: []any{1, 3, 3, 7},
		},
		{
			name: "int with error",
			args: args{
				sliceInput: []any{"1", "foo", "3", "7"},
				typeInput:  "int",
			},
			wantErr: true,
		},
		{
			name: "unknown type",
			args: args{
				sliceInput: []any{"1", "foo", "3", "7"},
				typeInput:  "foo",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Slice(tt.args.sliceInput, tt.args.typeInput)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpgradeSlice() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && !cmp.Equal(tt.args.sliceInput, tt.want) {
				t.Errorf(cmp.Diff(tt.want, tt.args.sliceInput))
			}
		})
	}
}
