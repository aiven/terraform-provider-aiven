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

func NewResource() resource.Resource {
	return adapter.NewResource(adapter.ResourceOptions[*resourceModel, tfModel]{
		TypeName: typeName,
		IDFields: idFields(),
		Schema:   resourceSchema,
		Read:     readOrganization,
		Create:   createOrganization,
		Update:   updateOrganization,
		Delete:   deleteOrganization,
	})
}

func NewDatasource() datasource.DataSource {
	return adapter.NewDatasource(adapter.DatasourceOptions[*datasourceModel, tfModel]{
		TypeName:         typeName,
		Schema:           datasourceSchemaPatched,
		Read:             readOrganization,
		ConfigValidators: datasourceConfigValidators,
	})
}

// datasourceSchemaPatched turns "id" and "name" into optional attributes so it can be found either by ID or name.
func datasourceSchemaPatched(ctx context.Context) dataschema.Schema {
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

func datasourceConfigValidators(ctx context.Context, client avngen.Client) []datasource.ConfigValidator {
	validatable := path.Expressions{
		path.MatchRoot("id"),
		path.MatchRoot("name"),
	}

	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(validatable...),
		datasourcevalidator.AtLeastOneOf(validatable...),
	}
}

func createOrganization(ctx context.Context, client avngen.Client, plan *tfModel) diag.Diagnostics {
	var req account.AccountCreateIn
	diags := expandData(ctx, plan, nil, &req)
	if diags.HasError() {
		return diags
	}

	rsp, err := client.AccountCreate(ctx, &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return diags
	}

	// Sets ID field to Read() the resource
	plan.SetID(rsp.OrganizationId)
	return readOrganization(ctx, client, plan)
}

func updateOrganization(ctx context.Context, client avngen.Client, plan, state, _ *tfModel) diag.Diagnostics {
	var req account.AccountUpdateIn
	diags := expandData(ctx, plan, state, &req)
	if diags.HasError() {
		return diags
	}

	accountID, err := getAccountID(ctx, client, state)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorUpdatingResource, err.Error())
		return diags
	}

	rsp, err := client.AccountUpdate(ctx, accountID, &req)
	if err != nil {
		diags.AddError(
			errmsg.SummaryErrorUpdatingResource,
			fmt.Sprintf("%q: %s", accountID, err.Error()),
		)
		return diags
	}

	// Sets ID field to Read() the resource
	plan.SetID(rsp.OrganizationId)
	return readOrganization(ctx, client, plan)
}

func deleteOrganization(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	accountID, err := getAccountID(ctx, client, state)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}

	err = client.AccountDelete(ctx, accountID)
	if err != nil {
		diags.AddError(
			errmsg.SummaryErrorDeletingResource,
			fmt.Sprintf("%q: %s", accountID, err.Error()),
		)
		return diags
	}

	return nil
}

func readOrganization(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	accountID, err := getAccountID(ctx, client, state)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	rsp, err := client.AccountGet(ctx, accountID)
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
func getAccountID(ctx context.Context, client avngen.Client, state *tfModel) (string, error) {
	id := state.ID.ValueString()
	switch {
	case strings.HasPrefix(id, "org"):
		// This is an organization ID (org123456) format
		org, err := client.OrganizationGet(ctx, id)
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

	list, err := client.AccountList(ctx)
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
