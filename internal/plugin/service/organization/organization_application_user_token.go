package organization

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var (
	_ resource.Resource                = &organizationApplicationUserToken{}
	_ resource.ResourceWithConfigure   = &organizationApplicationUserToken{}
	_ resource.ResourceWithImportState = &organizationApplicationUserToken{}

	_ util.TypeNameable = &organizationApplicationUserToken{}
)

// NewOrganizationApplicationUserToken is a constructor for the organization application user token resource.
func NewOrganizationApplicationUserToken() resource.Resource {
	return &organizationApplicationUserToken{}
}

// organizationApplicationUserToken is the organization application user token resource implementation.
type organizationApplicationUserToken struct {
	// client is the instance of the Aiven client to use.
	client *aiven.Client

	// typeName is the name of the resource type.
	typeName string
}

// organizationApplicationUserTokenModel is the model for the organization application user token resource.
type organizationApplicationUserTokenModel struct {
	// ID is the identifier of the organization application user token.
	ID types.String `tfsdk:"id"`
	// OrganizationID is the identifier of the organization the application user token belongs to.
	OrganizationID types.String `tfsdk:"organization_id"`
	// UserID is the identifier of the application user the token belongs to.
	UserID types.String `tfsdk:"user_id"`
	// FullToken is the full token.
	FullToken types.String `tfsdk:"full_token"`
	// TokenPrefix is the prefix of the token.
	TokenPrefix types.String `tfsdk:"token_prefix"`
	// Description is the description of the token.
	Description types.String `tfsdk:"description"`
	// MaxAgeSeconds is the time the token remains valid since creation (or since last use if extend_when_used
	// is true).
	MaxAgeSeconds types.Number `tfsdk:"max_age_seconds"`
	// ExtendWhenUsed is true to extend token expiration time when token is used. Only applicable if
	// max_age_seconds is specified.
	ExtendWhenUsed types.Bool `tfsdk:"extend_when_used"`
	// Scopes is the scopes this token is restricted to if specified.
	Scopes types.Set `tfsdk:"scopes"`
	// CurrentlyActive is true if API request was made with this access token.
	CurrentlyActive types.Bool `tfsdk:"currently_active"`
	// CreateTime is the time when the token was created.
	CreateTime types.String `tfsdk:"create_time"`
	// CreatedManually is true for tokens explicitly created via the access_tokens API, false for tokens created
	// via login.
	CreatedManually types.Bool `tfsdk:"created_manually"`
	// ExpiryTime is the timestamp when the access token will expire unless extended, if ever.
	ExpiryTime types.String `tfsdk:"expiry_time"`
	// LastIP is the IP address of the last request made with this token.
	LastIP types.String `tfsdk:"last_ip"`
	// LastUsedTime is the timestamp when the access token was last used, if ever.
	LastUsedTime types.String `tfsdk:"last_used_time"`
	// LastUserAgent is the user agent of the last request made with this token.
	LastUserAgent types.String `tfsdk:"last_user_agent"`
	// LastUserAgentHumanReadable is the user agent of the last request made with this token in
	// human-readable format.
	LastUserAgentHumanReadable types.String `tfsdk:"last_user_agent_human_readable"`
	// Timeouts is the configuration for resource-specific timeouts.
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the metadata for the organization application user token resource.
func (r *organizationApplicationUserToken) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization_application_user_token"

	r.typeName = resp.TypeName
}

// TypeName returns the resource type name for the organization application user token resource.
func (r *organizationApplicationUserToken) TypeName() string {
	return r.typeName
}

// Schema defines the schema for the organization application user token resource.
func (r *organizationApplicationUserToken) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = util.GeneralizeSchema(ctx, schema.Schema{
		Description: userconfig.Desc("Creates and manages an application user token. Review the" +
			" [best practices](https://aiven.io/docs/platform/concepts/application-users#security-best-practices) for securing application users and their tokens.").
			Build(),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Compound identifier of the application user token.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the organization the application user belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the application user the token is created for.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"full_token": schema.StringAttribute{
				Description: "Full token.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"token_prefix": schema.StringAttribute{
				Description: "Prefix of the token.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the token.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"max_age_seconds": schema.NumberAttribute{
				Description: "The number of hours after which a token expires. Default session duration is 10 hours.",
				Optional:    true,
				PlanModifiers: []planmodifier.Number{
					numberplanmodifier.RequiresReplace(),
				},
			},
			"extend_when_used": schema.BoolAttribute{
				Description: "Extends the token session duration when the token is used. Only applicable if " +
					"a value is set for `max_age_seconds`.",
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"scopes": schema.SetAttribute{
				Description: "Restricts the scopes for this token.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"currently_active": schema.BoolAttribute{
				Description: "True if the API request was made with this token.",
				Computed:    true,
			},
			"create_time": schema.StringAttribute{
				Description: "Time when the token was created.",
				Computed:    true,
			},
			"created_manually": schema.BoolAttribute{
				Description: "True for tokens explicitly created using the `access_tokens` API. False for tokens created " +
					"when a user logs in.",
				Computed: true,
			},
			"expiry_time": schema.StringAttribute{
				Description: "Timestamp when the access token will expire unless extended.",
				Computed:    true,
			},
			"last_ip": schema.StringAttribute{
				Description: "IP address of the last request made with this token.",
				Computed:    true,
			},
			"last_used_time": schema.StringAttribute{
				Description: "Timestamp when the access token was last used.",
				Computed:    true,
			},
			"last_user_agent": schema.StringAttribute{
				Description: "User agent of the last request made with this token.",
				Computed:    true,
			},
			"last_user_agent_human_readable": schema.StringAttribute{
				Description: "User agent of the last request made with this token in human-readable format.",
				Computed:    true,
			},
		},
	})
}

// Configure sets up the organization application user token resource.
func (r *organizationApplicationUserToken) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*aiven.Client)
	if !ok {
		resp.Diagnostics = util.DiagErrorUnexpectedProviderDataType(resp.Diagnostics, req.ProviderData)

		return
	}

	r.client = client
}

// fillModel fills the organization application user token resource model from the Aiven API.
func (r *organizationApplicationUserToken) fillModel(
	ctx context.Context,
	model *organizationApplicationUserTokenModel,
) (err error) {
	tokens, err := r.client.OrganizationApplicationUserHandler.ListTokens(
		ctx,
		model.OrganizationID.ValueString(),
		model.UserID.ValueString(),
	)
	if err != nil {
		return err
	}

	var token *aiven.ApplicationUserTokenInfo

	for _, t := range tokens.Tokens {
		if t.TokenPrefix == model.TokenPrefix.ValueString() {
			token = &t
			break
		}
	}

	if token == nil {
		return fmt.Errorf(errmsg.AivenResourceNotFound, r.TypeName(), model.ID.ValueString())
	}

	model.Description = util.ValueOrDefault(token.Description, types.StringNull())

	model.MaxAgeSeconds = types.NumberValue(util.ToBigFloat(token.MaxAgeSeconds))

	model.ExtendWhenUsed = util.ValueOrDefault(token.ExtendWhenUsed, types.BoolNull())

	if token.Scopes != nil {
		scopes, diags := types.SetValueFrom(ctx, types.StringType, *token.Scopes)
		if diags.HasError() {
			return fmt.Errorf(errmsg.UnableToSetValueFrom, *token.Scopes)
		}

		model.Scopes = scopes
	}

	model.CurrentlyActive = types.BoolValue(token.CurrentlyActive)

	model.CreateTime = util.ValueOrDefault(token.CreateTime, types.StringNull())

	model.CreatedManually = types.BoolValue(token.CreatedManually)

	model.ExpiryTime = util.ValueOrDefault(token.ExpiryTime, types.StringNull())

	model.LastIP = types.StringValue(token.LastIP)

	model.LastUsedTime = util.ValueOrDefault(token.LastUsedTime, types.StringNull())

	model.LastUserAgent = util.ValueOrDefault(token.LastUserAgent, types.StringNull())

	model.LastUserAgentHumanReadable = util.ValueOrDefault(token.LastUserAgentHumanReadable, types.StringNull())

	model.TokenPrefix = types.StringValue(token.TokenPrefix)

	return err
}

// Create creates an organization application user token resource.
func (r *organizationApplicationUserToken) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan organizationApplicationUserTokenModel

	if !util.PlanStateToModel(ctx, &req.Plan, &plan, &resp.Diagnostics) {
		return
	}

	aivenReq := &aiven.ApplicationUserTokenCreateRequest{}

	if !plan.Description.IsNull() {
		aivenReq.Description = util.Ref(plan.Description.ValueString())
	}

	if !plan.MaxAgeSeconds.IsNull() {
		aivenReq.MaxAgeSeconds = util.Ref((int)(util.First(plan.MaxAgeSeconds.ValueBigFloat().Int64())))
	}

	if !plan.ExtendWhenUsed.IsNull() {
		aivenReq.ExtendWhenUsed = util.Ref(plan.ExtendWhenUsed.ValueBool())
	}

	if !plan.Scopes.IsNull() {
		scopes := make([]string, 0, len(plan.Scopes.Elements()))

		diags := plan.Scopes.ElementsAs(ctx, &scopes, false)
		if diags.HasError() {
			resp.Diagnostics = append(resp.Diagnostics, diags...)

			return
		}

		aivenReq.Scopes = &scopes
	}

	token, err := r.client.OrganizationApplicationUserHandler.CreateToken(
		ctx,
		plan.OrganizationID.ValueString(),
		plan.UserID.ValueString(),
		*aivenReq,
	)
	if err != nil {
		resp.Diagnostics = util.DiagErrorCreatingResource(resp.Diagnostics, r, err)

		return
	}

	plan.ID = types.StringValue(
		util.ComposeID(plan.OrganizationID.ValueString(), plan.UserID.ValueString(), token.TokenPrefix),
	)

	plan.FullToken = types.StringValue(token.FullToken)

	plan.TokenPrefix = types.StringValue(token.TokenPrefix)

	err = r.fillModel(ctx, &plan)
	if err != nil {
		resp.Diagnostics = util.DiagErrorCreatingResource(resp.Diagnostics, r, err)

		return
	}

	if !util.ModelToPlanState(ctx, plan, &resp.State, &resp.Diagnostics) {
		return
	}
}

// Read reads an organization application user token resource.
func (r *organizationApplicationUserToken) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state organizationApplicationUserTokenModel

	if !util.PlanStateToModel(ctx, &req.State, &state, &resp.Diagnostics) {
		return
	}

	err := r.fillModel(ctx, &state)
	if err != nil {
		resp.Diagnostics = util.DiagErrorReadingResource(resp.Diagnostics, r, err)

		return
	}

	if !util.ModelToPlanState(ctx, state, &resp.State, &resp.Diagnostics) {
		return
	}
}

// Update updates an organization resource.
func (r *organizationApplicationUserToken) Update(
	_ context.Context,
	_ resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	resp.Diagnostics = util.DiagErrorUpdatingResourceNotSupported(resp.Diagnostics, r)
}

// Delete deletes an organization application user token resource.
func (r *organizationApplicationUserToken) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state organizationApplicationUserTokenModel

	if !util.PlanStateToModel(ctx, &req.State, &state, &resp.Diagnostics) {
		return
	}

	if err := r.client.OrganizationApplicationUserHandler.DeleteToken(
		ctx,
		state.OrganizationID.ValueString(),
		state.UserID.ValueString(),
		state.TokenPrefix.ValueString(),
	); err != nil {
		resp.Diagnostics = util.DiagErrorDeletingResource(resp.Diagnostics, r, err)

		return
	}
}

// ImportState handles resource's state import requests.
func (r *organizationApplicationUserToken) ImportState(
	_ context.Context,
	_ resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resp.Diagnostics = util.DiagErrorImportingResourceNotSupported(resp.Diagnostics, r)
}
