package billinggrouplist

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

func flattenModifier(_ context.Context, _ avngen.Client) util.MapModifier[tfModel] {
	return util.ComposeModifiers(
		// These emails are arrays of strings in Terraform.
		// But in the API they are arrays of objects with "email" key.
		flattenEmailsMap("billing_groups", "billing_contact_emails", "email"),
		flattenEmailsMap("billing_groups", "billing_emails", "email"),
	)
}

func flattenEmailsMap(arr, obj, key string) util.MapModifier[tfModel] {
	return func(r util.RawMap, _ *tfModel) error {
		data := r.GetData()

		var err error
		result := gjson.GetBytes(data, arr)
		result.ForEach(func(k, v gjson.Result) bool {
			// Get value by "key"
			value := v.Get(util.PathAny(obj, "#", key))

			// Deep replace the whole object with just the "value"
			data, err = sjson.SetRawBytes(data, util.PathAny(arr, k.Int(), obj), []byte(value.Raw))
			return err == nil // continue if no error
		})

		if err != nil {
			return err
		}

		return r.SetData(data)
	}
}
