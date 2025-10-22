package applicationusertoken

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/applicationuser"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

func NewResource() resource.Resource {
	return adapter.NewResource(adapter.ResourceOptions[*resourceModel, tfModel]{
		TypeName: aivenName,
		IDFields: idFields(),
		Schema:   newResourceSchema,
		Read:     readApplicationUserToken,
		Create:   createApplicationUserToken,
		Delete:   deleteApplicationUserToken,
	})
}

func createApplicationUserToken(ctx context.Context, client avngen.Client, plan *tfModel) diag.Diagnostics {
	var req applicationuser.ApplicationUserAccessTokenCreateIn
	diags := expandData(ctx, plan, nil, &req)
	if diags.HasError() {
		return diags
	}

	rsp, err := client.ApplicationUserAccessTokenCreate(ctx, plan.OrganizationID.ValueString(), plan.UserID.ValueString(), &req)
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
	return readApplicationUserToken(ctx, client, plan)
}

func readApplicationUserToken(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := client.ApplicationUserAccessTokensList(ctx, state.OrganizationID.ValueString(), state.UserID.ValueString())
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

func deleteApplicationUserToken(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	err := client.ApplicationUserAccessTokenDelete(ctx, state.OrganizationID.ValueString(), state.UserID.ValueString(), state.TokenPrefix.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}

	return nil
}
