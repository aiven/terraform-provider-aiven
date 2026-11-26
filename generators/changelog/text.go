package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	cl "github.com/aiven/terraform-provider-aiven/tools/changelog"
)

func hasBeta(description string) bool {
	return strings.Contains(description, util.AivenEnableBeta)
}

func hasLimited(description string) bool {
	description = strings.ToLower(description)
	return strings.Contains(description, strings.ToLower(userconfig.LimitedAvailabilityMarker)) ||
		strings.Contains(description, strings.ToLower(userconfig.LimitedAvailabilityMessage))
}

// parseEnum extracts enum values and returns the description without the enum sentence.
func parseEnum(description string) (values []string, rest string) {
	if strings.HasPrefix(description, userconfig.EnumPrefix) {
		end := indexSentenceEnd(description, false)
		if end < 0 {
			return nil, description
		}
		values = backtickValues(description[len(userconfig.EnumPrefix):end])
		rest = strings.TrimSpace(description[end+1:])
		return sortUnique(values), rest
	}

	idx, skip, ok := possibleValuesAt(description)
	if !ok {
		return nil, description
	}

	sentence := description[idx:]
	end := indexSentenceEnd(sentence, false)
	if end < 0 {
		return nil, description
	}
	valueText := sentence[len(userconfig.PossibleValuesPrefix)+skip : end]
	values = backtickValues(valueText)

	before := strings.TrimSpace(description[:idx])
	after := strings.TrimSpace(sentence[end+1:])
	switch {
	case before == "":
		rest = after
	case after == "":
		rest = before
	default:
		rest = before + " " + after
	}
	return sortUnique(values), rest
}

func possibleValuesAt(description string) (idx, skip int, ok bool) {
	const prefix = userconfig.PossibleValuesPrefix
	start := 0
	for start < len(description) {
		rel := strings.Index(description[start:], prefix)
		if rel < 0 {
			return 0, 0, false
		}
		idx = start + rel
		tail := description[idx+len(prefix):]
		switch {
		case strings.HasPrefix(tail, "s are "):
			skip = len("s are ")
		case strings.HasPrefix(tail, " is "):
			skip = len(" is ")
		default:
			start = idx + 1
			continue
		}
		if len(tail) <= skip || tail[skip] != '`' {
			start = idx + 1
			continue
		}
		return idx, skip, true
	}
	return 0, 0, false
}

func backtickValues(text string) []string {
	var values []string
	for i := 0; i < len(text); {
		if text[i] != '`' {
			i++
			continue
		}
		j := i + 1
		for j < len(text) && text[j] != '`' {
			j++
		}
		if j < len(text) {
			values = append(values, text[i+1:j])
		}
		i = j + 1
	}
	return values
}

func sortUnique(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	slices.Sort(values)
	return slices.Compact(values)
}

// indexSentenceEnd is the index of a sentence-ending '.'; requireSpaceAfter needs ". ".
func indexSentenceEnd(text string, requireSpaceAfter bool) int {
	inTick := false
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '`':
			inTick = !inTick
		case '.':
			if inTick || abbrevDot(text, i) {
				continue
			}
			if requireSpaceAfter {
				if i+1 < len(text) && text[i+1] == ' ' {
					return i
				}
				continue
			}
			if i == len(text)-1 {
				return i
			}
			if text[i+1] == ' ' {
				return i
			}
		}
	}
	if !requireSpaceAfter && len(text) > 0 && text[len(text)-1] == '.' {
		return len(text) - 1
	}
	return -1
}

func abbrevDot(text string, dot int) bool {
	start := dot - 1
	for start >= 0 && (text[start] == '.' || (text[start] >= 'a' && text[start] <= 'z')) {
		start--
	}
	word := text[start+1 : dot]
	switch word {
	case "e.g", "i.e", "etc":
		return true
	default:
		return false
	}
}

// addDetail is the changelog detail for a new field or block (Add).
func addDetail(description string) string {
	values, rest := parseEnum(description)
	b := userconfig.Desc(briefProse(rest))
	if len(values) > 0 {
		b = b.PossibleValuesString(values...)
	}
	return b.Build()
}

// brief shortens a description for changelog details.
func brief(description string) string {
	_, rest := parseEnum(description)
	return briefProse(rest)
}

func briefProse(description string) string {
	const maxLen = 120
	text := cl.NormalizeSpace(description)
	if text == "" {
		return ""
	}

	var parts []string
	chunk := text
	for chunk != "" {
		end := indexSentenceEnd(chunk, true)
		var sent string
		if end < 0 {
			sent = strings.TrimRight(chunk, ".")
			chunk = ""
		} else {
			sent = chunk[:end]
			chunk = strings.TrimSpace(chunk[end+2:])
		}
		if len(parts) == 0 {
			parts = append(parts, sent)
			if chunk == "" {
				break
			}
			continue
		}
		candidate := strings.Join(append(parts, sent), ". ")
		if len(candidate) > maxLen {
			break
		}
		parts = append(parts, sent)
	}
	return strings.TrimRight(strings.Join(parts, ". "), ". ")
}

func isEnum(description string) bool {
	v, _ := parseEnum(description)
	return len(v) > 0
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
