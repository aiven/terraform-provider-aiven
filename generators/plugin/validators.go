package main

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

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
	switch {
	case item.MinLength > 0 && item.MaxLength > 0:
		codes = append(codes, jen.Qual(pkg, "LengthBetween").Call(jen.Lit(item.MinLength), jen.Lit(item.MaxLength)))
	case item.MinLength > 0:
		codes = append(codes, jen.Qual(pkg, "LengthAtLeast").Call(jen.Lit(item.MinLength)))
	case item.MaxLength > 0:
		codes = append(codes, jen.Qual(pkg, "LengthAtMost").Call(jen.Lit(item.MaxLength)))
	}

	if item.Type == SchemaTypeString && isValidRegex(item.Pattern) {
		codes = append(codes, jen.Qual(pkg, "RegexMatches").Call(
			jen.Qual("regexp", "MustCompile").Call(jen.Lit(item.Pattern)),
			jen.Lit(fmt.Sprintf("must match pattern %q", item.Pattern)),
		))
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
		// Lit() takes any, so it works naturally with string and int enums
		// distinct() sorts and removes duplicates
		sorted := distinct(item.Enum...)
		enums := lo.Map[any, jen.Code](sorted, func(v any, _ int) jen.Code {
			return jen.Lit(v)
		})

		codes = append(codes, jen.Qual(pkg, "OneOf").Call(enums...))
	}

	// Scalar types have Required property
	if item.IsNested() {
		if item.Required {
			codes = append(codes, jen.Qual(pkg, "IsRequired").Call())
		}

		if entity.isResource() && item.IsObject() {
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
