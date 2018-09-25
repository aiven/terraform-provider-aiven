// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package main

import (
	"net/url"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

// Provider returns the Terraform Aiven Provider configuration object.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Aiven Authentication Token",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"aiven_connection_pool":              resourceConnectionPool(),
			"aiven_database":                     resourceDatabase(),
			"aiven_kafka_acl":                    resourceKafkaACL(),
			"aiven_kafka_topic":                  resourceKafkaTopic(),
			"aiven_project":                      resourceProject(),
			"aiven_project_user":                 resourceProjectUser(),
			"aiven_project_vpc":                  resourceProjectVPC(),
			"aiven_vpc_peering_connection":       resourceVPCPeeringConnection(),
			"aiven_service":                      resourceService(),
			"aiven_service_integration":          resourceServiceIntegration(),
			"aiven_service_integration_endpoint": resourceServiceIntegrationEndpoint(),
			"aiven_service_user":                 resourceServiceUser(),
		},

		ConfigureFunc: func(d *schema.ResourceData) (interface{}, error) {
			return aiven.NewTokenClient(d.Get("api_token").(string))
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
	if len(d.Id()) > 0 {
		return true
	}
	return false
}
