package common

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

type foo struct {
	ID  types.String
	Int types.Int64
	Foo types.String
}

type bar struct {
	ID  types.String
	Int types.Int64
	Bar types.String
}

func TestCopy(t *testing.T) {
	from := &foo{
		ID:  types.StringValue("aaa"),
		Int: types.Int64Value(42),
		Foo: types.StringValue("bbb"), // bar doesn't have this field
	}

	dest := &bar{}
	copyBack := Copy(from, dest)
	assert.Equal(t, "aaa", dest.ID.ValueString())
	assert.EqualValues(t, 42, dest.Int.ValueInt64())

	// Changes values and copies back
	dest.ID = types.StringValue("bbb")
	copyBack()
	assert.Equal(t, "bbb", from.ID.ValueString())
}
