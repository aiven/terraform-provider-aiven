package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jelmersnoeck/aiven"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "API Key to use communicating with Aiven. https://api.aiven.io/doc/#api-User-UserAuth",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"aiven_project":  resourceProject(),
			"aiven_service":  resourceService(),
			"aiven_database": resourceDatabase(),
		},

		ConfigureFunc: func(d *schema.ResourceData) (interface{}, error) {
			return aiven.NewTokenClient(d.Get("api_key").(string))
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
