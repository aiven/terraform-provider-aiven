package usergrouplist

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/usergroup"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

func NewDataSource() datasource.DataSource {
	return adapter.NewDataSource(adapter.DataSourceOptions[*datasourceModel, tfModel]{
		TypeName: typeName,
		Schema:   datasourceSchema,
		Read:     readUserGroupList,
	})
}

func readUserGroupList(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := client.UserGroupsList(ctx, state.OrganizationID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	list := organizationUserGroupList{UserGroups: rsp}
	return flattenData(ctx, state, &list)
}

type organizationUserGroupList struct {
	UserGroups []usergroup.UserGroupOut `json:"user_groups"`
}
