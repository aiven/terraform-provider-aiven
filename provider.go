package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jelmersnoeck/aiven"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"email": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Aiven email address",
			},
			"otp": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Aiven One-Time password",
			},
			"password": &schema.Schema{
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
			"aiven_kafka_topic":  resourceKafkaTopic(),
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
