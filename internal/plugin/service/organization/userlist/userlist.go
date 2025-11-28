package userlist

import (
	"context"
	"fmt"
	"strings"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

func planModifier(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if state.Name.ValueString() != "" {
		id, err := GetOrganizationByName(ctx, client, state.Name.ValueString())
		if err != nil {
			diags.Append(errmsg.FromError("GetOrganizationByName Error", err))
			return diags
		}
		state.ID = types.StringValue(id)
	}
	return diags
}

func GetOrganizationByName(ctx context.Context, client avngen.Client, name string) (string, error) {
	ids := make([]string, 0)
	list, err := client.UserOrganizationsList(ctx)
	if err != nil {
		return "", err
	}

	for _, o := range list {
		// Organization name is not unique
		if o.OrganizationName == name {
			ids = append(ids, o.OrganizationId)
		}
	}

	switch len(ids) {
	case 0:
		return "", fmt.Errorf("organization %q not found", name)
	case 1:
		return ids[0], nil
	}
	return "", fmt.Errorf("multiple organizations %q found, ids: %s", name, strings.Join(ids, ", "))
}
