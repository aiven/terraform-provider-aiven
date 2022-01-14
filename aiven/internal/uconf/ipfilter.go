// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package uconf

// NormalizeIpFilter compares a list of IP filters set in TF and a sorted version coming
// from Aiven and takes sort IP filters such that all matching entries will be in
// the same order as defined in the TF manifest.
func NormalizeIpFilter(tfUserConfig interface{}, userConfig []map[string]interface{}) []map[string]interface{} {
	tfInt, ok := tfUserConfig.([]interface{})
	if !ok || len(tfInt) == 0 {
		return userConfig
	}

	if len(userConfig) == 0 {
		return userConfig
	}

	if _, ok := userConfig[0]["ip_filter"]; !ok {
		return userConfig
	}

	if _, ok := tfInt[0].(map[string]interface{})["ip_filter"]; !ok {
		return userConfig
	}

	api := toStringSlice(userConfig[0]["ip_filter"].([]interface{}))
	tf := toStringSlice(tfInt[0].(map[string]interface{})["ip_filter"].([]interface{}))

	var newIpFilters []string
	var diff []string

	// First, we take all the elements from TF and if they match with the data
	// coming from API we preserve the same order as was defined in manifest
	for _, t := range tf {
		found := false
		if t == "" {
			continue
		}

		for _, a := range api {
			if t == a {
				newIpFilters = append(newIpFilters, t)
				found = true
				break
			}
		}

		// TF elements that are missing in API version we add to a diff
		if !found {
			diff = append(diff, t)
		}
	}

	// Second, we take an API version and only detect whether there is any
	// difference with the TF version and update diff
	for _, a := range api {
		found := false
		for _, t := range tf {
			if a == t {
				found = true
				break
			}
		}

		if !found {
			diff = append(diff, a)
		}
	}

	userConfig[0]["ip_filter"] = trim(append(newIpFilters, diff...))
	return userConfig
}

// toStringSlice converts []interface to a []string
func toStringSlice(s []interface{}) []string {
	if len(s) == 0 {
		return nil
	}

	r := make([]string, len(s))
	for i, e := range s {
		r[i] = e.(string)
	}

	return r
}

// trim removes all empty string "" elements from a string slice
func trim(s []string) []interface{} {
	var res []interface{}

	for _, i := range s {
		if i != "" {
			res = append(res, i)
		}
	}

	return res
}
