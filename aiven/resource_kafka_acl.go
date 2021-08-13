// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/pkg/cache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var aivenKafkaACLSchema = map[string]*schema.Schema{
	"project": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Project to link the Kafka ACL to",
		ForceNew:    true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9_-]*$"),
			"project name should be alphanumeric"),
	},
	"service_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Service to link the Kafka ACL to",
		ForceNew:    true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9_-]*$"),
			"service name should be alphanumeric"),
	},
	"permission": {
		Type:         schema.TypeString,
		Required:     true,
		Description:  "Kafka permission to grant (admin, read, readwrite, write)",
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice([]string{"admin", "read", "readwrite", "write"}, false),
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
		ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(\*$|[a-zA-Z0-9-_?][a-zA-Z0-9-_?*]+)$`),
			"username should be alphanumeric"),
	},
}

func resourceKafkaACL() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKafkaACLCreate,
		ReadContext:   resourceKafkaACLRead,
		DeleteContext: resourceKafkaACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKafkaACLState,
		},

		Schema: aivenKafkaACLSchema,
	}
}

func resourceKafkaACLCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		return diag.FromErr(err)
	}

	d.SetId(buildResourceID(project, serviceName, acl.ID))

	return resourceKafkaACLRead(ctx, d, m)
}

func resourceKafkaACLRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, aclID := splitResourceID3(d.Id())
	acl, err := cache.ACLCache{}.Read(project, serviceName, aclID, client)
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
	}

	err = copyKafkaACLPropertiesFromAPIResponseToTerraform(d, &acl, project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaACLDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, aclID := splitResourceID3(d.Id())
	err := client.KafkaACLs.Delete(projectName, serviceName, aclID)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaACLState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<acl_id>", d.Id())
	}

	di := resourceKafkaACLRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get kafka acl: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}

func copyKafkaACLPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	acl *aiven.KafkaACL,
	project string,
	serviceName string,
) error {
	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("topic", acl.Topic); err != nil {
		return err
	}
	if err := d.Set("permission", acl.Permission); err != nil {
		return err
	}
	if err := d.Set("username", acl.Username); err != nil {
		return err
	}

	return nil
}
