// Code generated by user config generator. DO NOT EDIT.

package serviceintegrationendpoint

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/diff"
)

func autoscalerUserConfig() *schema.Schema {
	return &schema.Schema{
		Description:      "Autoscaler user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later",
		DiffSuppressFunc: diff.SuppressUnchanged,
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{"autoscaling": {
			Description: "Configure autoscaling thresholds for a service",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"cap_gb": {
					Description: "The maximum total disk size (in gb) to allow autoscaler to scale up to. Example: `300`.",
					Required:    true,
					Type:        schema.TypeInt,
				},
				"type": {
					Description:  "Enum: `autoscale_disk`. Type of autoscale event.",
					Required:     true,
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"autoscale_disk"}, false),
				},
			}},
			MaxItems: 64,
			Required: true,
			Type:     schema.TypeList,
		}}},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	}
}
