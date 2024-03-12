package userconfig

import (
	"fmt"
	"strings"
)

// EntityType is a type that represents the type of an entity.
type EntityType int

const (
	// Resource is a constant that represents the resource entity type.
	Resource EntityType = iota
	// DataSource is a constant that represents the data source entity type.
	DataSource
)

// String is a function that returns the string representation of the entity type.
func (et EntityType) String() string {
	return [...]string{"resource", "datasource"}[et]
}

// AvailabilityType is a type that represents the availability type of an entity.
type AvailabilityType int

const (
	// Beta is a constant that represents the beta availability type.
	Beta AvailabilityType = iota + 1
	// Limited is a constant that represents the limited availability type.
	Limited
)

// DescriptionBuilder is a helper to build complex descriptions in a consistent way.
type DescriptionBuilder struct {
	// entityType is the type of the entity that the description is for.
	entityType EntityType
	// base is the base of the description.
	base string
	// availabilityType is the availability type of the entity that the description is for.
	availabilityType AvailabilityType
	// withForcedFirstLetterCapitalization is a flag that indicates if the first letter should be capitalized.
	withForcedFirstLetterCapitalization bool
	// withPossibleValues is a flag that indicates if the possible values should be included.
	withPossibleValues []any
	// withRequiredWith is a flag that indicates if the required with should be included.
	withRequiredWith []string
	// withMaxLen is a flag that indicates if the maximum length should be included.
	withMaxLen int
	// withDefaultValue is a flag that indicates if the default value should be included.
	withDefaultValue any
	// withUseReference is a flag that indicates if the reference should be used.
	withUseReference bool
	// withForceNew is a flag that indicates if the force new should be included.
	withForceNew bool
}

// Desc is a function that creates a new DescriptionBuilder.
func Desc(base string) *DescriptionBuilder {
	return &DescriptionBuilder{base: base}
}

// MarkAsDataSource is a function that marks the entity whose description is being built as a data source.
//
// All entities are considered resources by default, so this function is only needed when the entity is a data source.
//
// If you want to mark the entity as a resource, you don't need to call any additional functions, and you can proceed
// further with the description building.
func (db *DescriptionBuilder) MarkAsDataSource() *DescriptionBuilder {
	db.entityType = DataSource
	return db
}

// AvailabilityType is a function that sets the availabilityType field.
func (db *DescriptionBuilder) AvailabilityType(t AvailabilityType) *DescriptionBuilder {
	db.availabilityType = t
	return db
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

	if db.availabilityType != 0 {
		builder.WriteRune(' ')

		const availabilityCommonPart = `
**This %[1]s is in the %[2]s stage and may change without notice.** %[3]s
the ` + "`PROVIDER_AIVEN_ENABLE_BETA`" + ` environment variable to use the %[1]s.`

		switch db.availabilityType {
		case Beta:
			builder.WriteString(fmt.Sprintf(
				availabilityCommonPart,
				db.entityType.String(),
				"beta",
				"Set",
			))
		case Limited:
			builder.WriteString(fmt.Sprintf(
				availabilityCommonPart,
				db.entityType.String(),
				"limited availability",
				"\nTo enable this feature, contact the [sales team](mailto:sales@aiven.io). After it's enabled, set",
			))
		}
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
		builder.WriteString(fmt.Sprintf(
			"Changing this property forces recreation of the %s.", db.entityType.String(),
		))
	}

	return builder.String()
}
