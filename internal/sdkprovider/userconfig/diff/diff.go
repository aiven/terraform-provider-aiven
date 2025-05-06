package diff

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/exp/slices"
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

	// SuppressUnchanged is applied to each nested field.
	// Ip filter items handled with a special suppressor.
	if reIsIPFilterStrings.MatchString(k) {
		return suppressIPFilterSet(k, oldValue, newValue, d)
	}

	// Doesn't suppress "set" items.
	if reIsSetElement.MatchString(k) {
		// A hash mostly calculated for the whole object.
		return false
	}

	// Lists, sets and objects (object is list with one item).
	// But never a set of nested objects: it is impossible to find an object in a set by its hash.
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

	// Object properties.
	// "oldValue" — is something read from API
	// "newValue" — is what is read from tf file
	// If value is "computed" (received as default) it has non-empty oldValue (any value) and empty "newValue" (zero value).
	// For instance, when you create kafka it gets "kafka_version = 3.5",
	// while it's not in your tf file, terraform shows a diff.
	// This switch suppresses that, as well, as other "default" values.
	switch newValue {
	case "", "0", "false":
		// 1. d.HasChange(k): detects field changes — it compares the old value with a new value.
		//    When the old value is "unknown" (not set) and the new value is "false",
		//    it compares "false" with "false" and detects no changes.
		//    The same applies to "0" and "" (but an empty strings can't be a real value).
		// 2. oldValue == "": it means the field previously had no value, but now it does.
		//    So it kind of fixes HasChange() for booleans.
		// Still might not detect changes for int fields, because TF stores 0 for them in the state,
		// and the oldValue is always "0", not "".
		// This case can be fixed with the Plugin Framework only.
		return !d.HasChange(k) && oldValue != ""
	}
	return false
}

// suppressIPFilterSet ip_filter list has specific logic
// Doesn't support ip_filter_object.
// By default services have a non-empty ip_filter networks: either ["0.0.0.0/0"] or ["0.0.0.0/0", "::/0"]
// Suppress the diff when user didn't define the ip_filter networks in the config but got it from the API:
// - ip_filter = [0.0.0.0/0, ::/0] -- suppress
// - ip_filter = [0.0.0.0/0] 	   -- suppress
// - ip_filter = [::/0] 		   -- don't
// - ip_filter = [127.0.0.1/32]    -- don't
// As a side effect suppresses "default" networks removal from the config.
// But terraform doesn't know the difference, so nothing we can do here.
func suppressIPFilterSet(k, _, _ string, d *schema.ResourceData) bool {
	const ipFilterDepth = 3 // foo_user_config.0.ip_filter
	// Turns ~ip_filter.1234 to ~ip_filter to get the set.
	ipFilterKey := strings.Join(strings.Split(k, ".")[0:ipFilterDepth], ".")
	if d.HasChange(ipFilterKey) {
		// Something is changed in the ip_filter networks.
		// Shows the diff.
		return false
	}

	// The ip_filter networks is removed.
	// Makes sure it has only default values.
	set, ok := d.Get(ipFilterKey).(*schema.Set)
	if !ok {
		// Can't happen, but shouldn't panic
		return false
	}

	return IsDefaultIPFilterList(set.List())
}

// isObjectSet returns true if the given k is a set, and its elements are objects
// Doesn't support sets with nested objects!
// Terraform fails to calculate hash for nested objects, and can simply not detect changes.
func isObjectSet(k string, d *schema.ResourceData) bool {
	path := strings.Split(strings.TrimSuffix(k, ".#"), ".")
	value := d.GetRawState().AsValueMap()[path[0]]

	// user_config field
	path = path[1:]

	// Drills down the field: foo_user_config.0.ip_filter_object
	for _, v := range path {
		// When mets ".0." it is an index of list/set.
		index, err := strconv.Atoi(v)
		if err == nil {
			value = value.AsValueSlice()[index]
			continue
		}

		value = value.GetAttr(v)
	}

	t := value.Type()
	return t.IsSetType() && t.ElementType().IsObjectType()
}

// defaultIPFilterLists returns service default ip_filter lists
// the old one: ["0.0.0.0/0"]
// the new one: ["0.0.0.0/0", "::/0"]
func defaultIPFilterLists() [][]string {
	return [][]string{
		{"0.0.0.0/0"},
		{"0.0.0.0/0", "::/0"},
	}
}

func IsDefaultIPFilterList(list []any) bool {
	norm := make([]string, len(list))
	for i, v := range list {
		n, ok := NormalizeIPFilter(v)
		if !ok {
			return false
		}
		norm[i] = n
	}

	slices.Sort(norm)
	for _, def := range defaultIPFilterLists() {
		if slices.Equal(norm, def) {
			return true
		}
	}
	return false
}

// NormalizeIPFilter returns ip_filter CIDR:
// 1. when it is a string, it returns string
// 2. when it is a map, it returns the value of the "network" key
func NormalizeIPFilter(v any) (string, bool) {
	s, ok := v.(string)
	if ok {
		return s, ok
	}
	m, ok := v.(map[string]any)
	if ok {
		return NormalizeIPFilter(m["network"])
	}
	return "", false
}
