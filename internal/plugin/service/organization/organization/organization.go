package organization

import (
	"context"
	"fmt"
	"strings"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/account"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

const resourceName = "aiven_organization"

func NewOrganizationResource() resource.Resource {
	return adapter.NewResource(
		resourceName,
		new(view),
		resourceSchema,
		func() adapter.DataModel[dataModel] {
			return new(resourceDataModel)
		},
		idFields(),
	)
}

func NewOrganizationDatasource() datasource.DataSource {
	return adapter.NewDatasource(
		resourceName,
		new(view),
		getPatchedDataSchema,
		func() adapter.DataModel[dataModel] {
			return new(datasourceDataModel)
		},
	)
}

func getPatchedDataSchema(ctx context.Context) dataschema.Schema {
	patch := dataschema.Schema{
		Attributes: map[string]dataschema.Attribute{
			"id": dataschema.StringAttribute{
				MarkdownDescription: "ID of the organization.",
				Optional:            true,
			},
			"name": dataschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Name of the organization.",
			},
		},
	}
	result := datasourceSchema(ctx)
	for k, v := range patch.Attributes {
		result.Attributes[k] = v
	}
	return result
}

var _ adapter.DatViewValidators[dataModel] = (*view)(nil)

type view struct {
	client avngen.Client
}

func (vw *view) Configure(client avngen.Client) {
	vw.client = client
}

func (vw *view) DatValidators(_ context.Context) []datasource.ConfigValidator {
	validatable := path.Expressions{
		path.MatchRoot("id"),
		path.MatchRoot("name"),
	}

	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(validatable...),
		datasourcevalidator.AtLeastOneOf(validatable...),
	}
}

func (vw *view) Create(ctx context.Context, plan *dataModel) diag.Diagnostics {
	var req account.AccountCreateIn
	diags := expandData(ctx, plan, nil, &req)
	if diags.HasError() {
		return diags
	}

	rsp, err := vw.client.AccountCreate(ctx, &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return diags
	}

	// Sets ID field to Read() the resource
	plan.SetID(rsp.OrganizationId)
	return vw.Read(ctx, plan)
}

func (vw *view) Update(ctx context.Context, plan, state *dataModel) diag.Diagnostics {
	var req account.AccountUpdateIn
	diags := expandData(ctx, plan, state, &req)
	if diags.HasError() {
		return diags
	}

	accountID, err := vw.getAccountID(ctx, state)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorUpdatingResource, err.Error())
		return diags
	}

	rsp, err := vw.client.AccountUpdate(ctx, accountID, &req)
	if err != nil {
		diags.AddError(
			errmsg.SummaryErrorUpdatingResource,
			fmt.Sprintf("%q: %s", accountID, err.Error()),
		)
		return diags
	}

	// Sets ID field to Read() the resource
	plan.SetID(rsp.OrganizationId)
	return vw.Read(ctx, plan)
}

func (vw *view) Delete(ctx context.Context, state *dataModel) diag.Diagnostics {
	var diags diag.Diagnostics
	accountID, err := vw.getAccountID(ctx, state)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}

	err = vw.client.AccountDelete(ctx, accountID)
	if err != nil {
		diags.AddError(
			errmsg.SummaryErrorDeletingResource,
			fmt.Sprintf("%q: %s", accountID, err.Error()),
		)
		return diags
	}

	return nil
}

func (vw *view) Read(ctx context.Context, state *dataModel) diag.Diagnostics {
	var diags diag.Diagnostics
	accountID, err := vw.getAccountID(ctx, state)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	rsp, err := vw.client.AccountGet(ctx, accountID)
	if err != nil {
		diags.AddError(
			errmsg.SummaryErrorReadingResource,
			fmt.Sprintf("%q: %s", accountID, err.Error()),
		)
		return diags
	}

	return flattenData(ctx, state, rsp)
}

// getAccountID is a helper function that returns the account ID to use for API calls.
func (vw *view) getAccountID(ctx context.Context, state *dataModel) (string, error) {
	id := state.ID.ValueString()
	switch {
	case strings.HasPrefix(id, "org"):
		// This is an organization ID (org123456) format
		org, err := vw.client.OrganizationGet(ctx, id)
		if err != nil {
			return "", fmt.Errorf("faild to resolve organization id %q", id)
		}
		return org.AccountId, nil
	case id != "":
		// This is likely an account ID (a123456) format.
		return id, nil
	}

	orgName := state.Name.ValueString()
	if orgName == "" {
		return "", fmt.Errorf("no Organization ID or name provided")
	}

	list, err := vw.client.AccountList(ctx)
	if err != nil {
		return "", err
	}

	for _, org := range list {
		if org.AccountName == orgName {
			return org.AccountId, nil
		}
	}
	return "", fmt.Errorf("can't find organization with name %q", orgName)
}
