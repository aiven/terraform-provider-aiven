package templates

import "log"

const (
	// User configuration options based on resource type

	UserConfigSchemaEndpoint    = "endpoint"
	UserConfigSchemaIntegration = "integration"
	UserConfigSchemaService     = "service"
)

// getUserConfigurationOptionsSchemaFilenames gets a list of user configuration
// options filenames based on resource type
func getUserConfigurationOptionsSchemaFilenames() map[string]struct{} {
	return map[string]struct{}{
		UserConfigSchemaEndpoint:    {},
		UserConfigSchemaIntegration: {},
		UserConfigSchemaService:     {},
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
