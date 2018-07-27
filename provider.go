package main

import (
	"errors"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jelmersnoeck/aiven"
)

// Provider returns the Terraform Aiven Provider configuration object.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Aiven email address",
				Default:     "",
			},
			"otp": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Aiven One-Time password",
				Default:     "",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Aiven password",
				Default:     "",
			},
			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Aiven Authentication Token",
				Default:     "",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"aiven_project":      resourceProject(),
			"aiven_service":      resourceService(),
			"aiven_database":     resourceDatabase(),
			"aiven_service_user": resourceServiceUser(),
			"aiven_kafka_topic":  resourceKafkaTopic(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"aiven_project": dataSourceProject(),
		},

		ConfigureFunc: func(d *schema.ResourceData) (interface{}, error) {
			if d.Get("api_token") == "" && (d.Get("email") == "" || d.Get("password") == "") {
				return nil, errors.New("Must provide an API Token or email and password")
			}
			if d.Get("api_token") != "" {
				return aiven.NewTokenClient(
					d.Get("api_token").(string),
				)
			}
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

func optionalStringPointer(d *schema.ResourceData, key string) *string {
	str, ok := d.Get(key).(string)
	if !ok {
		return nil
	}
	return &str
}

func optionalIntPointer(d *schema.ResourceData, key string) *int {
	val, ok := d.Get(key).(int)
	if !ok {
		return nil
	}
	return &val
}
