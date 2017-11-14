package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jelmersnoeck/aiven"
)

func resourceKafkaService() *schema.Resource {
	return &schema.Resource{
		Create: resourceKafkaServiceCreate,
		Read:   resourceKafkaServiceRead,
		Update: resourceKafkaServiceUpdate,
		Delete: resourceKafkaServiceDelete,

		Schema: map[string]*schema.Schema{
			"project": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Target cloud",
				ForceNew:    true,
			},
			"service_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service name",
				ForceNew:    true,
			},
			"cloud": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Target cloud",
			},
			"group_name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Service group name",
			},
			"plan": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Subscription plan",
			},
			"uri": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Service URI",
			},
			"hostname": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Service hostname",
			},
			"port": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Service port",
			},
			"state": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Service state",
			},
		},
	}
}

func resourceKafkaServiceCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	service, err := client.Services.Create(
		d.Get("project").(string),
		aiven.CreateServiceRequest{
			d.Get("cloud").(string),
			d.Get("group_name").(string),
			d.Get("plan").(string),
			d.Get("service_name").(string),
			"kafka",
		},
	)

	if err != nil {
		d.SetId("")
		return err
	}

	err = resourceServiceWait(d, m)

	if err != nil {
		d.SetId("")
		return err
	}

	d.SetId(service.Name + "!")

	return resourceServiceRead(d, m)
}

func resourceKafkaServiceRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	service, err := client.Services.Get(
		d.Get("project").(string),
		d.Get("service_name").(string),
	)
	if err != nil {
		return err
	}

	d.Set("name", service.Name)
	d.Set("state", service.State)
	d.Set("plan", service.Plan)

	hn, err := service.Hostname()
	if err != nil {
		return err
	}
	port, err := service.Port()
	if err != nil {
		return err
	}

	d.Set("hostname", hn)
	d.Set("port", port)

	return nil
}

func resourceKafkaServiceUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	_, err := client.Services.Update(
		d.Get("project").(string),
		d.Get("service_name").(string),
		aiven.UpdateServiceRequest{
			d.Get("cloud").(string),
			d.Get("group_name").(string),
			d.Get("plan").(string),
			true,
		},
	)
	if err != nil {
		return err
	}

	err = resourceServiceWait(d, m)

	if err != nil {
		return err
	}

	return resourceServiceRead(d, m)
}

func resourceKafkaServiceDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	return client.Services.Delete(
		d.Get("project").(string),
		d.Get("service_name").(string),
	)
}
