package schemautil

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
func normalizeIPFilter(old, new map[string]interface{}) {
	oldIPFilters, _ := old["ip_filter"].([]interface{})
	newIPFilters, _ := new["ip_filter"].([]interface{})
	fieldToWrite := "ip_filter"

	if oldIPFilters == nil || newIPFilters == nil {
		var ok bool

		oldIPFilters, _ = old["ip_filter_string"].([]interface{})

		newIPFilters, ok = new["ip_filter_string"].([]interface{})

		fieldToWrite = "ip_filter_string"

		if !ok {
			oldIPFilters, ok = old["ip_filter_object"].([]interface{})
			if !ok {
				return
			}

			newIPFilters, ok = new["ip_filter_object"].([]interface{})
			if !ok {
				return
			}

			fieldToWrite = "ip_filter_object"
		}
	}

	var normalizedIPFilters []interface{}
	var nonexistentIPFilters []interface{}

	// First, we take all the elements from old and if they match with the elements in new,
	// we preserve them in the same order as they were defined in old.
	for _, o := range oldIPFilters {
		for _, n := range newIPFilters {
			// Define two comparison variables to avoid code duplication in the loop.
			var comparableO interface{}

			var comparableN interface{}

			// If we're dealing with a string format, we need to compare the values directly.
			if fieldToWrite == "ip_filter" || fieldToWrite == "ip_filter_string" {
				comparableO = o

				comparableN = n
			} else {
				// If we're dealing with an object format, we need to compare the values of the "network" field.
				comparableO = o.(map[string]interface{})["network"]

				comparableN = n.(map[string]interface{})["network"]
			}

			if comparableO == comparableN {
				normalizedIPFilters = append(normalizedIPFilters, o)
				break
			}
		}
	}

	// Second, we take the new and check whether there are any differences with the old, and
	// append those to nonexistentIPFilters.
	for _, n := range newIPFilters {
		found := false

		for _, o := range oldIPFilters {
			var comparableO interface{}

			var comparableN interface{}

			if fieldToWrite == "ip_filter" || fieldToWrite == "ip_filter_string" {
				comparableO = o

				comparableN = n
			} else {
				comparableO = o.(map[string]interface{})["network"]

				comparableN = n.(map[string]interface{})["network"]
			}

			if comparableO == comparableN {
				found = true

				break
			}
		}

		if !found {
			nonexistentIPFilters = append(nonexistentIPFilters, n)
		}
	}

	new[fieldToWrite] = append(normalizedIPFilters, nonexistentIPFilters...)
}

// stringSuffixForIPFilters adds a _string suffix to the IP filters.
func stringSuffixForIPFilters(new map[string]interface{}) {
	if new["ip_filter"] == nil {
		return
	}

	ipFilters := new["ip_filter"].([]interface{})

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
