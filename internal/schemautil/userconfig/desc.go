package userconfig

import (
	"fmt"
	"strings"
)

// DescriptionBuilder is a helper to build complex descriptions in a consistent way.
type DescriptionBuilder struct {
	base                                string
	withForcedFirstLetterCapitalization bool
	withPossibleValues                  []any
	withRequiredWith                    []string
	withMaxLen                          int
	withDefaultValue                    any
	withUseReference                    bool
	withForceNew                        bool
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
func (db *DescriptionBuilder) PossibleValues(values ...any) *DescriptionBuilder {
	db.withPossibleValues = values
	return db
}

// RequiredWith is a function that sets the withRequiredWith flag.
func (db *DescriptionBuilder) RequiredWith(values ...string) *DescriptionBuilder {
	db.withRequiredWith = values
	return db
}

// MaxLen is a function that sets the withMaxLen flag.
func (db *DescriptionBuilder) MaxLen(length int) *DescriptionBuilder {
	db.withMaxLen = length
	return db
}

// DefaultValue is a function that sets the withDefaultValue flag.
func (db *DescriptionBuilder) DefaultValue(value any) *DescriptionBuilder {
	db.withDefaultValue = value
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
	builder := new(strings.Builder)

	// Capitalize the first letter, if needed.
	if db.withForcedFirstLetterCapitalization {
		builder.WriteRune(rune(strings.ToUpper(string(db.base[0]))[0]))
		builder.WriteString(db.base[1:])
	} else {
		builder.WriteString(db.base)
	}

	// Add a trailing dot if it's missing.
	if !strings.HasSuffix(db.base, ".") {
		builder.WriteString(".")
	}

	if db.withPossibleValues != nil {
		builder.WriteRune(' ')
		builder.WriteString("The possible values are ")
		for i, value := range db.withPossibleValues {
			if i > 0 {
				if i == len(db.withPossibleValues)-1 {
					builder.WriteString(" and ")
				} else {
					builder.WriteString(", ")
				}
			}
			builder.WriteString(fmt.Sprintf("`%v`", value))
		}
		builder.WriteRune('.')
	}

	if db.withRequiredWith != nil {
		builder.WriteRune(' ')
		builder.WriteString("The field is required with")
		for i, value := range db.withRequiredWith {
			if i > 0 {
				if i == len(db.withRequiredWith)-1 {
					builder.WriteString(" and ")
				} else {
					builder.WriteString(", ")
				}
			}
			builder.WriteString(fmt.Sprintf("`%v`", value))
		}
		builder.WriteRune('.')
	}

	if db.withMaxLen > 0 {
		builder.WriteRune(' ')
		builder.WriteString(fmt.Sprintf("Maximum length: `%v`.", db.withMaxLen))
	}

	if db.withDefaultValue != nil {
		builder.WriteRune(' ')
		builder.WriteString(fmt.Sprintf("The default value is `%v`.", db.withDefaultValue))
	}

	if db.withUseReference {
		builder.WriteRune(' ')
		builder.WriteString("To set up proper dependencies please refer to this variable as a reference.")
	}

	if db.withForceNew {
		builder.WriteRune(' ')
		builder.WriteString("This property cannot be changed, doing so forces recreation of the resource.")
	}

	return builder.String()
}
