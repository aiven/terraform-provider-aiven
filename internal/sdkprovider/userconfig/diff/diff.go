package diff

import (
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	// reIsSetElement Set item ends with a 9-length hash int.
	reIsSetElement      = regexp.MustCompile(`\.[0-9]{9}$`)
	reIsIPFilterStrings = regexp.MustCompile(`\.(ip_filter|ip_filter_string)\.`)
)

// SuppressUnchanged suppresses diff for unchanged fields.
// Applied for all nested values: both for objects and arrays.
func SuppressUnchanged(k, oldValue, newValue string, d *schema.ResourceData) bool {
	if d.Id() == "" {
		// Do not suppress diff for new resources, must show the whole diff.
		return false
	}

	// schema.TypeMap
	if strings.HasSuffix(k, ".%") {
		return oldValue == newValue
	}

	// Lists, sets and objects (object is list with one item).
	if strings.HasSuffix(k, ".#") {
		if d.HasChange(k) {
			// By some reason terraform might mark objects as "changed".
			// In that case, terraform returns a list with a nil value.
			// "nil" means that the object has no value.
			key := strings.TrimSuffix(k, ".#")
			v, ok := d.Get(key).([]any)
			return ok && len(v) == 1 && v[0] == nil
		}

		// Suppresses object diff, because it might have been received as default value from the API.
		// So the diff happens when the object's field is changed.
		// Object is a list, and both set and list end with "#".
		// So set == list == object (by type).
		// A set of objects is different.
		// Because hash is calculated for the whole object, not per field.
		return !isObjectSet(k, d)
	}

	// SuppressUnchanged is applied to each nested field.
	// Ip filter items handled with a special suppressor.
	if reIsIPFilterStrings.MatchString(k) {
		return suppressIPFilterSet(k, oldValue, newValue, d)
	}

	// Doesn't suppress "set" items.
	if reIsSetElement.MatchString(k) {
		return false
	}

	// Object properties.
	// "oldValue" — is something read from API
	// "newValue" — is what is read from tf file
	// If value is "computed" (received as default) it has non-empty oldValue (any value) and empty "newValue" (zero value).
	// For instance, when you create kafka it gets "kafka_version = 3.5",
	// while it's not in your tf file, terraform shows a diff.
	// This switch suppresses that, as well, as other "default" values.
	switch newValue {
	case "", "0", "false":
		// "" — kafka_version = "3.5" -> ""
		// 0 — backup_hour = "4" -> 0
		// false — allow_sign_up = true -> false
		return !d.HasChange(k)
	}
	return false
}

// suppressIPFilterSet ip_filter list has specific logic, like default list value
func suppressIPFilterSet(k, oldValue, newValue string, d *schema.ResourceData) bool {
	// Suppresses ip_filter = [0.0.0.0/0]
	path := strings.Split(k, ".")
	// Turns ~ip_filter.1234 to ~ip_filter.#
	v, ok := d.GetOk(strings.Join(path[:len(path)-1], ".") + ".#")
	// Literally, if the value is "0.0.0.0/0" and the parent's length is "1"
	return oldValue == "0.0.0.0/0" && newValue == "" && ok && v.(int) == 1
}

// isObjectSet returns true if given k is for collection of objects
func isObjectSet(k string, d *schema.ResourceData) bool {
	path := strings.Split(strings.TrimSuffix(k, ".#"), ".")
	value := d.GetRawState().AsValueMap()[path[0]]

	// user_config field
	path = path[1:]

	// Drills down the field
	for _, v := range path {
		if v == "0" {
			value = value.AsValueSlice()[0]
			continue
		}
		value = value.GetAttr(v)
	}

	t := value.Type()
	return t.IsSetType() && t.ElementType().IsObjectType()
}
