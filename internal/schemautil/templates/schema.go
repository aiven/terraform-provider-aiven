package templates

import "log"

const (
	// User configuration options based on resource type
	UserConfigSchemaEndpoint    = "endpoint"
	UserConfigSchemaIntegration = "integration"
	UserConfigSchemaService     = "common"

	// JSON file names for user configuration options
	EndpointFileName    = "integration_endpoints_user_config_schema.json"
	IntegrationFileName = "integrations_user_config_schema.json"
	ServiceFileName     = "service_user_config_schema.json"
)

// getUserConfigurationOptionsSchemaFilenames gets a list of user configuration
// options filenames based on resource type
func getUserConfigurationOptionsSchemaFilenames() map[string]string {
	return map[string]string{
		UserConfigSchemaEndpoint:    EndpointFileName,
		UserConfigSchemaIntegration: IntegrationFileName,
		UserConfigSchemaService:     ServiceFileName,
	}
}

// userConfigSchemas contains a list of generated user configuration options
var userConfigSchemas = make(map[string]map[string]interface{}, 3)

// GetUserConfigSchema get a user configuration options schema by resource type
func GetUserConfigSchema(t string) map[string]interface{} {
	if _, ok := getUserConfigurationOptionsSchemaFilenames()[t]; !ok {
		log.Panicf("user configuration options schema type `%s` is not available", t)
	}

	return userConfigSchemas[t]
}
