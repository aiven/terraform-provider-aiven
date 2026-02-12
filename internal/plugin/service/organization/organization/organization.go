package organization

import (
	"context"
	"fmt"
	"strings"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/account"
	"github.com/aiven/go-client-codegen/handler/organization"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func NewResource() resource.Resource {
	return adapter.NewResource(adapter.ResourceOptions{
		TypeName:       typeName,
		IDFields:       idFields(),
		Schema:         resourceSchema,
		SchemaInternal: resourceSchemaInternal(),
		RefreshState:   true,
		Read:           readOrganization,
		Create:         createOrganization,
		Update:         updateOrganization,
		Delete:         deleteOrganization,
	})
}

func NewDataSource() datasource.DataSource {
	return adapter.NewDataSource(adapter.DataSourceOptions{
		TypeName:         typeName,
		IDFields:         idFields(),
		Schema:           datasourceSchema,
		SchemaInternal:   datasourceSchemaInternal(),
		Read:             readOrganization,
		ConfigValidators: datasourceConfigValidators,
	})
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

func createOrganization(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	var req account.AccountCreateIn
	err := d.Expand(&req, adapter.RenameFields(map[string]string{"name": "account_name"}))
	if err != nil {
		return err
	}
	rsp, err := client.AccountCreate(ctx, &req)
	if err != nil {
		return err
	}
	return d.Flatten(rsp, adapter.RenameFields(map[string]string{"organization_id": "id", "account_name": "name"}))
}

func updateOrganization(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	var req account.AccountUpdateIn
	err := d.Expand(&req, adapter.RenameFields(map[string]string{"name": "account_name"}))
	if err != nil {
		return err
	}

	accountID, err := getAccountID(ctx, client, d)
	if err != nil {
		return err
	}

	rsp, err := client.AccountUpdate(ctx, accountID, &req)
	if err != nil {
		return err
	}

	return d.Flatten(rsp, adapter.RenameFields(map[string]string{"organization_id": "id", "account_name": "name"}))
}

func deleteOrganization(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	var err error
	orgID := d.ID()
	if strings.HasPrefix(orgID, "org") {
		return client.OrganizationDelete(ctx, orgID, organization.OrganizationDeleteRecursive(true))
	}

	orgID, err = getAccountID(ctx, client, d)
	if err != nil {
		return err
	}
	return client.AccountDelete(ctx, orgID)
}

func readOrganization(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	accountID, err := getAccountID(ctx, client, d)
	if err != nil {
		return err
	}

	rsp, err := client.AccountGet(ctx, accountID)
	if err != nil {
		return err
	}
	return d.Flatten(rsp, adapter.RenameFields(map[string]string{
		"organization_id": "id",
		"account_name":    "name",
	}))
}

// getAccountID is a helper function that returns the account ID to use for API calls.
func getAccountID(ctx context.Context, client avngen.Client, d adapter.ResourceData) (string, error) {
	id := d.ID()
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

	orgName := d.Get("name").(string)
	if orgName == "" {
		return "", fmt.Errorf("no Organization ID or name provided")
	}

	var err error
	var list []account.AccountOut
	err = retry.Do(
		func() error {
			list, err = client.AccountList(ctx)
			return err
		},
		// This error can randomly occur.
		// Granted, the token is valid, we can retry `AccountList` safely, because the token is the only "parameter".
		// This might take up to a minute or more.
		retry.RetryIf(avngen.IsNotFound),
		retry.LastErrorOnly(true),
		retry.Attempts(10),
		retry.Delay(time.Second*10),
		retry.Context(ctx),
	)
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
