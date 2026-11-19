package byoc_test

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

const (
	entityAddress    = "aiven_byoc_azure_entity.test"
	provisionAddress = "aiven_byoc_azure_provision.test"
	subscriptionID   = "12345678-9abc-4def-8123-ab3456cd9876"
	tenantID         = "23456789-abcd-4def-8123-ab3456cd9876"
	contactBlock     = `
  contact_emails {
    email = "ops@example.com"
    role  = "Admin"
  }
`
)

func azureEntity(extra string) string {
	return fmt.Sprintf(`
resource "aiven_byoc_azure_entity" "test" {
  organization_id  = "org1234567890"
  display_name     = "azure-test"
  cloud_region     = "westeurope"
  deployment_model = "standard"
  reserved_cidr    = "10.0.0.0/16"
  timeouts {
    delete = "3s"
  }
  %s
}
`, extra)
}

func azureProvision(subscription, tenant, extra string) string {
	return fmt.Sprintf(`
resource "aiven_byoc_azure_provision" "test" {
  organization_id             = aiven_byoc_azure_entity.test.organization_id
  custom_cloud_environment_id = aiven_byoc_azure_entity.test.custom_cloud_environment_id
  azure_subscription_id       = %q
  azure_tenant_id             = %q
  timeouts {
    create = "3s"
    delete = "3s"
    update = "3s"
  }
  %s
}
`, subscription, tenant, extra)
}

func entityID(t *testing.T, tf *azureTerraform) string {
	t.Helper()
	id, ok := tf.resource(entityAddress)["custom_cloud_environment_id"].(string)
	require.True(t, ok, "missing environment ID in entity state")
	require.NotEmpty(t, id)
	return id
}

func existingEnvironment(api *cloudAPI, provider, region, state string) string {
	return api.add(map[string]any{
		"cloud_provider": provider, "cloud_region": region, "state": state,
		"display_name": "azure-test", "deployment_model": "standard", "reserved_cidr": "10.0.0.0/16",
	})
}

func TestAccBYOCAzureLifecycle(t *testing.T) {
	acc.SkipIfNotAcc(t)
	t.Setenv("PROVIDER_AIVEN_ENABLE_BETA", "true")
	entity := azureEntity(contactBlock)
	stack := entity + azureProvision(subscriptionID, tenantID, "")

	t.Run("retry after late activation", func(t *testing.T) {
		api := &cloudAPI{provisionState: "creating"}
		tf := newAzureTerraform(t, api, stack)
		require.ErrorContains(t, tf.apply(), "context deadline exceeded")
		id := entityID(t, tf)
		require.Equal(t, "creating", api.get(id)["state"])

		// Aiven finishes after the first CLI invocation has exited.
		api.set(id, "state", "active")
		before := len(api.writes())
		err := tf.apply()
		assert.NotContains(t, api.writes()[before:], http.MethodDelete+" "+environmentsPath+"/"+id)
		require.NoError(t, err)
		require.Equal(t, id, entityID(t, tf))
		provision := tf.resource(provisionAddress)
		require.Equal(t, id, provision["custom_cloud_environment_id"])
		require.Equal(t, "active", provision["state"])
		tf.requireEmptyPlan()
	})

	t.Run("server default contact", func(t *testing.T) {
		api := &cloudAPI{}
		tf := newAzureTerraform(t, api, azureEntity(""))
		require.NoError(t, tf.apply())
		require.NotEmpty(t, api.get(entityID(t, tf))["contact_emails"])
		require.Empty(t, tf.resource(entityAddress)["contact_emails"])
		tf.requireEmptyPlan()
	})

	for _, state := range []string{"active", "validating", "creating"} {
		t.Run("destroy dependencies/"+state, func(t *testing.T) {
			revoked := filepath.Join(t.TempDir(), "role-revoked")
			api := &cloudAPI{provisionState: state, beforeDelete: func() {
				assert.NoFileExists(t, revoked, "the environment still needed the role when its DELETE was issued")
			}}
			// Record when Terraform actually destroys the dependency in its graph.
			role := fmt.Sprintf(`
resource "terraform_data" "role" {
  input = aiven_byoc_azure_entity.test.custom_cloud_environment_id
  provisioner "local-exec" {
    when        = destroy
    command     = "echo revoked > role-revoked"
    working_dir = %q
  }
}
`, filepath.ToSlash(filepath.Dir(revoked)))
			tf := newAzureTerraform(t, api, entity+role+azureProvision(subscriptionID, tenantID, "depends_on = [terraform_data.role]"))
			if state == "active" {
				require.NoError(t, tf.apply())
			} else {
				require.ErrorContains(t, tf.apply(), "context deadline exceeded")
			}
			id := entityID(t, tf)
			require.Equal(t, state, api.get(id)["state"])
			err := tf.destroy()
			if state != "active" && err != nil {
				// Keeping the dependency and surfacing the in-flight operation is also safe.
				require.Regexp(t, `(?i)validating|creating|in.progress|pending|deadline exceeded`, err.Error())
				require.NoFileExists(t, revoked)
				require.NotNil(t, tf.resource("terraform_data.role"))
				require.Equal(t, id, entityID(t, tf))
				return
			}
			require.NoError(t, err)
			require.FileExists(t, revoked)
			require.Nil(t, api.get(id))
		})
	}

	for _, deleted := range []string{"404", "deleted"} {
		t.Run("recreate missing entity/"+deleted, func(t *testing.T) {
			api := &cloudAPI{}
			if deleted == "deleted" {
				api.deleteState = deleted
			}
			tf := newAzureTerraform(t, api, stack)
			require.NoError(t, tf.apply())
			id := entityID(t, tf)
			tf.config(entity)
			require.NoError(t, tf.apply())
			require.Contains(t, api.writes(), http.MethodDelete+" "+environmentsPath+"/"+id)
			require.Equal(t, tfjson.Actions{tfjson.ActionCreate}, tf.plannedActions(entityAddress))
			require.NoError(t, tf.apply())
			require.NotEqual(t, id, entityID(t, tf))
		})
	}

	t.Run("imported entity deletion failure", func(t *testing.T) {
		api := &cloudAPI{deleteState: "deletion_failed"}
		id := existingEnvironment(api, "azure", "westeurope", "active")
		tf := newAzureTerraform(t, api, entity)
		require.NoError(t, tf.importResource(entityAddress, id))
		err := tf.destroy()
		require.Equal(t, "deletion_failed", api.get(id)["state"])
		assert.Equal(t, id, tf.resource(entityAddress)["custom_cloud_environment_id"])
		require.ErrorContains(t, err, "deletion_failed")
		api.allowDelete()
		require.NoError(t, tf.destroy())
		require.Nil(t, api.get(id))
	})

	t.Run("provision deletion failure", func(t *testing.T) {
		api := &cloudAPI{deleteState: "deletion_failed"}
		tf := newAzureTerraform(t, api, stack)
		require.NoError(t, tf.apply())
		id := entityID(t, tf)
		err := tf.destroy()
		require.ErrorContains(t, err, "deletion_failed")
		assert.NotContains(t, err.Error(), "context deadline exceeded", "deletion_failed must terminate polling")
		require.Equal(t, id, tf.resource(provisionAddress)["custom_cloud_environment_id"])
		api.allowDelete()
		require.NoError(t, tf.destroy())
		require.Nil(t, api.get(id))
	})

	t.Run("reject AWS entity import", func(t *testing.T) {
		api := &cloudAPI{}
		id := existingEnvironment(api, "aws", "eu-west-1", "active")
		tf := newAzureTerraform(t, api, entity)
		err := tf.importResource(entityAddress, id)
		if err == nil {
			t.Fatalf("import accepted AWS environment %s as Azure; next plan has actions %v", id, tf.plannedActions(entityAddress))
		}
		require.Regexp(t, `(?i)aws`, err.Error())
		require.Nil(t, tf.resource(entityAddress))
		require.Equal(t, "aws", api.get(id)["cloud_provider"])
	})

	t.Run("draft import must not bypass provisioning", func(t *testing.T) {
		api := &cloudAPI{}
		tf := newAzureTerraform(t, api, entity)
		require.NoError(t, tf.apply())
		id := entityID(t, tf)
		tf.config(stack)
		if err := tf.importResource(provisionAddress, id); err != nil {
			require.Regexp(t, `(?i)draft`, err.Error())
			require.Nil(t, tf.resource(provisionAddress))
			require.Equal(t, "draft", api.get(id)["state"])
			return
		}
		// If the provider accepts a draft import, apply must actually activate it.
		before := len(api.writes())
		err := tf.apply()
		assert.Contains(t, api.writes()[before:], http.MethodPost+" "+environmentsPath+"/"+id+"/provision")
		require.NoError(t, err)
		require.Equal(t, "active", api.get(id)["state"])
		require.Equal(t, id, tf.resource(provisionAddress)["custom_cloud_environment_id"])
	})

	tagged := azureEntity(contactBlock + `tags = { team = "payments" }`)
	t.Run("remove last configured tag", func(t *testing.T) {
		api := &cloudAPI{}
		tf := newAzureTerraform(t, api, tagged)
		require.NoError(t, tf.apply())
		id := entityID(t, tf)
		require.Equal(t, map[string]any{"team": "payments"}, api.get(id)["tags"])
		tf.config(entity)
		err := tf.apply()
		assert.Empty(t, api.get(id)["tags"])
		require.NoError(t, err)
		tf.requireEmptyPlan()
	})

	t.Run("restore last tag removed outside Terraform", func(t *testing.T) {
		api := &cloudAPI{}
		tf := newAzureTerraform(t, api, tagged)
		require.NoError(t, tf.apply())
		id := entityID(t, tf)
		api.set(id, "tags", map[string]any{})
		require.Equal(t, tfjson.Actions{tfjson.ActionUpdate}, tf.plannedActions(entityAddress))
		require.NoError(t, tf.apply())
		require.Equal(t, map[string]any{"team": "payments"}, api.get(id)["tags"])
	})

	t.Run("update subscription and tenant in place", func(t *testing.T) {
		api := &cloudAPI{}
		tf := newAzureTerraform(t, api, stack)
		require.NoError(t, tf.apply())
		id := entityID(t, tf)
		const newSubscription = "32345678-9abc-4def-8123-ab3456cd9876"
		const newTenant = "43456789-abcd-4def-8123-ab3456cd9876"
		for _, step := range []struct{ subscription, tenant string }{{newSubscription, tenantID}, {newSubscription, newTenant}} {
			tf.config(entity + azureProvision(step.subscription, step.tenant, ""))
			before := len(api.writes())
			require.Equal(t, tfjson.Actions{tfjson.ActionUpdate}, tf.plannedActions(provisionAddress))
			require.NoError(t, tf.apply())
			require.Equal(t, []string{http.MethodPut + " " + environmentsPath + "/" + id}, api.writes()[before:])
			attrs := tf.resource(provisionAddress)
			require.Equal(t, id, attrs["custom_cloud_environment_id"])
			for _, fields := range []map[string]any{attrs, api.get(id)} {
				require.Equal(t, step.subscription, fields["azure_subscription_id"])
				require.Equal(t, step.tenant, fields["azure_tenant_id"])
			}
		}
	})
}
