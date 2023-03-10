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

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

//goland:noinspection GoDeprecation
func GetACLUserValidateFunc() schema.SchemaValidateFunc { //nolint:staticcheck
	return validation.StringMatch(
		regexp.MustCompile(`^[-._*?A-Za-z0-9]+$`),
		"Must consist of alpha-numeric characters, underscores, dashes, dots and glob characters '*' and '?'")
}

//goland:noinspection GoDeprecation
func GetServiceUserValidateFunc() schema.SchemaValidateFunc { //nolint:staticcheck
	return validation.StringMatch(
		regexp.MustCompile(`^(\*$|[a-zA-Z0-9_?][a-zA-Z0-9-_?*.].{0,62})$`),
		"username should be alphanumeric, may not start with dash or dot, max 64 characters")
}

var (
	CommonSchemaProjectReference = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9_-]*$"), "project name should be alphanumeric"),
		Description:  userconfig.Desc("Identifies the project this resource belongs to.").ForceNew().Referenced().Build(),
	}

	CommonSchemaServiceNameReference = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9_-]*$"), "common name should be alphanumeric"),
		Description:  userconfig.Desc("Specifies the name of the service that this resource belongs to.").ForceNew().Referenced().Build(),
	}
)

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

func JoinQuoted[T string | int](elems []T, sep, quote string) (result string) {
	for i, v := range elems {
		if i != 0 {
			result += sep
		}
		result = fmt.Sprintf("%s%s%v%s", result, quote, v, quote)
	}
	return result
}
