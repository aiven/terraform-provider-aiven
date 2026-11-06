package schemautil

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/aiven/go-client-codegen/handler/organization"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// ResourceStateOrResourceDiff either *schema.ResourceState or *schema.ResourceDiff
type ResourceStateOrResourceDiff interface {
	GetRawConfig() cty.Value
	GetOk(key string) (any, bool)
	Get(key string) any
}

// PlanParameters service plan aparameters
type PlanParameters struct {
	DiskSizeMBDefault int
	DiskSizeMBStep    int
	DiskSizeMBMax     int
}

// IsEmpty checks whether PlanParameters contains only zero values,
// which usually indicates that the service plan is not compatible with legacy endpoints.
// In such cases, calculations are performed with zero values as placeholders.
func (p PlanParameters) IsEmpty() bool {
	return p.DiskSizeMBDefault == 0
}

// IsOrganizationID is a helper function that returns true if the string is an organization ID.
func IsOrganizationID(s string) bool {
	return strings.HasPrefix(s, "org")
}

// NormalizeOrganizationID is a helper function that returns the ID to use for the API call.
// If the ID is an organization ID, it will be converted to an account ID via the API.
// If the ID is an account ID, it will be returned as is, without performing any API calls.
func NormalizeOrganizationID(ctx context.Context, client *aiven.Client, id string) (string, error) {
	if IsOrganizationID(id) {
		r, err := client.Organization.Get(ctx, id)
		if err != nil {
			return "", err
		}

		id = r.AccountID
	}

	return id, nil
}

// organizationGetter helper type to shrinks the avngen.Client interface size.
type organizationGetter interface {
	OrganizationGet(ctx context.Context, id string) (*organization.OrganizationGetOut, error)
}

// ConvertOrganizationToAccountID transforms provided ID to an account ID via API call if it is an organization ID.
func ConvertOrganizationToAccountID(ctx context.Context, id string, client organizationGetter) (string, error) {
	if IsOrganizationID(id) {
		resp, err := client.OrganizationGet(ctx, id)
		if err != nil {
			return "", fmt.Errorf("error converting organization ID %q to account ID: %w", id, err)
		}

		return resp.AccountId, nil
	}

	return id, nil
}

// StringToDiagWarning is a function that converts a string to a diag warning.
func StringToDiagWarning(msg string) diag.Diagnostics {
	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  msg,
	}}
}

// ErrorToDiagWarning is a function that converts an error to a diag warning.
func ErrorToDiagWarning(err error) diag.Diagnostics {
	return StringToDiagWarning(err.Error())
}
