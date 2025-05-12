package project

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationprojects"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const resourceName = "aiven_organization_project"

func NewOrganizationProjectResource() resource.Resource {
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

func NewOrganizationProjectDatasource() datasource.DataSource {
	return adapter.NewDatasource(
		resourceName,
		new(view),
		datasourceSchema,
		func() adapter.DataModel[dataModel] {
			return new(datasourceDataModel)
		},
	)
}

type view struct {
	client avngen.Client
}

func (vw *view) Configure(client avngen.Client) {
	vw.client = client
}

func (vw *view) Create(ctx context.Context, plan *dataModel) diag.Diagnostics {
	var req organizationprojects.OrganizationProjectsCreateIn
	diags := expandData(ctx, plan, nil, &req, vw.modifyReq(ctx))
	if diags.HasError() {
		return diags
	}

	rsp, err := vw.client.OrganizationProjectsCreate(ctx, plan.OrganizationID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return diags
	}

	// Sets ID field to Read() the resource
	plan.SetID(rsp.OrganizationId, rsp.ProjectId)
	return vw.Read(ctx, plan)
}

func (vw *view) Update(ctx context.Context, plan, state *dataModel) diag.Diagnostics {
	var req organizationprojects.OrganizationProjectsUpdateIn
	diags := expandData(ctx, plan, state, &req, vw.modifyReq(ctx))
	if diags.HasError() {
		return diags
	}

	// OrganizationID is a mutable field, must take it from the state
	rsp, err := vw.client.OrganizationProjectsUpdate(ctx, state.OrganizationID.ValueString(), state.ProjectID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorUpdatingResource, err.Error())
		return diags
	}

	// Sets ID field to Read() the resource.
	// When parent_id is changed, this mutates the ID.
	plan.SetID(rsp.OrganizationId, rsp.ProjectId)
	return vw.Read(ctx, plan)
}

func (vw *view) Delete(ctx context.Context, state *dataModel) diag.Diagnostics {
	err := schemautil.WaitUntilNotFound(ctx, func() error {
		return vw.client.OrganizationProjectsDelete(ctx, state.OrganizationID.ValueString(), state.ProjectID.ValueString())
	})
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}
	return nil
}

func (vw *view) Read(ctx context.Context, state *dataModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := vw.client.OrganizationProjectsGet(ctx, state.OrganizationID.ValueString(), state.ProjectID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	return flattenData(ctx, state, rsp, vw.modifyRsp(ctx, state.ParentID.ValueString()))
}

func (vw *view) modifyReq(ctx context.Context) util.MapModifier[dtoModel] {
	return func(req map[string]any, in *dtoModel) error {
		// Converts OrganizationID to AccountID
		if in.ParentID != nil {
			parentID, err := schemautil.ConvertOrganizationToAccountID(ctx, *in.ParentID, vw.client)
			if err != nil {
				return err
			}
			req["parent_id"] = parentID
		}

		// Converts a list of strings to a list of maps
		if in.TechnicalEmails != nil {
			emails := make([]map[string]any, 0)
			for _, v := range *in.TechnicalEmails {
				emails = append(emails, map[string]any{"email": v})
			}
			req["tech_emails"] = emails
		}

		// Converts key-value pairs to a map
		// Tags is a required field
		tags := make(map[string]string)
		if in.Tag != nil {
			for _, v := range *in.Tag {
				k := *v.Key
				if _, ok := tags[k]; ok {
					return fmt.Errorf("duplicate tag found: %q", k)
				}
				tags[k] = *v.Value
			}
		}
		req["tags"] = tags
		return nil
	}
}

func (vw *view) modifyRsp(ctx context.Context, stateParentID string) util.MapModifier[organizationprojects.OrganizationProjectsGetOut] {
	return func(rsp map[string]any, in *organizationprojects.OrganizationProjectsGetOut) error {
		// Sets CA certificate
		cert, err := vw.client.ProjectKmsGetCA(ctx, in.ProjectId)
		if err != nil {
			return err
		}
		rsp["certificate"] = cert

		// The ParentID in the response is the AccountID,
		// while user could have set the OrganizationID in the state.
		// Overrides it with the state value to avoid an unnecessary plan output.
		if stateParentID != "" && schemautil.IsOrganizationID(stateParentID) {
			rsp["parent_id"] = stateParentID
		}

		// Converts a list of maps to a list of strings
		if _, ok := rsp["tech_emails"]; ok {
			emails := make([]string, 0)
			for _, v := range in.TechEmails {
				emails = append(emails, v.Email)
			}
			rsp["tech_emails"] = emails
		}

		// Converts a map to key-value pairs
		if _, ok := rsp["tags"]; ok {
			tags := make([]map[string]string, 0)
			for k, v := range in.Tags {
				tags = append(tags, map[string]string{"key": k, "value": v})
			}
			rsp["tags"] = tags
		}
		return nil
	}
}
