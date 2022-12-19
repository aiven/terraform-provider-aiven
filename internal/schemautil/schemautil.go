package schemautil

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/docker/go-units"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

func OptionalString(d *schema.ResourceData, key string) string {
	str, ok := d.Get(key).(string)
	if !ok {
		return ""
	}
	return str
}

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

func ParseOptionalStringToFloat64(val interface{}) *float64 {
	v, ok := val.(string)
	if !ok {
		return nil
	}

	if v == "" {
		return nil
	}

	res, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil
	}

	return &res
}

func ParseOptionalStringToBool(val interface{}) *bool {
	v, ok := val.(string)
	if !ok {
		return nil
	}

	if v == "" {
		return nil
	}

	res, err := strconv.ParseBool(v)
	if err != nil {
		return nil
	}

	return &res
}

func CreateOnlyDiffSuppressFunc(_, _, _ string, d *schema.ResourceData) bool {
	return d.IsNewResource()
}

// EmptyObjectDiffSuppressFunc suppresses a diff for service user configuration options when
// fields are not set by the user but have default or previously defined values.
func EmptyObjectDiffSuppressFunc(k, old, new string, _ *schema.ResourceData) bool {
	// When a map inside a list contains only default values without explicit values set by
	// the user Terraform interprets the map as not being present and the array length being
	// zero, resulting in bogus update that does nothing. Allow ignoring those.
	if old == "1" && new == "0" && strings.HasSuffix(k, ".#") {
		return true
	}

	// Ignore the field when it is not set to any value, but had a non-empty parameter before. This also accounts
	// for the case when the field is not set to any value, but has a default value returned by the API.
	// TODO: Uncomment this when we use actual types for the schema.
	// if !d.HasChange(k) && (new == "" && old != "" || new == "0" && old != "0" || new == "false" && old == "true") {
	if new == "" && old != "" {
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
		if sh.Type == schema.TypeList {
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

// IpFilterArrayDiffSuppressFunc Terraform does not allow default values for arrays but
// the IP filter user config value has default. We don't want to force users to always
// define explicit value just because of the Terraform restriction so suppress the
// change from default to empty (which would be nonsensical operation anyway)
func IpFilterArrayDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
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

func IpFilterValueDiffSuppressFunc(k, old, new string, _ *schema.ResourceData) bool {
	// TODO: Add support for ip_filter_object.

	return old == "0.0.0.0/0" && new == "" && strings.HasSuffix(k, ".ip_filter.0")
}

func TrimSpaceDiffSuppressFunc(_, old, new string, _ *schema.ResourceData) bool {
	return strings.TrimSpace(old) == strings.TrimSpace(new)
}

// ValidateDurationString is a ValidateFunc that ensures a string parses
// as time.Duration format
func ValidateDurationString(v interface{}, k string) (ws []string, errors []error) {
	if _, err := time.ParseDuration(v.(string)); err != nil {
		errors = append(errors, fmt.Errorf("%q: invalid duration", k))
	}
	return
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

// unmarshalUserConfig unmarshals the user config from the state to []map[string]interface{} format.
func unmarshalUserConfig(src interface{}) ([]map[string]interface{}, error) {
	configList, ok := src.([]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: expected []interface{}", userconfig.ErrInvalidStateType)
	}

	if len(configList) == 0 {
		return nil, nil
	}

	config, ok := configList[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: expected map[string]interface{}", userconfig.ErrInvalidStateType)
	}

	return []map[string]interface{}{config}, nil
}
