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

// normalizeIpFilter compares a list of IP filters set in the old user config and a sorted version coming from the new
// user config and returns the re-sorted IP filters, such that all matching entries will be in the same order as
// defined in the old user config.
func normalizeIpFilter(old, new map[string]interface{}) {
	oldIpFilters, _ := old["ip_filter"].([]interface{})

	newIpFilters, _ := new["ip_filter"].([]interface{})

	fieldToWrite := "ip_filter"

	if oldIpFilters == nil || newIpFilters == nil {
		var ok bool

		oldIpFilters, ok = old["ip_filter_object"].([]interface{})
		if !ok {
			return
		}

		newIpFilters, ok = new["ip_filter_object"].([]interface{})
		if !ok {
			return
		}

		fieldToWrite = "ip_filter_object"
	}

	var normalizedIpFilters []interface{}

	var nonexistentIpFilters []interface{}

	// First, we take all the elements from old and if they match with the elements in new,
	// we preserve them in the same order as they were defined in old.
	for _, o := range oldIpFilters {
		found := false

		for _, n := range newIpFilters {
			// Define two comparison variables to avoid code duplication in the loop.
			var comparableO interface{}

			var comparableN interface{}

			// If we're dealing with a string format, we need to compare the values directly.
			if fieldToWrite == "ip_filter" {
				comparableO = o

				comparableN = n
			} else {
				// If we're dealing with an object format, we need to compare the values of the "network" field.
				comparableO = o.(map[string]interface{})["network"]

				comparableN = n.(map[string]interface{})["network"]
			}

			if comparableO == comparableN {
				normalizedIpFilters = append(normalizedIpFilters, o)

				found = true

				break
			}
		}

		// We append old elements that are missing in new to nonexistentIpFilters.
		if !found {
			nonexistentIpFilters = append(nonexistentIpFilters, o)
		}
	}

	// Second, we take the new and check whether there are any differences with the old, and
	// append those to nonexistentIpFilters.
	for _, n := range newIpFilters {
		found := false

		for _, o := range oldIpFilters {
			var comparableO interface{}

			var comparableN interface{}

			if fieldToWrite == "ip_filter" {
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
			nonexistentIpFilters = append(nonexistentIpFilters, n)
		}
	}

	new[fieldToWrite] = append(normalizedIpFilters, nonexistentIpFilters...)
}
