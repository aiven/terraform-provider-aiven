package templates

import (
	"github.com/aiven/aiven-go-client/tools/exp/dist"
	"gopkg.in/yaml.v3"
)

func init() {
	var integrationSchema map[string]interface{}
	if err := yaml.Unmarshal(dist.IntegrationTypes, &integrationSchema); err != nil {
		panic("cannot unmarshal user configuration options integration JSON', error :" + err.Error())
	}
	userConfigSchemas["integration"] = integrationSchema
}
