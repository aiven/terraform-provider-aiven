// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package main

import (
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceKafkaACL() *schema.Resource {
	return &schema.Resource{
		Create: resourceKafkaACLCreate,
		Read:   resourceKafkaACLRead,
		Delete: resourceKafkaACLDelete,
		Exists: resourceKafkaACLExists,
		Importer: &schema.ResourceImporter{
			State: resourceKafkaACLState,
		},

		Schema: map[string]*schema.Schema{
			"project": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Project to link the Kafka ACL to",
				ForceNew:    true,
			},
			"service_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service to link the Kafka ACL to",
				ForceNew:    true,
			},
			"permission": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Kafka permission to grant (admin, read, readwrite, write)",
				ForceNew:    true,
			},
			"topic": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Topic name pattern for the ACL entry",
				ForceNew:    true,
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username pattern for the ACL entry",
				ForceNew:    true,
			},
		},
	}
}

func resourceKafkaACLCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	acl, err := client.KafkaACLs.Create(
		project,
		serviceName,
		aiven.CreateKafkaACLRequest{
			Permission: d.Get("permission").(string),
			Topic:      d.Get("topic").(string),
			Username:   d.Get("username").(string),
		},
	)
	if err != nil {
		return err
	}

	d.SetId(buildResourceID(project, serviceName, acl.ID))
	return copyKafkaACLPropertiesFromAPIResponseToTerraform(d, acl, project, serviceName)
}

func resourceKafkaACLRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project, serviceName, aclID := splitResourceID3(d.Id())
	acl, err := client.KafkaACLs.Get(project, serviceName, aclID)
	if err != nil {
		return err
	}

	return copyKafkaACLPropertiesFromAPIResponseToTerraform(d, acl, project, serviceName)
}

func resourceKafkaACLDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName, aclID := splitResourceID3(d.Id())
	return client.KafkaACLs.Delete(projectName, serviceName, aclID)
}

func resourceKafkaACLExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	projectName, serviceName, aclID := splitResourceID3(d.Id())
	_, err := client.KafkaACLs.Get(projectName, serviceName, aclID)
	return resourceExists(err)
}

func resourceKafkaACLState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<acl_id>", d.Id())
	}

	err := resourceKafkaACLRead(d, m)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func copyKafkaACLPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	acl *aiven.KafkaACL,
	project string,
	serviceName string,
) error {
	d.Set("project", project)
	d.Set("service_name", serviceName)
	d.Set("topic", acl.Topic)
	d.Set("permission", acl.Permission)
	d.Set("username", acl.Username)

	return nil
}
