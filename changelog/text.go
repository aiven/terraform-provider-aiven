package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

func hasBeta(description string) bool {
	return strings.Contains(description, util.AivenEnableBeta)
}

var reEnum = regexp.MustCompile("(?i)enum: `.+`\\.?\\s*")

// removeEnum removes enum values from the description to keep it brief
func removeEnum(text string) string {
	return reEnum.ReplaceAllString(text, "")
}

var reCode = regexp.MustCompile("`[^`]+`")

func findEnums(description string) []string {
	parts := strings.Split(description, userconfig.PossibleValuesPrefix)
	if len(parts) != 2 {
		return nil
	}

	return reCode.FindAllString(parts[1], -1)
}

// strValue formats Go value into humanreadable string
func strValue(src any) string {
	switch v := src.(type) {
	case string:
		return v
	case []string:
		return strings.Join(v, ", ")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// strValueType returns the string representation of the schema.ValueType
func strValueType(t schema.ValueType) string {
	switch t {
	case schema.TypeBool:
		return "bool"
	case schema.TypeString:
		return "string"
	case schema.TypeInt:
		return "int"
	case schema.TypeFloat:
		return "float"
	case schema.TypeList:
		return "list"
	case schema.TypeMap:
		return "map"
	case schema.TypeSet:
		return "set"
	default:
		return "unknown"
	}
}

// shorten shortens the text to the given size.
func shorten(size int, text string) string {
	if size < 1 || len(text) <= size {
		return text
	}

	const sep = ". "
	brief := ""
	chunks := strings.Split(text, sep)
	for i := 0; len(brief) <= size && i < len(chunks); i++ {
		if i > 0 {
			brief += sep
		}
		brief += chunks[i]
	}

	return brief
}
