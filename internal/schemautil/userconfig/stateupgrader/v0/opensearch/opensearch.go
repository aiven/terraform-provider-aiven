package opensearch

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/typeupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/dist"
)

func opensearchSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchema()
	s[schemautil.ServiceTypeOpenSearch] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "OpenSearch server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"opensearch_dashboards_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "URI for OpenSearch dashboard frontend",
					Sensitive:   true,
				},
			},
		},
	}
	s[schemautil.ServiceTypeOpenSearch+"_user_config"] = dist.ServiceTypeOpenSearch()

	return s
}

func ResourceOpenSearch() *schema.Resource {
	return &schema.Resource{
		Description:   "The OpenSearch resource allows the creation and management of Aiven OpenSearch services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeOpenSearch),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeOpenSearch),
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
				schemautil.CustomizeDiffCheckStaticIPDisassociation,
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

		Schema: opensearchSchema(),
	}
}

func ResourceOpenSearchStateUpgrade(
	_ context.Context,
	rawState map[string]interface{},
	_ interface{},
) (map[string]interface{}, error) {
	userConfigSlice, ok := rawState["opensearch_user_config"].([]interface{})
	if !ok {
		return rawState, nil
	}

	if len(userConfigSlice) == 0 {
		return rawState, nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]interface{})
	if !ok {
		return rawState, nil
	}

	err := typeupgrader.Map(userConfig, map[string]string{
		"disable_replication_factor_adjustment": "bool",
		"keep_index_refresh_interval":           "bool",
		"max_index_count":                       "int",
		"static_ips":                            "bool",
	})
	if err != nil {
		return rawState, err
	}

	indexPatternsSlice, ok := userConfig["index_patterns"].([]interface{})
	if ok && len(indexPatternsSlice) > 0 {
		for _, v := range indexPatternsSlice {
			indexPattern, ok := v.(map[string]interface{})
			if !ok {
				continue
			}

			err = typeupgrader.Map(indexPattern, map[string]string{
				"max_index_count": "int",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	indexTemplateSlice, ok := userConfig["index_template"].([]interface{})
	if ok && len(indexTemplateSlice) > 0 {
		indexTemplate, ok := indexTemplateSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(indexTemplate, map[string]string{
				"mapping_nested_objects_limit": "int",
				"number_of_replicas":           "int",
				"number_of_shards":             "int",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	opensearchSlice, ok := userConfig["opensearch"].([]interface{})
	if ok && len(opensearchSlice) > 0 {
		opensearch, ok := opensearchSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(opensearch, map[string]string{
				"action_auto_create_index_enabled":                      "bool",
				"action_destructive_requires_name":                      "bool",
				"cluster_max_shards_per_node":                           "int",
				"cluster_routing_allocation_node_concurrent_recoveries": "int",
				"http_max_content_length":                               "int",
				"http_max_header_size":                                  "int",
				"http_max_initial_line_length":                          "int",
				"indices_fielddata_cache_size":                          "int",
				"indices_memory_index_buffer_size":                      "int",
				"indices_queries_cache_size":                            "int",
				"indices_query_bool_max_clause_count":                   "int",
				"indices_recovery_max_bytes_per_sec":                    "int",
				"indices_recovery_max_concurrent_file_chunks":           "int",
				"override_main_response_version":                        "bool",
				"search_max_buckets":                                    "int",
				"thread_pool_analyze_queue_size":                        "int",
				"thread_pool_analyze_size":                              "int",
				"thread_pool_force_merge_size":                          "int",
				"thread_pool_get_queue_size":                            "int",
				"thread_pool_get_size":                                  "int",
				"thread_pool_search_queue_size":                         "int",
				"thread_pool_search_size":                               "int",
				"thread_pool_search_throttled_queue_size":               "int",
				"thread_pool_search_throttled_size":                     "int",
				"thread_pool_write_queue_size":                          "int",
				"thread_pool_write_size":                                "int",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	opensearchDashboardsSlice, ok := userConfig["opensearch_dashboards"].([]interface{})
	if ok && len(opensearchDashboardsSlice) > 0 {
		opensearchDashboards, ok := opensearchDashboardsSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(opensearchDashboards, map[string]string{
				"enabled":                    "bool",
				"max_old_space_size":         "int",
				"opensearch_request_timeout": "int",
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
				"opensearch":            "bool",
				"opensearch_dashboards": "bool",
				"prometheus":            "bool",
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
				"opensearch":            "bool",
				"opensearch_dashboards": "bool",
				"prometheus":            "bool",
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
				"opensearch":            "bool",
				"opensearch_dashboards": "bool",
				"prometheus":            "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	return rawState, nil
}
