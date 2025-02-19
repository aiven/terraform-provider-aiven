package schemautil

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/docker/go-units"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/samber/lo"
)

// OptionalStringPointer retrieves a string pointer to a field, empty string
// will be converted to nil
func OptionalStringPointer(d ResourceData, key string) *string {
	val, ok := d.GetOk(key)
	if !ok {
		return nil
	}
	str, ok := val.(string)
	if !ok {
		return nil
	}
	return &str
}

// OptionalIntPointer retrieves an int pointer to a field, if the field is not set, returns nil.
func OptionalIntPointer(d ResourceData, key string) *int {
	val, ok := d.GetOk(key)
	if !ok {
		return nil
	}
	intValue, ok := val.(int)
	if !ok {
		return nil
	}
	return &intValue
}

// OptionalBoolPointer retrieves a bool pointer to a field, if the field is not set, returns nil.
func OptionalBoolPointer(d ResourceData, key string) *bool {
	val, ok := d.GetOk(key)
	if !ok {
		return nil
	}

	boolValue, ok := val.(bool)
	if !ok {
		return nil
	}

	return &boolValue
}

func ToOptionalString(val interface{}) string {
	switch v := val.(type) {
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case string:
		return v
	default:
		return ""
	}
}

func CreateOnlyDiffSuppressFunc(_, _, _ string, d *schema.ResourceData) bool {
	return len(d.Id()) > 0
}

// EmptyObjectDiffSuppressFunc suppresses a diff for service user configuration options when
// fields are not set by the user but have default or previously defined values.
func EmptyObjectDiffSuppressFunc(k, oldValue, newValue string, d *schema.ResourceData) bool {
	// When a map inside a list contains only default values without explicit values set by
	// the user Terraform interprets the map as not being present and the array length being
	// zero, resulting in bogus update that does nothing. Allow ignoring those.
	if oldValue == "1" && newValue == "0" && strings.HasSuffix(k, ".#") {
		return true
	}

	// Ignore the field when it is not set to any value, but had a non-empty parameter before. This also accounts
	// for the case when the field is not set to any value, but has a default value returned by the API.
	if !d.HasChange(k) && (newValue == "" && oldValue != "" || newValue == "0" && oldValue != "0" || newValue == "false" && oldValue == "true") {
		return true
	}

	// There is a bug in Terraform 0.11 which interprets "true" as "0" and "false" as "1"
	if (newValue == "0" && oldValue == "false") || (newValue == "1" && oldValue == "true") {
		return true
	}

	return false
}

// EmptyObjectDiffSuppressFuncSkipArrays generates a DiffSuppressFunc that skips all the array/list fields
// and uses schemautil.EmptyObjectDiffSuppressFunc in all others cases
func EmptyObjectDiffSuppressFuncSkipArrays(s map[string]*schema.Schema) schema.SchemaDiffSuppressFunc {
	var skipKeys []string
	for key, sh := range s {
		switch sh.Type {
		case schema.TypeList:
			skipKeys = append(skipKeys, key)
		}
	}

	return func(k, oldValue, newValue string, d *schema.ResourceData) bool {
		for _, key := range skipKeys {
			if strings.Contains(k, fmt.Sprintf(".%s.", key)) {
				return false
			}
		}

		return EmptyObjectDiffSuppressFunc(k, oldValue, newValue, d)
	}
}

// EmptyObjectNoChangeDiffSuppressFunc it suppresses a diff if a field is empty but have not
// been set before to any value
func EmptyObjectNoChangeDiffSuppressFunc(k, _, newValue string, d *schema.ResourceData) bool {
	if d.HasChange(k) {
		return false
	}

	if newValue == "" {
		return true
	}

	return false
}

// IPFilterArrayDiffSuppressFunc Terraform does not allow default values for arrays but
// the IP filter user config value has default. We don't want to force users to always
// define explicit value just because of the Terraform restriction so suppress the
// change from default to empty (which would be nonsensical operation anyway)
func IPFilterArrayDiffSuppressFunc(k, oldValue, newValue string, d *schema.ResourceData) bool {
	// TODO: Add support for ip_filter_object.

	if oldValue == "1" && newValue == "0" && strings.HasSuffix(k, ".ip_filter.#") {
		if list, ok := d.Get(strings.TrimSuffix(k, ".#")).([]interface{}); ok {
			if len(list) == 1 {
				return list[0] == "0.0.0.0/0"
			}
		}
	}

	return false
}

func IPFilterValueDiffSuppressFunc(k, oldValue, newValue string, _ *schema.ResourceData) bool {
	// TODO: Add support for ip_filter_object.

	return oldValue == "0.0.0.0/0" && newValue == "" && strings.HasSuffix(k, ".ip_filter.0")
}

func TrimSpaceDiffSuppressFunc(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	return strings.TrimSpace(oldValue) == strings.TrimSpace(newValue)
}

// ValidateHumanByteSizeString is a ValidateFunc that ensures a string parses
// as units.Bytes format
func ValidateHumanByteSizeString(v interface{}, k string) (ws []string, errors []error) {
	// only allow `^[1-9][0-9]*(GiB|G)*` without fractions
	if ok, _ := regexp.MatchString("^[1-9][0-9]*(GiB|G)$", v.(string)); !ok {
		return ws, append(errors, fmt.Errorf("%q: configured string must match ^[1-9][0-9]*(G|GiB)", k))
	}
	if _, err := units.RAMInBytes(v.(string)); err != nil {
		return ws, append(errors, fmt.Errorf("%q: invalid human readable byte size", k))
	}
	return
}

// ValidateEmailAddress is a ValidateFunc that ensures a string is a valid email address
func ValidateEmailAddress(v any, k string) (ws []string, errors []error) {
	addr, err := mail.ParseAddress(v.(string))
	if err != nil {
		errors = append(errors, err)

		return
	}

	if strings.ToLower(addr.Address) != addr.Address {
		errors = append(errors, fmt.Errorf("%q: invalid email address", k))
	}

	return
}

func BuildResourceID(parts ...string) string {
	finalParts := make([]string, len(parts))
	for idx, part := range parts {
		finalParts[idx] = url.PathEscape(part)
	}
	return strings.Join(finalParts, "/")
}

func SplitResourceID(resourceID string, n int) (parts []string, err error) {
	parts = strings.SplitN(resourceID, "/", n)

	for idx, part := range parts {
		part, _ := url.PathUnescape(part)
		parts[idx] = part
	}

	if len(parts) != n {
		err = fmt.Errorf("invalid resource id: %s", resourceID)
		return nil, err
	}

	return
}

func SplitResourceID2(resourceID string) (string, string, error) {
	parts, err := SplitResourceID(resourceID, 2)
	if err != nil {
		return "", "", err
	}

	return parts[0], parts[1], nil
}

func SplitResourceID3(resourceID string) (string, string, string, error) {
	parts, err := SplitResourceID(resourceID, 3)
	if err != nil {
		return "", "", "", err
	}

	return parts[0], parts[1], parts[2], nil
}

func SplitResourceID4(resourceID string) (string, string, string, string, error) {
	parts, err := SplitResourceID(resourceID, 4)
	if err != nil {
		return "", "", "", "", err
	}

	return parts[0], parts[1], parts[2], parts[3], nil
}

func FlattenToString[T any](a []T) []string {
	r := make([]string, len(a))
	for i, v := range a {
		r[i] = fmt.Sprint(v)
	}

	return r
}

func CopyServiceUserPropertiesFromAPIResponseToTerraform(
	d ResourceData,
	user *aiven.ServiceUser,
	projectName string,
	serviceName string,
) error {
	if err := d.Set("project", projectName); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("username", user.Username); err != nil {
		return err
	}
	if err := d.Set("password", user.Password); err != nil {
		return err
	}
	if err := d.Set("type", user.Type); err != nil {
		return err
	}

	if len(user.AccessCert) > 0 {
		if err := d.Set("access_cert", user.AccessCert); err != nil {
			return err
		}
	}
	if len(user.AccessKey) > 0 {
		if err := d.Set("access_key", user.AccessKey); err != nil {
			return err
		}
	}

	return nil
}

func CopyServiceUserGenPropertiesFromAPIResponseToTerraform(
	d ResourceData,
	user *service.ServiceUserGetOut,
	projectName string,
	serviceName string,
) error {
	if err := d.Set("project", projectName); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("username", user.Username); err != nil {
		return err
	}
	if err := d.Set("password", user.Password); err != nil {
		return err
	}
	if err := d.Set("type", user.Type); err != nil {
		return err
	}

	if user.AccessCert != nil {
		if err := d.Set("access_cert", user.AccessCert); err != nil {
			return err
		}
	}
	if user.AccessKey != nil {
		if err := d.Set("access_key", user.AccessKey); err != nil {
			return err
		}
	}

	return nil
}

// ResourceDataGet Marshals schema.ResourceData into the given dto.
// Instead of setting every field individually and dealing with pointers,
// it creates a map of values using the schema keys,
// and then marshals the result into given DTO.
// Instead of for each field:
//
//	v, ok := d.GetOk("foo")
//	if ok {
//		dto.Foo = &v
//	}
//
// Use:
//
//	err := ResourceDataGet(d, s, dto)
//
// Warning: doesn't support nested sets.
// Warning: not tested with nested objects.
func ResourceDataGet(d ResourceData, dto any, fns ...MapModifier) error {
	rawConfig := d.GetRawConfig()
	if rawConfig.IsNull() {
		return nil
	}

	config := rawConfig.AsValueMap()
	m := make(map[string]any)
	for k, c := range config {
		// If the field in the tf config, or array doesn't have changes
		if c.IsNull() && !d.HasChange(k) {
			continue
		}

		value := d.Get(k)
		set, ok := value.(*schema.Set)
		if ok {
			value = set.List()
		}

		m[k] = serializeGet(value)
	}

	for _, f := range fns {
		if err := f(d, m); err != nil {
			return err
		}
	}

	return Remarshal(&m, dto)
}

func serializeGet(value any) any {
	switch t := value.(type) {
	case *schema.Set:
		return serializeGet(t.List())
	case []any:
		for i, v := range t {
			t[i] = serializeGet(v)
		}
		return t
	case map[string]any:
		for k, v := range t {
			t[k] = serializeGet(v)
		}
	}
	return value
}

// MapModifier modifier for key/value pair
type MapModifier func(d ResourceData, dto map[string]any) error

// errMissingForceNew `terraform import` requires all ForceNew fields to be set
// otherwise it will output "replace" plan.
// All ForceNew fields must be set.
var errMissingForceNew = fmt.Errorf("missing required ForceNew field")

// ResourceDataSet Sets the given dto values to schema.ResourceData
// Instead of setting every field individually and dealing with pointers,
// it creates a map of values using the schema keys,
// and then sets the result to schema.ResourceData.
// Instead of for each field:
//
//	if dto.Foo != nil {
//		if err := d.Set("foo", *dto.Foo); err != nil {
//			return err
//		}
//	}
//
// Use:
//
//	err := ResourceDataSet(d, dto, s)
func ResourceDataSet(d ResourceData, dto any, s map[string]*schema.Schema, fns ...MapModifier) error {
	var m map[string]any
	err := Remarshal(dto, &m)
	if err != nil {
		return err
	}

	for _, f := range fns {
		if err := f(d, m); err != nil {
			return err
		}
	}

	if err = serializeSet(s, m); err != nil {
		return err
	}

	for k, v := range m {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}

	return nil
}

func serializeSet(s map[string]*schema.Schema, m map[string]any) error {
	keys := lo.Uniq(append(lo.Keys(m), lo.Keys(s)...))
	for _, k := range keys {
		prop, ok := s[k]
		if !ok {
			// The field not in the schema, removes it.
			delete(m, k)
			continue
		}

		value, ok := m[k]
		if !ok {
			if prop.ForceNew && prop.Required {
				return fmt.Errorf("%w: %q", errMissingForceNew, k)
			}

			// The field is not in the dto, moves on.
			continue
		}

		res, ok := prop.Elem.(*schema.Resource)
		if !ok {
			continue
		}

		// When we have an object, we need to convert it to a list.
		// So there is no difference between a single object and a list of objects.
		var list []any
		if v, ok := value.([]any); ok {
			list = v
		} else {
			list = append(list, value)
		}

		m[k] = list
		for _, v := range list {
			o, ok := v.(map[string]any)
			if !ok {
				return fmt.Errorf("expected object for field %q, got type %T", k, v)
			}

			if err := serializeSet(res.Schema, o); err != nil {
				return fmt.Errorf("failed to serialize object %q: %w", k, err)
			}
		}
	}

	return nil
}

// RenameAlias renames field names terraform name -> dto name
// Example: RenameAlias("hasFoo", "wantBar", "hasBaz", "wantEgg")
func RenameAlias(keys ...string) MapModifier {
	m := make(map[string]string, len(keys)/2)
	for i := 0; i < len(keys); i += 2 {
		m[keys[i]] = keys[i+1]
	}
	return RenameAliases(m)
}

// RenameAliases renames field names terraform name -> dto name
func RenameAliases(aliases map[string]string) MapModifier {
	return func(_ ResourceData, dto map[string]any) error {
		for k, v := range dto {
			alias, ok := aliases[k]
			if ok {
				dto[alias] = v
				delete(dto, k)
			}
		}
		return nil
	}
}

// RenameAliasesReverse reverse version of RenameAliases
func RenameAliasesReverse(aliases map[string]string) MapModifier {
	m := make(map[string]string, len(aliases))
	for k, v := range aliases {
		m[v] = k
	}
	return RenameAliases(m)
}

// ResourceIDKeys spits the resource id into parts and sets them to the given keys (if not set)
func ResourceIDKeys(keys ...string) MapModifier {
	return func(d ResourceData, dto map[string]any) error {
		chunks, err := SplitResourceID(d.Id(), len(keys))
		if err != nil {
			return err
		}

		for i, k := range keys {
			if _, ok := dto[k]; !ok {
				dto[k] = chunks[i]
			}
		}
		return nil
	}
}

// Remarshal marshals "in" object to "out" through json
func Remarshal(in, out any) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}
