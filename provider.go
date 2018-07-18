package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jelmersnoeck/aiven"
)

// Provider returns the Terraform Aiven Provider configuration object.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Aiven email address",
			},
			"otp": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Aiven One-Time password",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Aiven password",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"aiven_project":      resourceProject(),
			"aiven_service":      resourceService(),
			"aiven_database":     resourceDatabase(),
			"aiven_service_user": resourceServiceUser(),
		},

		ConfigureFunc: func(d *schema.ResourceData) (interface{}, error) {
			return aiven.NewMFAUserClient(
				d.Get("email").(string),
				d.Get("otp").(string),
				d.Get("password").(string),
			)
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
