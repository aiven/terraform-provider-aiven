package billinggroup

import "github.com/aiven/terraform-provider-aiven/internal/plugin/util"

func expandModifier(r util.RawMap, plan *tfModel) error {
	return util.ComposeModifiers(
		util.ExpandArrayToObjects[tfModel](false, "email", "billing_contact_emails"),
		util.ExpandArrayToObjects[tfModel](false, "email", "billing_emails"),
	)(r, plan)
}

func flattenModifier(r util.RawMap, plan *tfModel) error {
	return util.ComposeModifiers(
		// These emails are arrays of strings in Terraform.
		// But in the API they are arrays of objects with "email" key.
		util.FlattenObjectsToArray[tfModel]("email", "billing_contact_emails"),
		util.FlattenObjectsToArray[tfModel]("email", "billing_emails"),
	)(r, plan)
}
