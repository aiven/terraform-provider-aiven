// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"github.com/aiven/terraform-provider-aiven/pkg/cache"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("AIVEN_TOKEN", nil),
				Description: "Aiven Authentication Token",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"aiven_connection_pool":              datasourceConnectionPool(),
			"aiven_database":                     datasourceDatabase(),
			"aiven_kafka_acl":                    datasourceKafkaACL(),
			"aiven_kafka_topic":                  datasourceKafkaTopic(),
			"aiven_kafka_connector":              datasourceKafkaConnector(),
			"aiven_kafka_schema":                 datasourceKafkaSchema(),
			"aiven_kafka_schema_configuration":   datasourceKafkaSchemaConfiguration(),
			"aiven_elasticsearch_acl":            datasourceElasticsearchACL(),
			"aiven_project":                      datasourceProject(),
			"aiven_project_user":                 datasourceProjectUser(),
			"aiven_project_vpc":                  datasourceProjectVPC(),
			"aiven_vpc_peering_connection":       datasourceVPCPeeringConnection(),
			"aiven_service":                      datasourceService(),
			"aiven_service_integration":          datasourceServiceIntegration(),
			"aiven_service_integration_endpoint": datasourceServiceIntegrationEndpoint(),
			"aiven_service_user":                 datasourceServiceUser(),
			"aiven_account":                      datasourceAccount(),
			"aiven_account_team":                 datasourceAccountTeam(),
			"aiven_account_team_project":         datasourceAccountTeamProject(),
			"aiven_account_team_member":          datasourceAccountTeamMember(),
			"aiven_mirrormaker_replication_flow": datasourceMirrorMakerReplicationFlowTopic(),
			"aiven_account_authentication":       datasourceAccountAuthentication(),
			"aiven_kafka":                        datasourceKafka(),
			"aiven_kafka_connect":                datasourceKafkaConnect(),
			"aiven_kafka_mirrormaker":            datasourceKafkaMirrormaker(),
			"aiven_pg":                           datasourcePG(),
			"aiven_mysql":                        datasourceMySQL(),
			"aiven_cassandra":                    datasourceCassandra(),
			"aiven_elasticsearch":                datasourceElasticsearch(),
			"aiven_grafana":                      datasourceGrafana(),
			"aiven_influxdb":                     datasourceInfluxDB(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"aiven_connection_pool":              resourceConnectionPool(),
			"aiven_database":                     resourceDatabase(),
			"aiven_kafka_acl":                    resourceKafkaACL(),
			"aiven_kafka_topic":                  resourceKafkaTopic(),
			"aiven_kafka_connector":              resourceKafkaConnector(),
			"aiven_kafka_schema":                 resourceKafkaSchema(),
			"aiven_kafka_schema_configuration":   resourceKafkaSchemaConfiguration(),
			"aiven_elasticsearch_acl":            resourceElasticsearchACL(),
			"aiven_project":                      resourceProject(),
			"aiven_project_user":                 resourceProjectUser(),
			"aiven_project_vpc":                  resourceProjectVPC(),
			"aiven_vpc_peering_connection":       resourceVPCPeeringConnection(),
			"aiven_service":                      resourceService(),
			"aiven_service_integration":          resourceServiceIntegration(),
			"aiven_service_integration_endpoint": resourceServiceIntegrationEndpoint(),
			"aiven_service_user":                 resourceServiceUser(),
			"aiven_account":                      resourceAccount(),
			"aiven_account_team":                 resourceAccountTeam(),
			"aiven_account_team_project":         resourceAccountTeamProject(),
			"aiven_account_team_member":          resourceAccountTeamMember(),
			"aiven_mirrormaker_replication_flow": resourceMirrorMakerReplicationFlow(),
			"aiven_account_authentication":       resourceAccountAuthentication(),
			"aiven_kafka":                        resourceKafka(),
			"aiven_kafka_connect":                resourceKafkaConnect(),
			"aiven_kafka_mirrormaker":            resourceKafkaMirrormaker(),
			"aiven_pg":                           resourcePG(),
			"aiven_mysql":                        resourceMySQL(),
			"aiven_cassandra":                    resourceCassandra(),
			"aiven_elasticsearch":                resourceElasticsearch(),
			"aiven_grafana":                      resourceGrafana(),
			"aiven_influxdb":                     resourceInfluxDB(),
		},

		ConfigureFunc: func(d *schema.ResourceData) (interface{}, error) {
			_ = cache.NewTopicCache()

			return aiven.NewTokenClient(d.Get("api_token").(string), "terraform-provider-aiven/")
		},
	}
}

func optionalString(d *schema.ResourceData, key string) string {
	str, ok := d.Get(key).(string)
	if !ok {
		return ""
	}
	return str
}

func optionalStringPointer(d *schema.ResourceData, key string) *string {
	val, ok := d.GetOk(key)
	if !ok {
		return nil
	}
	str, ok := val.(string)
	if !ok {
		return nil
	}
	return &str
}

func optionalIntPointer(d *schema.ResourceData, key string) *int {
	val, ok := d.GetOk(key)
	if !ok {
		return nil
	}
	intValue, ok := val.(int)
	if !ok {
		return nil
	}
	return &intValue
}

func buildResourceID(parts ...string) string {
	finalParts := make([]string, len(parts))
	for idx, part := range parts {
		finalParts[idx] = url.PathEscape(part)
	}
	return strings.Join(finalParts, "/")
}

func splitResourceID(resourceID string, n int) []string {
	parts := strings.SplitN(resourceID, "/", n)
	for idx, part := range parts {
		part, _ := url.PathUnescape(part)
		parts[idx] = part
	}
	return parts
}

func splitResourceID2(resourceID string) (string, string) {
	parts := splitResourceID(resourceID, 2)
	return parts[0], parts[1]
}

func splitResourceID3(resourceID string) (string, string, string) {
	parts := splitResourceID(resourceID, 3)
	return parts[0], parts[1], parts[2]
}

func splitResourceID4(resourceID string) (string, string, string, string) {
	parts := splitResourceID(resourceID, 4)
	return parts[0], parts[1], parts[2], parts[3]
}

func resourceExists(err error) (bool, error) {
	if err == nil {
		return true, nil
	}

	aivenError, ok := err.(aiven.Error)
	if !ok {
		return true, err
	}

	if aivenError.Status == 404 {
		return false, nil
	}
	if aivenError.Status < 200 || aivenError.Status >= 300 {
		return true, err
	}
	return true, nil
}

func createOnlyDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	return len(d.Id()) > 0
}

// emptyObjectDiffSuppressFunc suppresses a diff for service user configuration options when
// fields are not set by the user but have default or previously defined values.
func emptyObjectDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	// When a map inside a list contains only default values without explicit values set by
	// the user Terraform inteprets the map as not being present and the array length being
	// zero, resulting in bogus update that does nothing. Allow ignoring those.
	if old == "1" && new == "0" && strings.HasSuffix(k, ".#") {
		return true
	}

	// When a field is not set to any value and consequently is null (empty string) but had
	// a non-empty parameter before. Allow ignoring those.
	if new == "" && old != "" {
		return true
	}

	// There is a bug in Terraform 0.11 which interprets "true" as "0" and "false" as "1"
	if (new == "0" && old == "false") || (new == "1" && old == "true") {
		return true
	}

	return false
}

// Terraform does not allow default values for arrays but the IP filter user config value
// has default. We don't want to force users to always define explicit value just because
// of the Terraform restriction so suppress the change from default to empty (which would
// be nonsensical operation anyway)
func ipFilterArrayDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	return old == "1" && new == "0" && strings.HasSuffix(k, ".ip_filter.#")
}

func ipFilterValueDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	return old == "0.0.0.0/0" && new == "" && strings.HasSuffix(k, ".ip_filter.0")
}

// validateDurationString is a ValidateFunc that ensures a string parses
// as time.Duration format
func validateDurationString(v interface{}, k string) (ws []string, errors []error) {
	if _, err := time.ParseDuration(v.(string)); err != nil {
		log.Printf("[DEBUG] invalid duration: %s", err)
		errors = append(errors, fmt.Errorf("%q: invalid duration", k))
	}

	return
}

func flattenToString(a []interface{}) []string {
	r := make([]string, len(a))
	for i, v := range a {
		r[i] = fmt.Sprint(v)
	}

	return r
}
