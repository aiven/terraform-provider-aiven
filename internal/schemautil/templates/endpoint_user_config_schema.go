package templates

import (
	"github.com/aiven/aiven-go-client/tools/exp/dist"
	"gopkg.in/yaml.v3"
)

func init() {
	var endpointSchema map[string]interface{}
	if err := yaml.Unmarshal(dist.IntegrationEndpointTypes, &endpointSchema); err != nil {
		panic("cannot unmarshal user configuration options endpoint JSON', error :" + err.Error())
	}
	userConfigSchemas["endpoint"] = endpointSchema
}
