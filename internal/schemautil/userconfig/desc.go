package userconfig

import (
	"fmt"
	"strings"
)

// DescriptionBuilder is a helper to build complex descriptions in a consistent way.
type DescriptionBuilder struct {
	// base is the base description.
	base string

	// withMaxLen is a flag that indicates if the max length should be included in the description.
	withMaxLen int

	// withForceNew is a flag that indicates if the force new should be included in the description.
	withForceNew bool

	// withRequiredWith is a flag that indicates if the required with should be included in the description.
	withRequiredWith []string

	// withUseReference is a flag that indicates if the use reference should be included in the description.
	withUseReference bool

	// withDefaultValue is a flag that indicates if the default value should be included in the description.
	withDefaultValue interface{}

	// withPossibleValues is a flag that indicates if the possible values should be included in the description.
	withPossibleValues []interface{}
}

// Desc is a function that creates a new DescriptionBuilder.
func Desc(base string) *DescriptionBuilder {
	return &DescriptionBuilder{base: base}
}

// ForceNew is a function that sets the withForceNew flag.
func (b *DescriptionBuilder) ForceNew() *DescriptionBuilder {
	b.withForceNew = true
	return b
}

// Referenced is a function that sets the withUseReference flag.
func (b *DescriptionBuilder) Referenced() *DescriptionBuilder {
	b.withUseReference = true
	return b
}

// RequiredWith is a function that sets the withRequiredWith flag.
func (b *DescriptionBuilder) RequiredWith(sv ...string) *DescriptionBuilder {
	b.withRequiredWith = sv
	return b
}

// MaxLen is a function that sets the withMaxLen flag.
func (b *DescriptionBuilder) MaxLen(i int) *DescriptionBuilder {
	b.withMaxLen = i
	return b
}

// DefaultValue is a function that sets the withDefaultValue flag.
func (b *DescriptionBuilder) DefaultValue(v interface{}) *DescriptionBuilder {
	b.withDefaultValue = v
	return b
}

// PossibleValues is a function that sets the withPossibleValues flag.
func (b *DescriptionBuilder) PossibleValues(vv ...interface{}) *DescriptionBuilder {
	b.withPossibleValues = vv
	return b
}

// Build is a function that builds the description.
func (b *DescriptionBuilder) Build() string {
	builder := new(strings.Builder)

	builder.WriteString(b.base)
	// TODO: Uncomment in a separate PR.
	//// Capitalize the first letter.
	//builder.WriteRune(rune(strings.ToUpper(string(b.base[0]))[0]))
	//builder.WriteString(b.base[1:])
	//
	//// Add a trailing dot if it's missing.
	//if !strings.HasSuffix(b.base, ".") {
	//	builder.WriteString(".")
	//}

	if b.withPossibleValues != nil {
		builder.WriteRune(' ')

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

		builder.WriteRune('.')
	}

	if b.withRequiredWith != nil {
		builder.WriteRune(' ')

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

		builder.WriteRune('.')
	}

	if b.withMaxLen > 0 {
		builder.WriteRune(' ')

		// TODO: Change to lowercase `l` in a separate PR.
		builder.WriteString(fmt.Sprintf("Maximum Length: `%v`.", b.withMaxLen))
	}

	if b.withDefaultValue != nil {
		builder.WriteRune(' ')

		builder.WriteString(fmt.Sprintf("The default value is `%v`.", b.withDefaultValue))
	}

	if b.withUseReference {
		builder.WriteRune(' ')

		builder.WriteString("To set up proper dependencies please refer to this variable as a reference.")
	}

	if b.withForceNew {
		builder.WriteRune(' ')

		builder.WriteString("This property cannot be changed, doing so forces recreation of the resource.")
	}

	return builder.String()
}
