package main

import (
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

func genValidators(item *Item) ([]jen.Code, error) {
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
			if item.IsObject() {
				codes = append(codes, jen.Qual(pkg, "SizeAtMost").Call(jen.Lit(1)))
			}
		}
	}

	return codes, nil
}
