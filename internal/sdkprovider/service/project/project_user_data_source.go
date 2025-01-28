package project

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceProjectUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceProjectUserRead),
		Description: "The Project User data source provides information about the existing Aiven Project User.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenProjectUserSchema,
			"project", "email"),
	}
}

func datasourceProjectUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		projectName = d.Get("project").(string)
		email       = d.Get("email").(string)
	)

	pul, err := client.ProjectUserList(ctx, projectName)
	if err != nil {
		return err
	}

	for _, user := range pul.Users {
		if user.UserEmail == email {
			d.SetId(schemautil.BuildResourceID(projectName, email))

			return resourceProjectUserRead(ctx, d, client)
		}
	}

	for _, invitation := range pul.Invitations {
		if invitation.InvitedUserEmail == email {
			d.SetId(schemautil.BuildResourceID(projectName, email))

			return resourceProjectUserRead(ctx, d, client)
		}
	}

	return fmt.Errorf("project user %s/%s not found", projectName, email)
}
