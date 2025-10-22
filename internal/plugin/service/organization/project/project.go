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

func NewResource() resource.Resource {
	return adapter.NewResource(adapter.ResourceOptions[*resourceModel, tfModel]{
		TypeName: aivenName,
		IDFields: composeID(),
		Schema:   newResourceSchema,
		Read:     readProject,
		Create:   createProject,
		Update:   updateProject,
		Delete:   deleteProject,
	})
}

func NewDatasource() datasource.DataSource {
	return adapter.NewDatasource(adapter.DatasourceOptions[*datasourceModel, tfModel]{
		TypeName: aivenName,
		Schema:   newDatasourceSchema,
		Read:     readProject,
	})
}

func createProject(ctx context.Context, client avngen.Client, plan *tfModel) diag.Diagnostics {
	var req organizationprojects.OrganizationProjectsCreateIn
	diags := expandData(ctx, plan, nil, &req, modifyReq(ctx, client))
	if diags.HasError() {
		return diags
	}

	rsp, err := client.OrganizationProjectsCreate(ctx, plan.OrganizationID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return diags
	}

	// Sets ID field to Read() the resource
	plan.SetID(rsp.OrganizationId, rsp.ProjectId)
	return readProject(ctx, client, plan)
}

func updateProject(ctx context.Context, client avngen.Client, plan, state, _ *tfModel) diag.Diagnostics {
	var req organizationprojects.OrganizationProjectsUpdateIn
	diags := expandData(ctx, plan, state, &req, modifyReq(ctx, client))
	if diags.HasError() {
		return diags
	}

	// OrganizationID is a mutable field, must take it from the state
	rsp, err := client.OrganizationProjectsUpdate(ctx, state.OrganizationID.ValueString(), state.ProjectID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorUpdatingResource, err.Error())
		return diags
	}

	// Sets ID field to Read() the resource.
	// When parent_id is changed, this mutates the ID.
	plan.SetID(rsp.OrganizationId, rsp.ProjectId)
	return readProject(ctx, client, plan)
}

func deleteProject(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	err := schemautil.WaitUntilNotFound(ctx, func() error {
		return client.OrganizationProjectsDelete(ctx, state.OrganizationID.ValueString(), state.ProjectID.ValueString())
	})
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}
	return nil
}

func readProject(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := client.OrganizationProjectsGet(ctx, state.OrganizationID.ValueString(), state.ProjectID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	return flattenData(ctx, state, rsp, modifyRsp(ctx, client, state.ParentID.ValueString()))
}

func modifyReq(ctx context.Context, client avngen.Client) util.MapModifier[apiModel] {
	return func(req util.RawMap, in *apiModel) error {
		// Converts OrganizationID to AccountID
		if in.ParentID != nil {
			parentID, err := schemautil.ConvertOrganizationToAccountID(ctx, *in.ParentID, client)
			if err != nil {
				return err
			}
			err = req.Set(parentID, "parent_id")
			if err != nil {
				return err
			}
		}

		// Converts a list of strings to a list of maps
		if in.TechnicalEmails != nil {
			emails := make([]map[string]any, 0)
			for _, v := range *in.TechnicalEmails {
				emails = append(emails, map[string]any{"email": v})
			}
			err := req.Set(emails, "tech_emails")
			if err != nil {
				return err
			}
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
		return req.Set(tags, "tags")
	}
}

func modifyRsp(ctx context.Context, client avngen.Client, stateParentID string) util.MapModifier[organizationprojects.OrganizationProjectsGetOut] {
	return func(rsp util.RawMap, in *organizationprojects.OrganizationProjectsGetOut) error {
		// Sets CA certificate
		cert, err := client.ProjectKmsGetCA(ctx, in.ProjectId)
		if err != nil {
			return err
		}

		err = rsp.Set(cert, "certificate")
		if err != nil {
			return err
		}

		// The ParentID in the response is the AccountID,
		// while user could have set the OrganizationID in the state.
		// Overrides it with the state value to avoid an unnecessary plan output.
		if stateParentID != "" && schemautil.IsOrganizationID(stateParentID) {
			err = rsp.Set(stateParentID, "parent_id")
			if err != nil {
				return err
			}
		}

		// Converts a list of maps to a list of strings
		if len(in.TechEmails) > 0 {
			emails := make([]string, 0)
			for _, v := range in.TechEmails {
				emails = append(emails, v.Email)
			}
			err = rsp.Set(emails, "tech_emails")
			if err != nil {
				return err
			}
		}

		// Converts a map to list of key-value pairs
		// Deletes the field first to avoid having an empty map where a list is expected.
		err = rsp.Delete("tags")
		if err != nil {
			return err
		}

		if len(in.Tags) > 0 {
			tags := make([]map[string]string, 0)
			for k, v := range in.Tags {
				tags = append(tags, map[string]string{"key": k, "value": v})
			}
			err = rsp.Set(tags, "tags")
			if err != nil {
				return err
			}
		}
		return nil
	}
}
