package billinggroup

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

func expandModifier(_ context.Context, _ avngen.Client) util.MapModifier[tfModel] {
	return util.ComposeModifiers(
		util.ExpandArrayToObjects[tfModel](false, "email", "billing_contact_emails"),
		util.ExpandArrayToObjects[tfModel](false, "email", "billing_emails"),
	)
}

func flattenModifier(_ context.Context, _ avngen.Client) util.MapModifier[tfModel] {
	return util.ComposeModifiers(
		// These emails are arrays of strings in Terraform.
		// But in the API they are arrays of objects with "email" key.
		util.FlattenObjectsToArray[tfModel]("email", "billing_contact_emails"),
		util.FlattenObjectsToArray[tfModel]("email", "billing_emails"),
	)
}
