package provision

// MapServiceRequest transforms the frontend service request to template format
func MapServiceRequest(frontendReq map[string]any, projectName string) (map[string]any, string) {
	// Determine the resource type from service_type
	serviceType := frontendReq["service_type"].(string)
	resourceType := "aiven_" + serviceType

	// Get the service name to use as resource name
	serviceName := frontendReq["service_name"].(string)

	// Create base template data
	templateData := map[string]any{
		"resource_name": serviceName,
		"project":       projectName,
		"service_name":  serviceName,
	}

	// Copy plan which has the same name in both formats
	if plan, ok := frontendReq["plan"].(string); ok {
		templateData["plan"] = plan
	}

	// Map cloud to cloud_name
	if cloud, ok := frontendReq["cloud"].(string); ok && cloud != "" {
		templateData["cloud_name"] = cloud
	}

	// Map project_vpc_id if it exists and is not null
	if vpcID, ok := frontendReq["project_vpc_id"]; ok && vpcID != nil {
		templateData["project_vpc_id"] = vpcID
	}

	// Handle service-specific transformations
	switch serviceType {
	case "pg":
		mapPGConfig(frontendReq, templateData)
	case "kafka":
		mapKafkaConfig(frontendReq, templateData)
	}

	return templateData, resourceType
}

func mapPGConfig(frontendReq, templateData map[string]any) {
	// Transform user_config to pg_user_config format
	if userConfig, ok := frontendReq["user_config"].(map[string]any); ok && len(userConfig) > 0 {
		templateData["pg_user_config"] = []map[string]any{userConfig}
	}

	// Transform tags to tag format
	if tags, ok := frontendReq["tags"].(map[string]any); ok && len(tags) > 0 {
		tagList := make([]map[string]any, 0)
		for k, v := range tags {
			tagList = append(tagList, map[string]any{
				"key":   k,
				"value": v,
			})
		}
		templateData["tag"] = tagList
	}
}

func mapKafkaConfig(frontendReq, templateData map[string]any) {
	// Transform user_config to kafka_user_config format
	if userConfig, ok := frontendReq["user_config"].(map[string]any); ok && len(userConfig) > 0 {
		templateData["kafka_user_config"] = []map[string]any{userConfig}
	}

	// Transform tags to tag format (same as for PG)
	if tags, ok := frontendReq["tags"].(map[string]any); ok && len(tags) > 0 {
		tagList := make([]map[string]any, 0)
		for k, v := range tags {
			tagList = append(tagList, map[string]any{
				"key":   k,
				"value": v,
			})
		}
		templateData["tag"] = tagList
	}
}
