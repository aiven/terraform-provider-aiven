// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	commonSchemaProjectReference = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9_-]*$"), "project name should be alphanumeric"),
		Description:  complex("Identifies the project this resource belongs to.").forceNew().referenced().build(),
	}

	commonSchemaServiceNameReference = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9_-]*$"), "service name should be alphanumeric"),
		Description:  complex("Specifies the name of the service that this resource belongs to.").forceNew().referenced().build(),
	}
)

// descriptionBuilder is a helper to build complex descriptions in a consistent way.
type descriptionBuilder struct {
	base               string
	withMaxLen         int
	withForceNew       bool
	withRequiredWith   []string
	withDeprecation    string
	withUseReference   bool
	withDefaultValue   interface{}
	withPossibleValues []interface{}
}

func complex(base string) *descriptionBuilder {
	return &descriptionBuilder{base: base}
}

func (b *descriptionBuilder) forceNew() *descriptionBuilder {
	b.withForceNew = true
	return b
}

func (b *descriptionBuilder) deprecate(msg string) *descriptionBuilder {
	b.withDeprecation = msg
	return b
}

func (b *descriptionBuilder) referenced() *descriptionBuilder {
	b.withUseReference = true
	return b
}

func (b *descriptionBuilder) requiredWith(s ...string) *descriptionBuilder {
	b.withRequiredWith = s
	return b
}

func (b *descriptionBuilder) maxLen(i int) *descriptionBuilder {
	b.withMaxLen = i
	return b
}

func (b *descriptionBuilder) defaultValue(i interface{}) *descriptionBuilder {
	b.withDefaultValue = i
	return b
}

func (b *descriptionBuilder) possibleValues(is ...interface{}) *descriptionBuilder {
	b.withPossibleValues = is
	return b
}

func (b *descriptionBuilder) build() string {
	builder := new(strings.Builder)

	if b.withDeprecation != "" {
		builder.WriteString("**DEPRECATED ")
		builder.WriteString(b.withDeprecation)
		builder.WriteString("** ")
	}

	builder.WriteString(b.base)
	if b.withPossibleValues != nil {
		builder.WriteByte(' ')
		builder.WriteString("The possible values are ")
		for i := range b.withPossibleValues {
			if i > 0 {
				if i == len(b.withPossibleValues)-1 {
					builder.WriteString(" and ")
				} else {
					builder.WriteString(", ")
				}
			}
			builder.WriteString(fmt.Sprintf("`%v`", b.withPossibleValues[i]))
		}
		builder.WriteByte('.')
	}
	if b.withRequiredWith != nil {
		builder.WriteByte(' ')
		builder.WriteString("The field is required with")
		for i := range b.withRequiredWith {
			if i > 0 {
				if i == len(b.withRequiredWith)-1 {
					builder.WriteString(" and ")
				} else {
					builder.WriteString(", ")
				}
			}
			builder.WriteString(fmt.Sprintf("`%v`", b.withRequiredWith[i]))
		}
		builder.WriteByte('.')
	}
	if b.withMaxLen > 0 {
		builder.WriteByte(' ')
		builder.WriteString(fmt.Sprintf("Maximum Length: `%v`.", b.withMaxLen))
	}
	if b.withDefaultValue != nil {
		builder.WriteByte(' ')
		builder.WriteString(fmt.Sprintf("The default value is `%v`.", b.withDefaultValue))
	}
	if b.withUseReference {
		builder.WriteByte(' ')
		builder.WriteString("To set up proper dependencies please refer to this variable as a reference.")
	}
	if b.withForceNew {
		builder.WriteByte(' ')
		builder.WriteString("This property cannot be changed, doing so forces recreation of the resource.")
	}
	return builder.String()
}

func stringSliceToInterfaceSlice(s []string) []interface{} {
	res := make([]interface{}, len(s))
	for i := range s {
		res[i] = s[i]
	}
	return res
}
