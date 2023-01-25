package schemautil

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/exp/slices"
)

const DeprecationMessage = "This resource is deprecated and will be removed in the next major release."

//goland:noinspection GoDeprecation
func GetACLUserValidateFunc() schema.SchemaValidateFunc { //nolint:staticcheck
	return validation.StringMatch(
		regexp.MustCompile(`^[-._*?A-Za-z0-9]+$`),
		"Must consist of alpha-numeric characters, underscores, dashes, dots and glob characters '*' and '?'")
}

//goland:noinspection GoDeprecation
func GetServiceUserValidateFunc() schema.SchemaValidateFunc { //nolint:staticcheck
	return validation.StringMatch(
		regexp.MustCompile(`^(\*$|[a-zA-Z0-9_?][a-zA-Z0-9-_?*\.].{0,62})$`),
		"username should be alphanumeric, may not start with dash or dot, max 64 characters")
}

var (
	CommonSchemaProjectReference = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9_-]*$"), "project name should be alphanumeric"),
		Description:  Complex("Identifies the project this resource belongs to.").ForceNew().Referenced().Build(),
	}

	CommonSchemaServiceNameReference = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9_-]*$"), "common name should be alphanumeric"),
		Description:  Complex("Specifies the name of the service that this resource belongs to.").ForceNew().Referenced().Build(),
	}
)

// DescriptionBuilder is a helper to build complex descriptions in a consistent way.
type DescriptionBuilder struct {
	base               string
	withMaxLen         int
	withForceNew       bool
	withRequiredWith   []string
	withDeprecation    string
	withUseReference   bool
	withDefaultValue   interface{}
	withPossibleValues []interface{}
}

func Complex(base string) *DescriptionBuilder {
	return &DescriptionBuilder{base: base}
}

func (b *DescriptionBuilder) ForceNew() *DescriptionBuilder {
	b.withForceNew = true
	return b
}

func (b *DescriptionBuilder) Deprecate(msg string) *DescriptionBuilder {
	b.withDeprecation = msg
	return b
}

func (b *DescriptionBuilder) Referenced() *DescriptionBuilder {
	b.withUseReference = true
	return b
}

func (b *DescriptionBuilder) RequiredWith(s ...string) *DescriptionBuilder {
	b.withRequiredWith = s
	return b
}

func (b *DescriptionBuilder) MaxLen(i int) *DescriptionBuilder {
	b.withMaxLen = i
	return b
}

func (b *DescriptionBuilder) DefaultValue(i interface{}) *DescriptionBuilder {
	b.withDefaultValue = i
	return b
}

func (b *DescriptionBuilder) PossibleValues(is ...interface{}) *DescriptionBuilder {
	b.withPossibleValues = is
	return b
}

func (b *DescriptionBuilder) Build() string {
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

func StringSliceToInterfaceSlice(s []string) []interface{} {
	res := make([]interface{}, len(s))
	for i := range s {
		res[i] = s[i]
	}
	return res
}

func SetTagsTerraformProperties(t map[string]string) []map[string]interface{} {
	var tags []map[string]interface{}
	for k, v := range t {
		tags = append(tags, map[string]interface{}{
			"key":   k,
			"value": v,
		})
	}

	return tags
}

func GetTagsFromSchema(d *schema.ResourceData) map[string]string {
	tags := make(map[string]string)

	for _, tag := range d.Get("tag").(*schema.Set).List() {
		tagVal := tag.(map[string]interface{})
		tags[tagVal["key"].(string)] = tagVal["value"].(string)
	}

	return tags
}

func ValidateEnum[T string | int](enum ...T) schema.SchemaValidateDiagFunc {
	return func(i interface{}, path cty.Path) diag.Diagnostics {
		value := i.(T)
		if !slices.Contains(enum, value) {
			allowed := make([]string, 0, len(enum))
			for _, s := range enum {
				allowed = append(allowed, fmt.Sprintf("%q", s))
			}
			return diag.Errorf("%q is not one of: %s", value, strings.Join(allowed, ", "))
		}
		return nil
	}
}

// PointerValueOrDefault returns pointer's value or default
func PointerValueOrDefault[T comparable](v *T, d T) T {
	if v == nil {
		return d
	}
	return *v
}
