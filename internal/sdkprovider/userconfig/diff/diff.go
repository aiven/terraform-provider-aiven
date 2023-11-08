package diff

import (
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// reIsSetElement Set item ends with a 9-length hash int.
var reIsSetElement = regexp.MustCompile(`\.[0-9]{9}$`)

// SuppressUnchanged suppresses diff for unchanged fields.
// Applied for all nested values: both for objects and arrays.
func SuppressUnchanged(k, old, new string, d *schema.ResourceData) bool {
	// Lists, sets and objects (object is list with one item).
	if k[len(k)-1:] == "#" {
		if d.HasChange(k) {
			// By some reason terraform might mark objects as "changed".
			// In that case, terraform returns a list with a nil value.
			// "nil" means that the object has no value.
			key := strings.TrimSuffix(k, ".#")
			v, ok := d.Get(key).([]any)
			return ok && len(v) == 1 && v[0] == nil
		}

		// Suppress empty objects and empty arrays
		return true
	}

	// Ip filter items handled with a special suppressor.
	if strings.Contains(k, ".ip_filter.") || strings.Contains(k, ".ip_filter_string.") {
		return suppressIPFilterSet(k, old, new, d)
	}

	// Doesn't suppress "set" items.
	if reIsSetElement.MatchString(k) {
		return false
	}

	// Object properties.
	// "old" — is something read from API
	// "new" — is what is read from tf file
	// If value is "computed" (received as default) it has non-empty old (any value) and empty "new" (zero value).
	// For instance, when you create kafka it gets "kafka_version = 3.5",
	// while it's not in your tf file, terraform shows a diff.
	// This switch suppresses that, as well, as other "default" values.
	switch new {
	case "", "0", "false":
		// "" — kafka_version = "3.5" -> ""
		// 0 — backup_hour = "4" -> 0
		// false — allow_sign_up = true -> false
		return !d.HasChange(k)
	}
	return false
}

// suppressIPFilterSet ip_filter list has specific logic, like default list value
func suppressIPFilterSet(k, old, new string, d *schema.ResourceData) bool {
	// Suppresses ip_filter = [0.0.0.0/0]
	path := strings.Split(k, ".")
	// Turns ~ip_filter.1234 to ~ip_filter.#
	v, ok := d.GetOk(strings.Join(path[:len(path)-1], ".") + ".#")
	// Literally, if the value is "0.0.0.0/0" and the parent's length is "1"
	return old == "0.0.0.0/0" && new == "" && ok && v.(int) == 1
}
