package main

import (
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestExampleJSONTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		val  cty.Value
		want string // empty when the value isn't rendered as jsonencode
	}{
		{
			name: "object",
			val:  cty.StringVal(`{"type": "record", "fields": [{"name": "id"}]}`),
			want: "jsonencode({\n  fields = [{\n    name = \"id\"\n  }]\n  type = \"record\"\n})",
		},
		{name: "array", val: cty.StringVal(`["a"]`)},
		{name: "invalid JSON", val: cty.StringVal(`{"type": `)},
		{name: "plain string", val: cty.StringVal("foo")},
		{name: "number", val: cty.NumberIntVal(1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tokens, ok := exampleJSONTokens(tt.val)
			assert.Equal(t, tt.want != "", ok)
			if ok {
				assert.Equal(t, tt.want, string(hclwrite.Format(tokens.Bytes())))
			}
		})
	}
}
