package schemautil

import "golang.org/x/exp/slices"

// sensitiveFields is a list of fields that are not returned by the API on a refresh, but are supposed to remain in the
// state to make Terraform work properly.
var sensitiveFields = []string{
	"admin_username",
	"admin_password",
}

// copySensitiveFields preserves sensitive fields in the state that are not returned by the API on a refresh.
func copySensitiveFields(old, new map[string]interface{}) {
	for _, k := range sensitiveFields {
		if v, ok := old[k]; ok {
			new[k] = v
		}
	}
}

// normalizeIPFilter compares a list of IP filters set in the old user config and a sorted version coming from the new
// user config and returns the re-sorted IP filters, such that all matching entries will be in the same order as
// defined in the old user config.
func normalizeIPFilter(_, new map[string]interface{}) {
	if ip, ok := new["ip_filter_object"].([]interface{}); ok {
		sortObjects(ip, "network")
	}

	if ip, ok := new["ip_filter_string"].([]interface{}); ok {
		sortStrings(ip)
	}

	if ip, ok := new["ip_filter"].([]interface{}); ok {
		sortStrings(ip)
	}
}

// stringSuffixForIPFilters adds a _string suffix to the IP filters.
func stringSuffixForIPFilters(new map[string]interface{}) {
	ipFilters := new["ip_filter"].([]interface{})

	if ipFilters == nil {
		return
	}

	new["ip_filter_string"] = ipFilters

	new["ip_filter"] = nil
}

// stringSuffixForNamespaces adds a _string suffix to the namespaces.
func stringSuffixForNamespaces(new map[string]interface{}) {
	namespaces := new["namespaces"].([]interface{})

	if namespaces == nil {
		return
	}

	new["namespace_string"] = namespaces

	new["namespaces"] = nil
}

// sortStrings sorts untyped list of strings
func sortStrings(list []interface{}) {
	slices.SortFunc(list, func(a, b interface{}) bool {
		aa, _ := a.(string)
		bb, _ := b.(string)
		return aa < bb
	})
}

// sortObjects sorts list of objects by give key
func sortObjects(list []interface{}, key string) {
	slices.SortFunc(list, func(a, b interface{}) bool {
		aa, _ := a.(map[string]interface{})[key].(string)
		bb, _ := b.(map[string]interface{})[key].(string)
		return aa < bb
	})
}
