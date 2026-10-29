package serviceintegrationendpoint_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// These endpoints need no backing Aiven service or connection to Datadog.
const testDatadogAPIKey = "11111111222233334444555555555555"

func TestAccAivenServiceIntegrationEndpoint_datadog_empty_prefix(t *testing.T) {
	const resourceName = "aiven_service_integration_endpoint.test"
	path := tfjsonpath.New("datadog_user_config").AtSliceIndex(0).AtMapKey("extra_tags_prefix")

	t.Run("create_empty", func(t *testing.T) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenServiceIntegraitonEndpointResourceDestroy,
			Steps: []resource.TestStep{{
				Config: testAccDatadogEndpointConfig(acc.RandStr(), testAccDatadogUserConfig("")),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, path, knownvalue.StringExact("")),
				},
				Check: testAccCheckEndpointUserConfig(resourceName, map[string]string{
					"datadog_api_key": testDatadogAPIKey, "extra_tags_prefix": "",
				}),
				ConfigPlanChecks: endpointNoDiffChecks(),
			}},
		})
	})

	t.Run("update_to_empty", func(t *testing.T) {
		suffix := acc.RandStr()
		sameID := statecheck.CompareValue(compare.ValuesSame())
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckAivenServiceIntegraitonEndpointResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccDatadogEndpointConfig(suffix, testAccDatadogUserConfig("custom")),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, path, knownvalue.StringExact("custom")),
						sameID.AddStateValue(resourceName, tfjsonpath.New("id")),
					},
					Check: testAccCheckEndpointUserConfig(resourceName, map[string]string{
						"datadog_api_key": testDatadogAPIKey, "extra_tags_prefix": "custom",
					}),
				},
				{
					Config: testAccDatadogEndpointConfig(suffix, testAccDatadogUserConfig("")),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(resourceName, path, knownvalue.StringExact("")),
						sameID.AddStateValue(resourceName, tfjsonpath.New("id")),
					},
					Check: testAccCheckEndpointUserConfig(resourceName, map[string]string{
						"datadog_api_key": testDatadogAPIKey, "extra_tags_prefix": "",
					}),
					ConfigPlanChecks: endpointNoDiffChecks(plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)),
				},
			},
		})
	})
}

func TestAccAivenServiceIntegrationEndpoint_datadog_empty_prefix_backwardCompat(t *testing.T) {
	const resourceName = "aiven_service_integration_endpoint.test"
	sameID := statecheck.CompareValue(compare.ValuesSame())
	steps := acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
		TFConfig:           testAccDatadogEndpointConfig(acc.RandStr(), testAccDatadogUserConfig("")),
		OldProviderVersion: "4.65.0",
		Checks: testAccCheckEndpointUserConfig(resourceName, map[string]string{
			"datadog_api_key": testDatadogAPIKey, "extra_tags_prefix": "",
		}),
	})
	for i := range steps {
		steps[i].ConfigStateChecks = []statecheck.StateCheck{
			statecheck.ExpectKnownValue(resourceName,
				tfjsonpath.New("datadog_user_config").AtSliceIndex(0).AtMapKey("extra_tags_prefix"), knownvalue.StringExact("")),
			sameID.AddStateValue(resourceName, tfjsonpath.New("id")),
		}
	}
	steps[1].ConfigPlanChecks = endpointNoDiffChecks(
		endpointPlanDiagnostics{t: t, resourceName: resourceName},
		plancheck.ExpectEmptyPlan(),
	)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckAivenServiceIntegraitonEndpointResourceDestroy,
		Steps:        steps,
	})
}

func TestAccAivenServiceIntegrationEndpoint_remove_user_config(t *testing.T) {
	const resourceName = "aiven_service_integration_endpoint.test"
	for _, tc := range []struct{ name, version string }{
		{name: "current"},
		{name: "sdkv2_4_65_0", version: "4.65.0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			suffix := acc.RandStr()
			sameID := statecheck.CompareValue(compare.ValuesSame())
			expected := map[string]string{
				"datadog_api_key": testDatadogAPIKey, "extra_tags_prefix": "custom",
			}
			var resourceID string
			checkAPI := func(s *terraform.State) error {
				id, err := endpointResourceID(s, resourceName)
				if err != nil {
					return err
				}
				resourceID = id
				return checkEndpointAPIUserConfig(resourceID, expected)
			}
			testCase := resource.TestCase{
				PreCheck:     func() { acc.TestAccPreCheck(t) },
				CheckDestroy: testAccCheckAivenServiceIntegraitonEndpointResourceDestroy,
				ErrorCheck: func(testErr error) error {
					// This hook runs before cleanup, even when apply fails before the normal Check.
					if resourceID != "" {
						if err := checkEndpointAPIUserConfig(resourceID, expected); err != nil {
							t.Logf("API probe after test failure, before cleanup: %v", err)
						} else {
							t.Log("API probe after test failure, before cleanup: API key and prefix still match their configured values")
						}
					}
					return testErr
				},
				Steps: []resource.TestStep{
					{
						Config: testAccDatadogEndpointConfig(suffix, testAccDatadogUserConfig("custom")),
						ConfigStateChecks: []statecheck.StateCheck{
							sameID.AddStateValue(resourceName, tfjsonpath.New("id")),
						},
						Check: checkAPI,
					},
					{
						Config: testAccDatadogEndpointConfig(suffix, ""),
						ConfigStateChecks: []statecheck.StateCheck{
							sameID.AddStateValue(resourceName, tfjsonpath.New("id")),
						},
						Check:            checkAPI,
						ConfigPlanChecks: endpointNoDiffChecks(endpointPlanDiagnostics{t: t, resourceName: resourceName}),
					},
				},
			}
			if tc.version == "" {
				testCase.ProtoV6ProviderFactories = acc.TestProtoV6ProviderFactories
				testCase.Steps[1].ConfigStateChecks = append(testCase.Steps[1].ConfigStateChecks,
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("datadog_user_config"), knownvalue.ListExact([]knownvalue.Check{})),
				)
				testCase.Steps[1].ConfigPlanChecks.PreApply = append(testCase.Steps[1].ConfigPlanChecks.PreApply,
					plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
				)
			} else {
				// The SDKv2 baseline may retain the block in state or suppress the update.
				// Compare API preservation without imposing the Framework's block-removal behavior.
				testCase.ExternalProviders = map[string]resource.ExternalProvider{
					"aiven": acc.ExternalAivenProvider(t, tc.version),
				}
			}
			resource.ParallelTest(t, testCase)
		})
	}
}

func TestAccAivenServiceIntegrationEndpoint_remove_password(t *testing.T) {
	const resourceName = "aiven_service_integration_endpoint.test"
	suffix := acc.RandStr()
	sameID := statecheck.CompareValue(compare.ValuesSame())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegraitonEndpointResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaRegistryEndpointConfig(suffix, "username", `basic_auth_password = "password"`),
				ConfigStateChecks: []statecheck.StateCheck{
					sameID.AddStateValue(resourceName, tfjsonpath.New("id")),
				},
				Check: testAccCheckEndpointUserConfig(resourceName, map[string]string{
					"basic_auth_username": "username", "basic_auth_password": "password",
				}),
			},
			{
				// Changing the username forces a PUT even if removing the password alone is a no-op.
				Config: testAccSchemaRegistryEndpointConfig(suffix, "new-username", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					sameID.AddStateValue(resourceName, tfjsonpath.New("id")),
					statecheck.ExpectKnownValue(resourceName,
						tfjsonpath.New("external_schema_registry_user_config").AtSliceIndex(0).AtMapKey("basic_auth_password"),
						knownvalue.StringExact("password")),
				},
				Check: testAccCheckEndpointUserConfig(resourceName, map[string]string{
					"basic_auth_username": "new-username", "basic_auth_password": "password",
				}),
				ConfigPlanChecks: endpointNoDiffChecks(plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)),
			},
		},
	})
}

func endpointNoDiffChecks(preApply ...plancheck.PlanCheck) resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{
		PreApply:             preApply,
		PostApplyPreRefresh:  []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
		PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
	}
}

func testAccDatadogEndpointConfig(suffix, userConfig string) string {
	return fmt.Sprintf(`
resource "aiven_service_integration_endpoint" "test" {
  project       = %q
  endpoint_name = "test-acc-ie-%s"
  endpoint_type = "datadog"
  %s
}
`, acc.ProjectName(), suffix, userConfig)
}

func testAccDatadogUserConfig(prefix string) string {
	return fmt.Sprintf(`
datadog_user_config {
  datadog_api_key    = %q
  extra_tags_prefix = %q
}
`, testDatadogAPIKey, prefix)
}

func testAccSchemaRegistryEndpointConfig(suffix, username, passwordConfig string) string {
	return fmt.Sprintf(`
resource "aiven_service_integration_endpoint" "test" {
  project       = %q
  endpoint_name = "test-acc-ie-%s"
  endpoint_type = "external_schema_registry"
  external_schema_registry_user_config {
    url                 = "https://schema-registry.example.com:8081"
    authentication      = "basic"
    basic_auth_username = %q
    %s
  }
}
`, acc.ProjectName(), suffix, username, passwordConfig)
}
