package org

import (
	"context"
	"strings"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OrganizationModel is the common model for both resource and data source
type OrganizationModel struct {
	// ID is the identifier of the organization
	ID types.String `tfsdk:"id"`

	// Name is the name of the organization
	Name types.String `tfsdk:"name"`

	// TenantID is the tenant identifier of the organization
	TenantID types.String `tfsdk:"tenant_id"`

	// CreateTime is the timestamp of the creation of the organization
	CreateTime types.String `tfsdk:"create_time"`

	// UpdateTime is the timestamp of the last update of the organization
	UpdateTime types.String `tfsdk:"update_time"`
}

// ResourceSchema returns the schema for the organization resource
func ResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "ID of the organization.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Description: "Name of the organization.",
			Required:    true,
		},
		"tenant_id": schema.StringAttribute{
			Description: "Tenant ID of the organization.",
			Computed:    true,
		},
		"create_time": schema.StringAttribute{
			Description: "Timestamp of the creation of the organization.",
			Computed:    true,
		},
		"update_time": schema.StringAttribute{
			Description: "Timestamp of the last update of the organization.",
			Computed:    true,
		},
	}
}

// ResolveAccountID is a helper function that returns the account ID to use for API calls.
// If the ID is an organization ID, it will be converted to an account ID via the API.
// If the ID is an account ID, it will be returned as is, without performing any API calls.
func ResolveAccountID(ctx context.Context, client avngen.Client, id string) (string, error) {
	if strings.HasPrefix(id, "org") {
		// For organization IDs, use the organization API to get the account ID
		org, err := client.OrganizationGet(ctx, id)
		if err != nil {
			return "", err
		}

		id = org.AccountId
	}

	return id, nil
}
