// Package organization contains the organization related resources and utilities.
package organization

import (
	"strings"

	"github.com/aiven/aiven-go-client"
)

// normalizeID is a helper function that returns the ID to use for the API call.
// If the ID is an organization ID, it will be converted to an account ID via the API.
// If the ID is an account ID, it will be returned as is, without performing any API calls.
func normalizeID(client *aiven.Client, id string) (string, error) {
	if strings.HasPrefix(id, "org") {
		r, err := client.Organization.Get(id)

		if err != nil {
			return "", err
		}

		id = r.AccountID
	}

	return id, nil
}
