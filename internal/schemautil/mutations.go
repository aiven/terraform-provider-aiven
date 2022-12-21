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

// normalizeIpFilter compares a list of IP filters set in the oldUserConfig and a sorted version coming from
// newUserConfig and returns the re-sorted IP filters, such that all matching entries will be in
// the same order as defined in the oldUserConfig.
func normalizeIpFilter(old, new map[string]interface{}) {
	oldIpFilters, ok := old["ip_filter"].([]interface{})
	if !ok {
		return
	}

	newIpFilters, ok := new["ip_filter"].([]interface{})
	if !ok {
		return
	}

	var normalizedIpFilters []interface{}

	var nonexistentIpFilters []interface{}

	// First, we take all the elements from old and if they match with the elements in new,
	// we preserve them in the same order as they were defined in old.
	for _, o := range oldIpFilters {
		found := false

		for _, n := range newIpFilters {
			if o == n {
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
			if n == o {
				found = true

				break
			}
		}

		if !found {
			nonexistentIpFilters = append(nonexistentIpFilters, n)
		}
	}

	new["ip_filter"] = append(normalizedIpFilters, nonexistentIpFilters...)
}
