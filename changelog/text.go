package main

import (
	"fmt"
	"regexp"
	"slices"
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

var reCode = regexp.MustCompile("`([^`]+)`")

func findEnums(description string) []string {
	var values []string
	switch {
	case strings.Contains(description, userconfig.PossibleValuesPrefix):
		// userconfig.PossibleValuesPrefix is used within "internal" package,
		// see userconfig.DescriptionBuilder.PossibleValuesString()
		parts := strings.Split(description, userconfig.PossibleValuesPrefix)
		if len(parts) != 2 {
			return nil
		}
		values = reCode.FindAllString(parts[1], -1)
	case strings.Contains(strings.ToLower(description), "enum"):
		// "enum" is used in the user config generator
		values = reCode.FindAllString(description, -1)
	}

	if len(values) == 0 {
		return nil
	}

	// Removes ` from the beginning and end of the string
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = v[1 : len(v)-1]
	}

	// Sorts the values for consistent output and comparison
	slices.Sort(result)
	return result
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
