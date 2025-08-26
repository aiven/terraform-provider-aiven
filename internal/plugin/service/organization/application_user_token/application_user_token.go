package applicationusertoken

import (
	"context"
	"fmt"

	"github.com/aiven/go-client-codegen/handler/applicationuser"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

func NewResource() resource.Resource {
	return adapter.NewResource(aivenName, new(view), newResourceSchema, newResourceModel, composeID())
}

type view struct{ adapter.View }

func (vw *view) Create(ctx context.Context, plan *tfModel) diag.Diagnostics {
	var req applicationuser.ApplicationUserAccessTokenCreateIn
	diags := expandData(ctx, plan, nil, &req)
	if diags.HasError() {
		return diags
	}

	rsp, err := vw.Client.ApplicationUserAccessTokenCreate(ctx, plan.OrganizationID.ValueString(), plan.UserID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return diags
	}

	// Sets ID fields to Read() the resource
	plan.SetID(plan.OrganizationID.ValueString(), plan.UserID.ValueString(), rsp.TokenPrefix)

	// FullToken is only available at creation time
	plan.FullToken = types.StringValue(rsp.FullToken)

	// This information is not available at creation time.
	// Terraform state requires all fields to be set:
	// "After the apply operation, the provider still indicated an unknown value for"
	plan.LastUsedTime = types.StringValue("")
	plan.LastIP = types.StringValue("")
	plan.LastUserAgent = types.StringValue("")
	plan.LastUserAgentHumanReadable = types.StringValue("")
	plan.ExpiryTime = types.StringValue("") // Is nil if `max_age_seconds` is not set
	return vw.Read(ctx, plan)
}

func (vw *view) Update(_ context.Context, _, _ *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(errmsg.SummaryErrorReadingResource, "This resource cannot be updated")
	return diags
}

func (vw *view) Read(ctx context.Context, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := vw.Client.ApplicationUserAccessTokensList(ctx, state.OrganizationID.ValueString(), state.UserID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	prefix := state.TokenPrefix.ValueString()
	for _, token := range rsp {
		if prefix == token.TokenPrefix {
			return flattenData(ctx, state, &token)
		}
	}

	diags.AddError(errmsg.SummaryErrorReadingResource, fmt.Sprintf("%s not found", state.TokenPrefix))
	return diags
}

func (vw *view) Delete(ctx context.Context, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	err := vw.Client.ApplicationUserAccessTokenDelete(ctx, state.OrganizationID.ValueString(), state.UserID.ValueString(), state.TokenPrefix.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}

	return nil
}
