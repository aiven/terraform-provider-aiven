package userconfig

import (
	"fmt"
	"strings"
)

// DescriptionBuilder is a helper to build complex descriptions in a consistent way.
type DescriptionBuilder struct {
	// base is the base description.
	base string

	// withForcedFirstLetterCapitalization is a flag that indicates if the first letter should be capitalized.
	withForcedFirstLetterCapitalization bool

	// withPossibleValues is a flag that indicates if the possible values should be included in the description.
	withPossibleValues []interface{}

	// withRequiredWith is a flag that indicates if the required with should be included in the description.
	withRequiredWith []string

	// withMaxLen is a flag that indicates if the max length should be included in the description.
	withMaxLen int

	// withDefaultValue is a flag that indicates if the default value should be included in the description.
	withDefaultValue interface{}

	// withUseReference is a flag that indicates if the use reference should be included in the description.
	withUseReference bool

	// withForceNew is a flag that indicates if the force new should be included in the description.
	withForceNew bool
}

// Desc is a function that creates a new DescriptionBuilder.
func Desc(base string) *DescriptionBuilder {
	return &DescriptionBuilder{base: base}
}

// ForceFirstLetterCapitalization is a function that sets the withForcedFirstLetterCapitalization flag.
func (db *DescriptionBuilder) ForceFirstLetterCapitalization() *DescriptionBuilder {
	db.withForcedFirstLetterCapitalization = true
	return db
}

// PossibleValues is a function that sets the withPossibleValues flag.
func (db *DescriptionBuilder) PossibleValues(vv ...interface{}) *DescriptionBuilder {
	db.withPossibleValues = vv
	return db
}

// RequiredWith is a function that sets the withRequiredWith flag.
func (db *DescriptionBuilder) RequiredWith(sv ...string) *DescriptionBuilder {
	db.withRequiredWith = sv
	return db
}

// MaxLen is a function that sets the withMaxLen flag.
func (db *DescriptionBuilder) MaxLen(i int) *DescriptionBuilder {
	db.withMaxLen = i
	return db
}

// DefaultValue is a function that sets the withDefaultValue flag.
func (db *DescriptionBuilder) DefaultValue(v interface{}) *DescriptionBuilder {
	db.withDefaultValue = v
	return db
}

// Referenced is a function that sets the withUseReference flag.
func (db *DescriptionBuilder) Referenced() *DescriptionBuilder {
	db.withUseReference = true
	return db
}

// ForceNew is a function that sets the withForceNew flag.
func (db *DescriptionBuilder) ForceNew() *DescriptionBuilder {
	db.withForceNew = true
	return db
}

// Build is a function that builds the description.
func (db *DescriptionBuilder) Build() string {
	b := new(strings.Builder)

	// Capitalize the first letter, if needed.
	if db.withForcedFirstLetterCapitalization {
		b.WriteRune(rune(strings.ToUpper(string(db.base[0]))[0]))

		b.WriteString(db.base[1:])
	} else {
		b.WriteString(db.base)
	}

	// Add a trailing dot if it's missing.
	if !strings.HasSuffix(db.base, ".") {
		b.WriteString(".")
	}

	if db.withPossibleValues != nil {
		b.WriteRune(' ')

		b.WriteString("The possible values are ")

		for i := range db.withPossibleValues {
			if i > 0 {
				if i == len(db.withPossibleValues)-1 {
					b.WriteString(" and ")
				} else {
					b.WriteString(", ")
				}
			}

			b.WriteString(fmt.Sprintf("`%v`", db.withPossibleValues[i]))
		}

		b.WriteRune('.')
	}

	if db.withRequiredWith != nil {
		b.WriteRune(' ')

		b.WriteString("The field is required with")

		for i := range db.withRequiredWith {
			if i > 0 {
				if i == len(db.withRequiredWith)-1 {
					b.WriteString(" and ")
				} else {
					b.WriteString(", ")
				}
			}

			b.WriteString(fmt.Sprintf("`%v`", db.withRequiredWith[i]))
		}

		b.WriteRune('.')
	}

	if db.withMaxLen > 0 {
		b.WriteRune(' ')

		b.WriteString(fmt.Sprintf("Maximum length: `%v`.", db.withMaxLen))
	}

	if db.withDefaultValue != nil {
		b.WriteRune(' ')

		b.WriteString(fmt.Sprintf("The default value is `%v`.", db.withDefaultValue))
	}

	if db.withUseReference {
		b.WriteRune(' ')

		b.WriteString("To set up proper dependencies please refer to this variable as a reference.")
	}

	if db.withForceNew {
		b.WriteRune(' ')

		b.WriteString("This property cannot be changed, doing so forces recreation of the resource.")
	}

	return b.String()
}
