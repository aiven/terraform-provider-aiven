package template

import (
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// normalizeHCL uses the official HCL library to parse and format HCL with consistent formatting
func normalizeHCL(input string) string {
	f, diags := hclwrite.ParseConfig([]byte(input), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return input // Return original if there are parsing errors
	}

	// Create a new file with consistent formatting
	newFile := hclwrite.NewEmptyFile()
	sortAndRewriteBody(f.Body(), newFile.Body())

	return string(newFile.Bytes())
}

// sortAndRewriteBody sorts and copies all attributes and blocks from src to dst with proper indentation
func sortAndRewriteBody(src, dst *hclwrite.Body) {
	// Get all attributes and sort them
	attrs := src.Attributes()
	attrNames := make([]string, 0, len(attrs))
	for name := range attrs {
		attrNames = append(attrNames, name)
	}
	sort.Strings(attrNames)

	// Write sorted attributes
	for _, name := range attrNames {
		dst.SetAttributeRaw(name, attrs[name].Expr().BuildTokens(nil))
	}

	// Get all blocks and sort them
	blocks := src.Blocks()
	sort.Slice(blocks, func(i, j int) bool {
		if blocks[i].Type() != blocks[j].Type() {
			return blocks[i].Type() < blocks[j].Type()
		}
		// If types are equal, compare labels
		iLabels := blocks[i].Labels()
		jLabels := blocks[j].Labels()
		for k := 0; k < len(iLabels) && k < len(jLabels); k++ {
			if iLabels[k] != jLabels[k] {
				return iLabels[k] < jLabels[k]
			}
		}
		return len(iLabels) < len(jLabels)
	})

	// Write sorted blocks with proper indentation
	for _, block := range blocks {
		newBlock := dst.AppendNewBlock(block.Type(), block.Labels())
		sortAndRewriteBody(block.Body(), newBlock.Body())
	}
}
