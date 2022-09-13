package templates

import (
	"github.com/aiven/aiven-go-client/tools/exp/dist"
	"gopkg.in/yaml.v3"
)

func init() {
	var serviceSchema map[string]interface{}
	if err := yaml.Unmarshal(dist.ServiceTypes, &serviceSchema); err != nil {
		panic("cannot unmarshal user configuration options service JSON', error :" + err.Error())
	}
	userConfigSchemas["service"] = serviceSchema
}
