package byoc

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// RejectIdentityChange returns the modifyPlan of a provision resource. It fails the plan when the identity
// fields that hand the customer account over to Aiven (e.g. `aws_iam_role_arn`) differ from the
// environment's: replace the entity instead.
// Blackbox-tested: Provision and Update store them even when the environment isn't provisioned again,
// and replacing the provision resource deletes the environment.
// A planned value unknown until apply isn't checked, so it is stored.
func RejectIdentityChange(identityFields ...string) func(context.Context, avngen.Client, adapter.ResourceData) error {
	return func(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
		// Unknown until the entity is created or replaced. Not d.Get: it would return the state's old ID.
		envID, ok := d.GetConfigOk("custom_cloud_environment_id")
		if !ok {
			return nil
		}

		env, err := client.CustomCloudEnvironmentGet(ctx, d.Get("organization_id").(string), envID.(string))
		if err != nil {
			return err
		}
		var current map[string]any
		if err := schemautil.Remarshal(env, &current); err != nil {
			return err
		}

		for _, k := range identityFields {
			v, _ := current[k].(string)
			planned := d.Get(k).(string)
			if v != "" && planned != "" && v != planned {
				return fmt.Errorf("the custom cloud environment is provisioned with %s %q: to change it, replace the entity resource, which replaces this resource too", k, v)
			}
		}
		return nil
	}
}
