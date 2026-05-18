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
	PossibleValuesPrefix       = "The possible value"
	LimitedAvailabilityMessage = "To enable this feature, contact the [sales team](http://aiven.io/contact)."
)

// String is a function that returns the string representation of the entity type.
func (et EntityType) String() string {
	return [...]string{"resource", "data source"}[et]
}

// DescriptionBuilder is a helper to build complex descriptions in a consistent way.
type DescriptionBuilder struct {
	// entityType is the type of the entity that the description is for.
	entityType EntityType
	// base is the base of the description.
	base string
	// withBeta adds a beta availability warning.
	withBeta bool
	// withLimitedAvailability adds a limited availability warning.
	withLimitedAvailability bool
	// withPossibleValues is a flag that indicates if the possible values should be included.
	withPossibleValues []string
	// withMinLen is a flag that indicates if the minimum length should be included.
	withMinLen *int
	// withMaxLen is a flag that indicates if the maximum length should be included.
	withMaxLen *int
	// withMinimum is a flag that indicates if the minimum value should be included.
	withMinimum *int
	// withMaximum is a flag that indicates if the maximum value should be included.
	withMaximum *int
	// withPattern is the regular expression that string values must match.
	withPattern string
	// withDefaultValue is a flag that indicates if the default value should be included.
	withDefaultValue any
	// withUseReference is a flag that indicates if the reference should be used.
	withUseReference bool
	// withForceNew is a flag that indicates if the force new should be included.
	withForceNew bool

	// TF validators https://developer.hashicorp.com/terraform/plugin/framework/migrating/attributes-blocks/validators-predefined#background
	withRequiredWith, withConflictsWith, withExactlyOneOf, withAtLeastOneOf []string

	// withLookupID and withLookupComposedOf describe a data source lookup contract:
	// users must supply either `withLookupID` or all of `withLookupComposedOf`.
	withLookupID         string
	withLookupComposedOf []string

	// withRemoveMissing removes the resource from the state if it's missing (i.e., if Read() returns an avngen.IsNotFound error).
	withRemoveMissing bool

	deprecationMessage string
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

func (db *DescriptionBuilder) Beta() *DescriptionBuilder {
	db.withBeta = true
	return db
}

func (db *DescriptionBuilder) LimitedAvailability() *DescriptionBuilder {
	db.withLimitedAvailability = true
	return db
}

func (db *DescriptionBuilder) PossibleValuesString(values ...string) *DescriptionBuilder {
	db.withPossibleValues = values
	return db
}

// RequiredWith is a function that sets the withRequiredWith flag.
// Also known as AlsoRequires in TF Plugin Framework.
func (db *DescriptionBuilder) RequiredWith(values ...string) *DescriptionBuilder {
	db.withRequiredWith = values
	return db
}

func (db *DescriptionBuilder) ConflictsWith(values ...string) *DescriptionBuilder {
	db.withConflictsWith = values
	return db
}

func (db *DescriptionBuilder) ExactlyOneOf(values ...string) *DescriptionBuilder {
	db.withExactlyOneOf = values
	return db
}

func (db *DescriptionBuilder) AtLeastOneOf(values ...string) *DescriptionBuilder {
	db.withAtLeastOneOf = values
	return db
}

// Lookup describes the data source lookup contract: callers must supply either `id`
// or all of `composedOf` together.
func (db *DescriptionBuilder) Lookup(id string, composedOf ...string) *DescriptionBuilder {
	db.withLookupID = id
	db.withLookupComposedOf = composedOf
	return db
}

func (db *DescriptionBuilder) MinLen(length int) *DescriptionBuilder {
	db.withMinLen = &length
	return db
}

func (db *DescriptionBuilder) MaxLen(length int) *DescriptionBuilder {
	db.withMaxLen = &length
	return db
}

func (db *DescriptionBuilder) Minimum(value int) *DescriptionBuilder {
	db.withMinimum = &value
	return db
}

func (db *DescriptionBuilder) Maximum(value int) *DescriptionBuilder {
	db.withMaximum = &value
	return db
}

// Pattern sets the regular expression that string values must match.
func (db *DescriptionBuilder) Pattern(pattern string) *DescriptionBuilder {
	db.withPattern = pattern
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

func (db *DescriptionBuilder) Deprecated(msg string) *DescriptionBuilder {
	db.deprecationMessage = msg
	return db
}

func (db *DescriptionBuilder) RemoveMissing() *DescriptionBuilder {
	db.withRemoveMissing = true
	return db
}

// Build is a function that builds the description.
func (db *DescriptionBuilder) Build() string {
	builder := new(strings.Builder)

	builder.WriteString(db.base)

	// Add a trailing dot if it's missing.
	if !strings.HasSuffix(db.base, ".") {
		builder.WriteString(".")
	}

	const availabilityCommonPart = `

**This %[1]s is in the %[2]s stage and may change without notice.** %[3]s`

	if db.withBeta {
		builder.WriteRune(' ')
		fmt.Fprintf(builder,
			availabilityCommonPart,
			db.entityType.String(),
			"beta",
			fmt.Sprintf("Set\nthe `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the %s.", db.entityType.String()))
	}

	if db.withLimitedAvailability {
		builder.WriteRune(' ')
		fmt.Fprintf(builder,
			availabilityCommonPart,
			db.entityType.String(),
			"limited availability",
			LimitedAvailabilityMessage)
	}

	if db.withPossibleValues != nil {
		builder.WriteRune(' ')
		builder.WriteString(PossibleValuesPrefix)
		if len(db.withPossibleValues) == 1 {
			builder.WriteString(" is ")
		} else {
			builder.WriteString("s are ")
		}
		builder.WriteString(listOfCodes("and", db.withPossibleValues...))
		builder.WriteRune('.')
	}

	// Adds validators information.
	validators := [][]string{
		db.withRequiredWith,
		db.withConflictsWith,
		db.withExactlyOneOf,
		db.withAtLeastOneOf,
	}

	validatorConjunctions := []string{"and", "and", "or", "or"}
	validatorTitles := []string{
		"The field is required with ",
		"The field conflicts with ",
		"Exactly one of the fields must be specified: ",
		"At least one of the fields must be specified: ",
	}

	for i, v := range validators {
		if len(v) > 0 {
			builder.WriteRune(' ')
			builder.WriteString(validatorTitles[i])
			builder.WriteString(listOfCodes(validatorConjunctions[i], v...))
			builder.WriteRune('.')
		}
	}

	if db.withLookupID != "" && len(db.withLookupComposedOf) > 0 {
		builder.WriteRune(' ')
		switch len(db.withLookupComposedOf) {
		case 1:
			builder.WriteString("Exactly one of the fields must be specified: ")
			builder.WriteString(listOfCodes("or", db.withLookupID, db.withLookupComposedOf[0]))
			builder.WriteRune('.')
		default:
			builder.WriteString("Provide either ")
			builder.WriteString(fmt.Sprintf("`%s`", db.withLookupID))
			builder.WriteString(", or all of ")
			builder.WriteString(listOfCodes("and", db.withLookupComposedOf...))
			builder.WriteString(" together.")
		}
	}

	switch {
	case db.withMinLen != nil && db.withMaxLen != nil:
		builder.WriteRune(' ')
		if *db.withMinLen == *db.withMaxLen {
			builder.WriteString(fmt.Sprintf("Length must be exactly `%d`.", *db.withMinLen))
		} else {
			builder.WriteString(fmt.Sprintf("Length must be between `%d` and `%d`.", *db.withMinLen, *db.withMaxLen))
		}
	case db.withMinLen != nil:
		builder.WriteRune(' ')
		builder.WriteString(fmt.Sprintf("Minimum length: `%d`.", *db.withMinLen))
	case db.withMaxLen != nil:
		builder.WriteRune(' ')
		builder.WriteString(fmt.Sprintf("Maximum length: `%d`.", *db.withMaxLen))
	}

	switch {
	case db.withMinimum != nil && db.withMaximum != nil:
		builder.WriteRune(' ')
		builder.WriteString(fmt.Sprintf("Value must be between `%d` and `%d`.", *db.withMinimum, *db.withMaximum))
	case db.withMinimum != nil:
		builder.WriteRune(' ')
		builder.WriteString(fmt.Sprintf("Minimum value: `%d`.", *db.withMinimum))
	case db.withMaximum != nil:
		builder.WriteRune(' ')
		builder.WriteString(fmt.Sprintf("Maximum value: `%d`.", *db.withMaximum))
	}

	if db.withPattern != "" {
		builder.WriteRune(' ')
		builder.WriteString(fmt.Sprintf("Must match pattern: `%s`.", db.withPattern))
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

	if db.withRemoveMissing {
		builder.WriteRune(' ')
		builder.WriteString("If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.")
	}

	// Avoids redundant descriptions.
	s := strings.TrimSpace(builder.String())
	if s == "." {
		return ""
	}

	if db.deprecationMessage != "" {
		s += fmt.Sprintf(" **Deprecated**: %s", db.deprecationMessage)
	}
	return s
}

// listOfCodes turns ["a", "b", "c"] into "`a`, `b` and `c`"
func listOfCodes(conj string, source ...string) string {
	lastOne := len(source) - 1
	items := make([]string, len(source))
	for i, v := range source {
		pre := ""
		switch i {
		case 0:
		case lastOne:
			pre = fmt.Sprintf(" %s ", conj)
		default:
			pre = ", "
		}

		items[i] = fmt.Sprintf("%s`%s`", pre, v)
	}

	return strings.Join(items, "")
}
