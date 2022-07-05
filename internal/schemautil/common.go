package schemautil

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func GetACLUserValidateFunc() schema.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile(`^[-._*?A-Za-z0-9]+$`),
		"Must consist of alpha-numeric characters, underscores, dashes, dots and glob characters '*' and '?'")
}

func GetServiceUserValidateFunc() schema.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile(`^(\*$|[a-zA-Z0-9-_?][a-zA-Z0-9-_?*]+)$`),
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

func Complex(base string) *descriptionBuilder {
	return &descriptionBuilder{base: base}
}

func (b *descriptionBuilder) ForceNew() *descriptionBuilder {
	b.withForceNew = true
	return b
}

func (b *descriptionBuilder) Deprecate(msg string) *descriptionBuilder {
	b.withDeprecation = msg
	return b
}

func (b *descriptionBuilder) Referenced() *descriptionBuilder {
	b.withUseReference = true
	return b
}

func (b *descriptionBuilder) RequiredWith(s ...string) *descriptionBuilder {
	b.withRequiredWith = s
	return b
}

func (b *descriptionBuilder) MaxLen(i int) *descriptionBuilder {
	b.withMaxLen = i
	return b
}

func (b *descriptionBuilder) DefaultValue(i interface{}) *descriptionBuilder {
	b.withDefaultValue = i
	return b
}

func (b *descriptionBuilder) PossibleValues(is ...interface{}) *descriptionBuilder {
	b.withPossibleValues = is
	return b
}

func (b *descriptionBuilder) Build() string {
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
