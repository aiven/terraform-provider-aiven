// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package service

import (
	"fmt"
	"strings"

	"github.com/aiven/terraform-provider-aiven/aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/uconf"
	"github.com/aiven/terraform-provider-aiven/aiven/templates"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// GenerateServiceUserConfigurationSchema generate service user_config
func GenerateServiceUserConfigurationSchema(t string) *schema.Schema {
	s := uconf.GenerateTerraformUserConfigSchema(
		templates.GetUserConfigSchema("service")[t].(map[string]interface{}))

	return &schema.Schema{
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      fmt.Sprintf("%s user configurable settings", strings.Title(t)),
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFuncSkipArrays(s),
		Elem:             &schema.Resource{Schema: s},
	}
}
