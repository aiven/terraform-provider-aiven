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
	"github.com/docker/go-units"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// OptionalStringPointer retrieves a string pointer to a field, empty string
// will be converted to nil
func OptionalStringPointer(d *schema.ResourceData, key string) *string {
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
func OptionalIntPointer(d *schema.ResourceData, key string) *int {
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
func OptionalBoolPointer(d *schema.ResourceData, key string) *bool {
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

func ParseOptionalStringToInt64(val interface{}) *int64 {
	v, ok := val.(string)
	if !ok {
		return nil
	}

	if v == "" {
		return nil
	}

	res, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return nil
	}

	return &res
}

func CreateOnlyDiffSuppressFunc(_, _, _ string, d *schema.ResourceData) bool {
	return len(d.Id()) > 0
}

// EmptyObjectDiffSuppressFunc suppresses a diff for service user configuration options when
// fields are not set by the user but have default or previously defined values.
func EmptyObjectDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	// When a map inside a list contains only default values without explicit values set by
	// the user Terraform interprets the map as not being present and the array length being
	// zero, resulting in bogus update that does nothing. Allow ignoring those.
	if old == "1" && new == "0" && strings.HasSuffix(k, ".#") {
		return true
	}

	// Ignore the field when it is not set to any value, but had a non-empty parameter before. This also accounts
	// for the case when the field is not set to any value, but has a default value returned by the API.
	if !d.HasChange(k) && (new == "" && old != "" || new == "0" && old != "0" || new == "false" && old == "true") {
		return true
	}

	// There is a bug in Terraform 0.11 which interprets "true" as "0" and "false" as "1"
	if (new == "0" && old == "false") || (new == "1" && old == "true") {
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

	return func(k, old, new string, d *schema.ResourceData) bool {
		for _, key := range skipKeys {
			if strings.Contains(k, fmt.Sprintf(".%s.", key)) {
				return false
			}
		}

		return EmptyObjectDiffSuppressFunc(k, old, new, d)
	}
}

// EmptyObjectNoChangeDiffSuppressFunc it suppresses a diff if a field is empty but have not
// been set before to any value
func EmptyObjectNoChangeDiffSuppressFunc(k, _, new string, d *schema.ResourceData) bool {
	if d.HasChange(k) {
		return false
	}

	if new == "" {
		return true
	}

	return false
}

// IPFilterArrayDiffSuppressFunc Terraform does not allow default values for arrays but
// the IP filter user config value has default. We don't want to force users to always
// define explicit value just because of the Terraform restriction so suppress the
// change from default to empty (which would be nonsensical operation anyway)
func IPFilterArrayDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	// TODO: Add support for ip_filter_object.

	if old == "1" && new == "0" && strings.HasSuffix(k, ".ip_filter.#") {
		if list, ok := d.Get(strings.TrimSuffix(k, ".#")).([]interface{}); ok {
			if len(list) == 1 {
				return list[0] == "0.0.0.0/0"
			}
		}
	}

	return false
}

func IPFilterValueDiffSuppressFunc(k, old, new string, _ *schema.ResourceData) bool {
	// TODO: Add support for ip_filter_object.

	return old == "0.0.0.0/0" && new == "" && strings.HasSuffix(k, ".ip_filter.0")
}

func TrimSpaceDiffSuppressFunc(_, old, new string, _ *schema.ResourceData) bool {
	return strings.TrimSpace(old) == strings.TrimSpace(new)
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

func FlattenToString(a []interface{}) []string {
	r := make([]string, len(a))
	for i, v := range a {
		r[i] = fmt.Sprint(v)
	}

	return r
}

func CopyServiceUserPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
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
func ResourceDataGet(d *schema.ResourceData, s map[string]*schema.Schema, dto any) error {
	config := d.GetRawConfig().AsValueMap()
	m := make(map[string]any)
	for k, v := range s {
		// If the field in the tf config
		if c, ok := config[k]; !ok || c.IsNull() {
			continue
		}

		value := d.Get(k)
		if v.Type == schema.TypeSet {
			set, ok := value.(*schema.Set)
			if !ok {
				return fmt.Errorf("expected type Set, got %T", value)
			}
			m[k] = set.List()
		} else {
			m[k] = value
		}
	}

	b, err := json.Marshal(&m)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, dto)
	return err
}

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
//	err := ResourceDataSet(d, s, dto)
func ResourceDataSet(d *schema.ResourceData, s map[string]*schema.Schema, dto any) error {
	b, err := json.Marshal(dto)
	if err != nil {
		return err
	}

	var m map[string]any
	err = json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	result, err := serializeMap(s, m)
	if err != nil {
		return err
	}

	for k, v := range result {
		if err = d.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func serializeMap(s map[string]*schema.Schema, m map[string]any) (map[string]any, error) {
	result := make(map[string]any)
	for name, field := range s {
		value, ok := m[name]
		if !ok {
			continue
		}

		val, err := serialize(field, value)
		if err != nil {
			return nil, err
		}
		result[name] = val
	}

	return result, nil
}

func serialize(s *schema.Schema, value any) (any, error) {
	switch s.Type {
	case schema.TypeList, schema.TypeSet:
	default:
		return value, nil
	}

	var err error
	list, isList := value.([]any)

	switch elem := s.Elem.(type) {
	case *schema.Schema:
		// This branch converts a list of scalars
		if !isList {
			return nil, fmt.Errorf("expected a list, but %T", value)
		}

		result := make([]any, len(list))
		for i, v := range list {
			result[i], err = serialize(elem, v)
			if err != nil {
				return nil, err
			}
		}

		return schema.NewSet(schema.HashSchema(elem), result), nil
	case *schema.Resource:
		// This branch converts a list of objects or a single object
		if m, ok := value.(map[string]any); ok {
			return serializeMap(elem.Schema, m)
		}

		if !isList {
			return nil, fmt.Errorf("expected a map or a list, but %T", value)
		}

		result := make([]any, len(list))
		for i, v := range list {
			m, ok := v.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("expected a map, got %T", v)
			}

			result[i], err = serializeMap(elem.Schema, m)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	}

	// It is either a schema.Resource or schema.Schema
	panic(fmt.Errorf("invalid schema type %T", s.Elem))
}
