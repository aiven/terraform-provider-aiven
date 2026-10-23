package main

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type stringFormatSpec struct {
	validator  string         // constructor in customValidatorsPackage
	hint       string         // doc hint, skipped when the description matches documented
	documented *regexp.Regexp // matches descriptions that already explain the format
	example    string         // wins over spec examples when the format is declared
}

var stringFormats = map[StringFormat]stringFormatSpec{
	StringFormatCIDR: {
		validator:  "CIDR",
		hint:       "Must be a valid CIDR block, for example `10.0.0.0/24`.",
		documented: regexp.MustCompile(`\b\d{1,3}(\.\d{1,3}){3}\b|[0-9a-fA-F]{1,4}::`),
		example:    "10.0.0.0/24",
	},
	StringFormatIPv4CIDR: {
		validator:  "IPv4CIDR",
		hint:       "Must be a valid IPv4 CIDR block, for example `10.0.0.0/24`.",
		documented: regexp.MustCompile(`\b\d{1,3}(\.\d{1,3}){3}\b`),
		example:    "10.0.0.0/24",
	},
	StringFormatUUID: {
		validator:  "UUID",
		hint:       "Must be a valid UUID.",
		documented: regexp.MustCompile(`(?i)uuid`),
		example:    uuidExample,
	},
}

// itemFormat returns the declared format, or guesses it from the name of a string field.
func itemFormat(item *Item) StringFormat {
	switch {
	case item.Format != "":
		return item.Format
	case item.Type != SchemaTypeString:
		return ""
	case item.Name == "uuid":
		return StringFormatUUID
	case strings.Contains(item.Name, "cidr") && item.MaxLength == ipv4CIDRMaxLength:
		return StringFormatIPv4CIDR
	case strings.Contains(item.Name, "cidr"):
		return StringFormatCIDR
	}
	return ""
}

const (
	uuidLength        = 36
	ipv4CIDRMaxLength = len("255.255.255.255/32")
)

// hasStringFormat reports whether the item gets a format validator, which supersedes length limits.
func hasStringFormat(item *Item) bool {
	_, ok := stringFormats[itemFormat(item)]
	return ok
}

func genValidators(def *Definition, entity entityType, item *Item) ([]jen.Code, error) {
	pkg := getTypedImport(item.Type, validatorTypedImport)
	codes := make([]jen.Code, 0)

	// Integers
	if item.Type == SchemaTypeInteger {
		switch {
		case item.Minimum > 0 && item.Maximum > 0:
			codes = append(codes, jen.Qual(pkg, "Between").Call(jen.Lit(item.Minimum), jen.Lit(item.Maximum)))
		case item.Minimum > 0:
			codes = append(codes, jen.Qual(pkg, "AtLeast").Call(jen.Lit(item.Minimum)))
		case item.Maximum > 0:
			codes = append(codes, jen.Qual(pkg, "AtMost").Call(jen.Lit(item.Maximum)))
		}
	}

	// Strings
	// Length constraints only exist on stringvalidator; guard by type so that a
	// scalar field overridden to an array/set (which may carry a stray maxLength)
	// does not emit a non-existent setvalidator.Length* call.
	if item.Type == SchemaTypeString && !hasStringFormat(item) {
		switch {
		case item.MinLength > 0 && item.MaxLength > 0:
			codes = append(codes, jen.Qual(pkg, "LengthBetween").Call(jen.Lit(item.MinLength), jen.Lit(item.MaxLength)))
		case item.MinLength > 0:
			codes = append(codes, jen.Qual(pkg, "LengthAtLeast").Call(jen.Lit(item.MinLength)))
		case item.MaxLength > 0:
			codes = append(codes, jen.Qual(pkg, "LengthAtMost").Call(jen.Lit(item.MaxLength)))
		}
	}

	if item.Type == SchemaTypeString && isValidRegex(item.Pattern) {
		codes = append(codes, jen.Qual(pkg, "RegexMatches").Call(
			jen.Qual("regexp", "MustCompile").Call(jen.Lit(item.Pattern)),
			jen.Lit(fmt.Sprintf("must match pattern %q", item.Pattern)),
		))
	}

	if item.Type == SchemaTypeString {
		if f, ok := stringFormats[itemFormat(item)]; ok {
			codes = append(codes, jen.Qual(customValidatorsPackage, f.validator).Call())
		}
	}

	// Slices
	switch {
	case item.MinItems > 0 && item.MaxItems > 0:
		codes = append(codes, jen.Qual(pkg, "SizeBetween").Call(jen.Lit(item.MinItems), jen.Lit(item.MaxItems)))
	case item.MinItems > 0:
		codes = append(codes, jen.Qual(pkg, "SizeAtLeast").Call(jen.Lit(item.MinItems)))
	case item.MaxItems > 0:
		codes = append(codes, jen.Qual(pkg, "SizeAtMost").Call(jen.Lit(item.MaxItems)))
	}

	// Enums
	if len(item.Enum) > 0 {
		codes = append(codes, oneOf(pkg, item.Enum))
	}

	// Array item enums, e.g., listvalidator.ValueStringsAre(stringvalidator.OneOf(...))
	if item.IsArray() && item.Items.IsScalar() && item.Items.Type != SchemaTypeBoolean && len(item.Items.Enum) > 0 {
		method := "Value" + typingMapping()[item.Items.Type] + "sAre"
		codes = append(codes, jen.Qual(pkg, method).Call(oneOf(getTypedImport(item.Items.Type, validatorTypedImport), item.Items.Enum)))
	}

	// Scalar types have Required property
	if item.IsNested() {
		if item.Required {
			codes = append(codes, jen.Qual(pkg, "IsRequired").Call())
		}

		if entity.isResource() && item.IsObject() {
			// todo: drop for objects in v5.0.0, see Item.TFType.
			codes = append(codes, jen.Qual(pkg, "SizeAtMost").Call(jen.Lit(1)))
		}
	}

	// A quick implementation of validators.
	// It might be that root level rules require a better way of building "paths".
	// https://developer.hashicorp.com/terraform/plugin/framework/migrating/attributes-blocks/validators-predefined#examples
	//
	// Overlap with ExactlyOneOf is suppressed; mirrored in fmtDescription.
	exactly := item.ExactlyOneOf
	if !entity.isResource() && item.IsRootProperty() && slices.Contains(def.Datasource.ExactlyOneOf, item.Name) {
		exactly = def.Datasource.ExactlyOneOf
	}
	validators := map[string][]string{
		"ConflictsWith": lo.Without(item.ConflictsWith, exactly...),
		"ExactlyOneOf":  item.ExactlyOneOf,
		"AtLeastOneOf":  item.AtLeastOneOf,
		"AlsoRequires":  item.AlsoRequires,
	}

	validatorKeys := sortedKeys(validators)
	for _, name := range validatorKeys {
		list := validators[name]
		if len(list) > 0 {
			paths, err := siblingPath(list...)
			if err != nil {
				return nil, err
			}
			codes = append(codes, jen.Qual(pkg, name).Call(paths...))
		}
	}

	return codes, nil
}

func oneOf(pkg string, values []any) jen.Code {
	// Lit() takes any, so it works naturally with string and int enums
	// distinct() sorts and removes duplicates
	enums := lo.Map[any, jen.Code](distinct(values...), func(v any, _ int) jen.Code {
		return jen.Lit(v)
	})
	return jen.Qual(pkg, "OneOf").Call(enums...)
}

// rePathName so far we don't support complex paths, just relative attributes
var rePathName = regexp.MustCompile(`[a-z_][a-z0-9_]*`)

func siblingPath(list ...string) ([]jen.Code, error) {
	paths := make([]jen.Code, 0, len(list))
	for _, v := range list {
		if !rePathName.MatchString(v) {
			return nil, fmt.Errorf("invalid path name: %q, must be: `[a-z_][a-z0-9_]*`", v)
		}

		p := jen.
			Qual(pathPackage, "MatchRelative").Call().
			Dot("AtParent").Call().
			Dot("AtName").Call(jen.Lit(v))
		paths = append(paths, p)
	}
	return paths, nil
}
