package v0

import (
	"context"
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/typeupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/dist"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func grafanaSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchema()
	s[schemautil.ServiceTypeGrafana] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Grafana server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	s[schemautil.ServiceTypeGrafana+"_user_config"] = dist.ServiceTypeGrafana()
	return s
}

func ResourceGrafanaResourceV0() *schema.Resource {
	return &schema.Resource{
		Description:   "The Grafana resource allows the creation and management of Aiven Grafana services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeGrafana),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeGrafana),
			schemautil.CustomizeDiffDisallowMultipleManyToOneKeys,
			customdiff.IfValueChange("tag",
				schemautil.TagsShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckUniqueTag,
			),
			customdiff.IfValueChange("disk_space",
				schemautil.DiskSpaceShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckDiskSpace,
			),
			customdiff.IfValueChange("additional_disk_space",
				schemautil.DiskSpaceShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckDiskSpace,
			),
			customdiff.IfValueChange("service_integrations",
				schemautil.ServiceIntegrationShouldNotBeEmpty,
				schemautil.CustomizeDiffServiceIntegrationAfterCreation,
			),
			customdiff.Sequence(
				schemautil.CustomizeDiffCheckPlanAndStaticIpsCannotBeModifiedTogether,
				schemautil.CustomizeDiffCheckStaticIpDisassociation,
			),
		),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: grafanaSchema(),
	}
}

func ResourceGrafanaStateUpgradeV0(
	_ context.Context,
	rawState map[string]interface{},
	_ interface{},
) (map[string]interface{}, error) {
	userConfigSlice, ok := rawState["grafana_user_config"].([]interface{})
	if !ok {
		return rawState, nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]interface{})
	if !ok {
		return rawState, nil
	}

	err := typeupgrader.Map(userConfig, map[string]string{
		"alerting_enabled":                 "bool",
		"alerting_max_annotations_to_keep": "int",
		"allow_embedding":                  "bool",
		"auth_basic_enabled":               "bool",
		"dashboard_previews_enabled":       "bool",
		"dashboards_versions_to_keep":      "int",
		"dataproxy_send_user_header":       "bool",
		"dataproxy_timeout":                "int",
		"disable_gravatar":                 "bool",
		"editors_can_admin":                "bool",
		"metrics_enabled":                  "bool",
		"static_ips":                       "bool",
		"user_auto_assign_org":             "bool",
		"viewers_can_edit":                 "bool",
	})
	if err != nil {
		return rawState, err
	}

	authAzureADSlice, ok := userConfig["auth_azuread"].([]interface{})
	if ok && len(authAzureADSlice) > 0 {
		authAzureAD, ok := authAzureADSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(authAzureAD, map[string]string{
				"allow_sign_up": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	authGenericOAuthSlice, ok := userConfig["auth_generic_oauth"].([]interface{})
	if ok && len(authGenericOAuthSlice) > 0 {
		authGenericOAuth, ok := authGenericOAuthSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(authGenericOAuth, map[string]string{
				"allow_sign_up": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	authGitHubSlice, ok := userConfig["auth_github"].([]interface{})
	if ok && len(authGitHubSlice) > 0 {
		authGitHub, ok := authGitHubSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(authGitHub, map[string]string{
				"allow_sign_up": "bool",
			})
			if err != nil {
				return rawState, err
			}

			authGitHubTeamIDs, ok := authGitHub["team_ids"].([]interface{})
			if ok {
				err = typeupgrader.Slice(authGitHubTeamIDs, "int")
				if err != nil {
					return rawState, err
				}
			}
		}
	}

	authGitLabSlice, ok := userConfig["auth_gitlab"].([]interface{})
	if ok && len(authGitLabSlice) > 0 {
		authGitLab, ok := authGitLabSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(authGitLab, map[string]string{
				"allow_sign_up": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	authGoogleSlice, ok := userConfig["auth_google"].([]interface{})
	if ok && len(authGoogleSlice) > 0 {
		authGoogle, ok := authGoogleSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(authGoogle, map[string]string{
				"allow_sign_up": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	privateAccessSlice, ok := userConfig["private_access"].([]interface{})
	if ok && len(privateAccessSlice) > 0 {
		privateAccess, ok := privateAccessSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(privateAccess, map[string]string{
				"grafana": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	privateLinkAccessSlice, ok := userConfig["privatelink_access"].([]interface{})
	if ok && len(privateLinkAccessSlice) > 0 {
		privateLinkAccess, ok := privateLinkAccessSlice[0].(map[string]interface{})
		if ok {
			err := typeupgrader.Map(privateLinkAccess, map[string]string{
				"grafana": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	publicAccessSlice, ok := userConfig["public_access"].([]interface{})
	if ok && len(publicAccessSlice) > 0 {
		publicAccess, ok := publicAccessSlice[0].(map[string]interface{})
		if ok {
			err := typeupgrader.Map(publicAccess, map[string]string{
				"grafana": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	smtpServerSlice, ok := userConfig["smtp_server"].([]interface{})
	if ok && len(smtpServerSlice) > 0 {
		smtpServer, ok := smtpServerSlice[0].(map[string]interface{})
		if ok {
			err := typeupgrader.Map(smtpServer, map[string]string{
				"port":        "int",
				"skip_verify": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	return rawState, nil
}
