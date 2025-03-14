package address

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organization"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/diagnostics"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	providertypes "github.com/aiven/terraform-provider-aiven/internal/plugin/types"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/validation"
)

var (
	_ resource.Resource                = &organizationAddressResource{}
	_ resource.ResourceWithConfigure   = &organizationAddressResource{}
	_ resource.ResourceWithImportState = &organizationAddressResource{}
)

// NewOrganizationAddressResource is a constructor for the organization resource.
func NewOrganizationAddressResource() resource.Resource {
	return &organizationAddressResource{}
}

// organizationAddressResource is the organization resource implementation.
type organizationAddressResource struct {
	// client is the instance of the Aiven client to use.
	client avngen.Client

	// diag is the diagnostics helper.
	diag *diagnostics.DiagnosticsHelper
}

// organizationAddressResourceModel is the model for the organization address resource.
type organizationAddressResourceModel struct {
	OrganizationAddressModel

	// Timeouts is the configuration for resource-specific timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// TypeName returns the resource type name.
func (r *organizationAddressResource) TypeName() string {
	return "organization_address_resource"
}

// Metadata returns the metadata for the organization resource.
func (r *organizationAddressResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization_address"
}

// Schema defines the schema for the organization address resource.
func (r *organizationAddressResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceSchema := ResourceSchema()

	// Add timeouts attribute
	resourceSchema["timeouts"] = timeouts.Attributes(ctx, timeouts.Opts{
		Create: true,
		Read:   true,
		Update: true,
		Delete: true,
	})

	resp.Schema = schema.Schema{
		Description: "Creates and manages an organization address.",
		Attributes:  resourceSchema,
	}
}

// Configure sets up the organization resource.
func (r *organizationAddressResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	p, ok := req.ProviderData.(providertypes.AivenClientProvider)
	if !ok {
		resp.Diagnostics.AddError(
			errmsg.SummaryUnexpectedProviderDataType,
			fmt.Sprintf(errmsg.DetailUnexpectedProviderDataType, req.ProviderData),
		)
		return
	}

	r.client = p.GetGenClient()
	r.diag = diagnostics.NewDiagnosticsHelper(r.TypeName())
}

// validateRequiredFields validates that all required fields are set.
func validateRequiredFields(
	ctx context.Context,
	model *organizationAddressResourceModel,
	diags *diag.Diagnostics,
	diagHelper *diagnostics.DiagnosticsHelper,
) {
	// Validate organization_id
	validation.ValidateRequiredStringField(model.OrganizationID, "organization_id", diags, diagHelper)

	// Validate address_lines
	validation.ValidateRequiredListField(ctx, model.AddressLines, "address_lines", diags, diagHelper)

	// Validate city
	validation.ValidateRequiredStringField(model.City, "city", diags, diagHelper)

	// Validate country_code
	validation.ValidateRequiredStringField(model.CountryCode, "country_code", diags, diagHelper)
}

// handleCompanyName handles the special case for company_name to prevent drift detection.
// This is a workaround that should be removed once the API is fixed.
func handleCompanyName(companyName string) basetypes.StringValue {
	// WORKAROUND: Handle empty company_name specially to prevent drift detection
	// The API returns an empty string ("") for company_name when it's not set,
	// but Terraform represents unset optional values as null.
	// This workaround should be removed once the API is fixed to return null for unset fields.
	if companyName == "" {
		return types.StringNull()
	}
	return types.StringValue(companyName)
}

// Create creates an organization address resource.
func (r *organizationAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationAddressResourceModel

	// Get plan values
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate required fields
	validateRequiredFields(ctx, &plan, &resp.Diagnostics, r.diag)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get address lines
	lines, err := getAddressLines(ctx, plan.OrganizationAddressModel)
	if err != nil {
		resp.Diagnostics.AddError("Error getting address lines", err.Error())
		return
	}

	// Prepare the request body
	createReq := &organization.OrganizationAddressCreateIn{
		AddressLines: lines,
		City:         plan.City.ValueString(),
		CountryCode:  plan.CountryCode.ValueString(),
		CompanyName:  plan.CompanyName.ValueStringPointer(),
		State:        plan.State.ValueStringPointer(),
		ZipCode:      plan.ZipCode.ValueStringPointer(),
	}

	// Create the address
	address, err := r.client.OrganizationAddressCreate(ctx, plan.OrganizationID.ValueString(), createReq)
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "creating", err)
		return
	}

	// Set state
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.OrganizationID.ValueString(), address.AddressId))
	plan.AddressID = types.StringValue(address.AddressId)

	// Convert address lines to types.List
	addressLines, diags := types.ListValueFrom(ctx, types.StringType, address.AddressLines)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.AddressLines = addressLines

	plan.City = types.StringPointerValue(address.City)
	plan.CompanyName = handleCompanyName(address.CompanyName)
	plan.CountryCode = types.StringValue(address.CountryCode)
	plan.State = types.StringPointerValue(address.State)
	plan.ZipCode = types.StringPointerValue(address.ZipCode)
	plan.CreateTime = types.StringValue(address.CreateTime.String())
	plan.UpdateTime = types.StringValue(address.UpdateTime.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads an organization address resource.
func (r *organizationAddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationAddressResourceModel

	// Get state values
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the address
	address, err := r.client.OrganizationAddressGet(ctx, state.OrganizationID.ValueString(), state.AddressID.ValueString())
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "reading", err)
		return
	}

	// Set state
	state.ID = types.StringValue(fmt.Sprintf("%s/%s", state.OrganizationID.ValueString(), address.AddressId))
	state.AddressID = types.StringValue(address.AddressId)

	// Convert address lines to types.List
	addressLines, diags := types.ListValueFrom(ctx, types.StringType, address.AddressLines)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.AddressLines = addressLines

	state.City = types.StringPointerValue(address.City)
	state.CompanyName = handleCompanyName(address.CompanyName)
	state.CountryCode = types.StringValue(address.CountryCode)
	state.State = types.StringPointerValue(address.State)
	state.ZipCode = types.StringPointerValue(address.ZipCode)
	state.CreateTime = types.StringValue(address.CreateTime.String())
	state.UpdateTime = types.StringValue(address.UpdateTime.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates an organization address resource.
func (r *organizationAddressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and state
	var plan, state organizationAddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate required fields
	validateRequiredFields(ctx, &plan, &resp.Diagnostics, r.diag)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get address lines
	lines, err := getAddressLines(ctx, plan.OrganizationAddressModel)
	if err != nil {
		resp.Diagnostics.AddError("Error getting address lines", err.Error())
		return
	}

	// Prepare the request body
	updateReq := &organization.OrganizationAddressUpdateIn{
		AddressLines: &lines,
		City:         plan.City.ValueStringPointer(),
		CountryCode:  plan.CountryCode.ValueStringPointer(),
		State:        plan.State.ValueStringPointer(),
		ZipCode:      plan.ZipCode.ValueStringPointer(),
	}

	// WORKAROUND: Handle company_name specially to prevent drift detection
	// The API returns an empty string ("") for company_name when it's not set,
	// but Terraform represents unset optional values as null.
	// This workaround should be removed once the API is fixed to return null for unset fields.
	if plan.CompanyName.IsNull() {
		// When company_name is null in the plan, explicitly set an empty string in the API request
		emptyString := ""
		updateReq.CompanyName = &emptyString
	} else {
		updateReq.CompanyName = plan.CompanyName.ValueStringPointer()
	}

	// Update the address
	address, err := r.client.OrganizationAddressUpdate(ctx, plan.OrganizationID.ValueString(), state.AddressID.ValueString(), updateReq)
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "updating", err)
		return
	}

	// Set state
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.OrganizationID.ValueString(), address.AddressId))
	plan.AddressID = types.StringValue(address.AddressId)

	// Convert address lines to types.List
	addressLines, diags := types.ListValueFrom(ctx, types.StringType, address.AddressLines)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.AddressLines = addressLines

	plan.City = types.StringPointerValue(address.City)
	plan.CompanyName = handleCompanyName(address.CompanyName)
	plan.CountryCode = types.StringValue(address.CountryCode)
	plan.State = types.StringPointerValue(address.State)
	plan.ZipCode = types.StringPointerValue(address.ZipCode)
	plan.CreateTime = types.StringValue(address.CreateTime.String())
	plan.UpdateTime = types.StringValue(address.UpdateTime.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes an organization address resource.
func (r *organizationAddressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationAddressResourceModel

	// Get state values
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the address
	err := r.client.OrganizationAddressDelete(ctx, state.OrganizationID.ValueString(), state.AddressID.ValueString())
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "deleting", err)
		return
	}
}

// ImportState imports an organization address resource.
func (r *organizationAddressResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Check if import ID is valid
	parts, err := validation.ValidateImportID(req.ID, "organization_id/address_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}

	// Set the import attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("address_id"), parts[1])...)
}
