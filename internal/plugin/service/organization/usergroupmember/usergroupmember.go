package usergroupmember

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/usergroup"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

func createView(ctx context.Context, client avngen.Client, plan, _ *tfModel) diag.Diagnostics {
	return upsert(ctx, client, plan, usergroup.OperationTypeAddMembers)
}

func deleteView(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	return upsert(ctx, client, state, usergroup.OperationTypeRemoveMembers)
}

func upsert(ctx context.Context, client avngen.Client, state *tfModel, operation usergroup.OperationType) diag.Diagnostics {
	err := client.UserGroupMembersUpdate(
		ctx,
		state.OrganizationID.ValueString(),
		state.GroupID.ValueString(),
		&usergroup.UserGroupMembersUpdateIn{
			Operation: operation,
			MemberIds: []string{state.UserID.ValueString()},
		},
	)

	var diags diag.Diagnostics
	if err != nil {
		summary := errmsg.SummaryErrorCreatingResource
		if operation == usergroup.OperationTypeRemoveMembers {
			summary = errmsg.SummaryErrorDeletingResource
		}

		diags.Append(errmsg.FromError(summary, err))
		return diags
	}

	return diags
}
