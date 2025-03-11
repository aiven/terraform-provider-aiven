package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplatePathExpression(t *testing.T) {
	t.Parallel()

	// Test empty path
	emptyPath := TemplatePath{}
	assert.Equal(t, "", emptyPath.Expression())

	// Test simple path
	simplePath := NewTemplatePath("field", false)
	assert.Equal(t, ".field", simplePath.Expression())

	// Test nested path with regular field
	nestedPath := NewTemplatePath("parent", false)
	nestedPath = nestedPath.AppendField("child", false)
	assert.Equal(t, "(index .parent \"child\")", nestedPath.Expression())

	// Test nested path with collection
	collectionPath := NewTemplatePath("items", true)
	collectionPath = collectionPath.AppendField("name", false)
	assert.Equal(t, "(index .items 0 \"name\")", collectionPath.Expression())

	// Test complex nested path with mixed collections
	complexPath := NewTemplatePath("parent", true)
	complexPath = complexPath.AppendField("child", false)
	complexPath = complexPath.AppendField("grandchild", true)

	// Should generate a complex path expression for a collection within a collection
	expr := complexPath.Expression()
	assert.Contains(t, expr, "index")
	assert.Contains(t, expr, "parent")
	assert.Contains(t, expr, "child")
	assert.Contains(t, expr, "grandchild")
	assert.Equal(t, "(index (index .parent 0 \"child\") \"grandchild\")", expr)
}
