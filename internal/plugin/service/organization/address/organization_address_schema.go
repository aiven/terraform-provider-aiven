package address

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OrganizationAddressModel is the common model for both resource and data source
type OrganizationAddressModel struct {
	// ID is a composite of organization_id and address_id
	ID types.String `tfsdk:"id"`

	// OrganizationID is the identifier of the organization
	OrganizationID types.String `tfsdk:"organization_id"`

	// AddressID is the identifier of the address
	AddressID types.String `tfsdk:"address_id"`

	// AddressLines is the array of address lines
	AddressLines types.List `tfsdk:"address_lines"`

	// City is the city name
	City types.String `tfsdk:"city"`

	// CompanyName is the name of the company
	CompanyName types.String `tfsdk:"company_name"`

	// CountryCode is the country code
	CountryCode types.String `tfsdk:"country_code"`

	// State is the state name
	State types.String `tfsdk:"state"`

	// ZipCode is the zip code
	ZipCode types.String `tfsdk:"zip_code"`

	// CreateTime is the timestamp of the creation
	CreateTime types.String `tfsdk:"create_time"`

	// UpdateTime is the timestamp of the last update
	UpdateTime types.String `tfsdk:"update_time"`
}

// ResourceSchema returns the schema for the organization address resource
func ResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Resource ID, a composite of organization_id and address_id.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"organization_id": schema.StringAttribute{
			Description: "ID of the organization.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"address_id": schema.StringAttribute{
			Description: "ID of the address.",
			Computed:    true,
		},
		"address_lines": schema.ListAttribute{
			Description: "Array of address lines.",
			Required:    true,
			ElementType: types.StringType,
		},
		"city": schema.StringAttribute{
			Description: "City name.",
			Required:    true,
		},
		"company_name": schema.StringAttribute{
			Description: "Name of the company.",
			Optional:    true,
		},
		"country_code": schema.StringAttribute{
			Description: "Country code.",
			Required:    true,
		},
		"state": schema.StringAttribute{
			Description: "State name.",
			Optional:    true,
		},
		"zip_code": schema.StringAttribute{
			Description: "Zip code.",
			Optional:    true,
		},
		"create_time": schema.StringAttribute{
			Description: "Timestamp of the creation.",
			Computed:    true,
		},
		"update_time": schema.StringAttribute{
			Description: "Timestamp of the last update.",
			Computed:    true,
		},
	}
}

// getAddressLines retrieves the address lines from the model.
func getAddressLines(ctx context.Context, model OrganizationAddressModel) ([]string, error) {
	var lines []string
	if model.AddressLines.IsNull() || model.AddressLines.IsUnknown() {
		return lines, nil
	}
	diags := model.AddressLines.ElementsAs(ctx, &lines, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to get address lines: %v", diags)
	}
	return lines, nil
}
