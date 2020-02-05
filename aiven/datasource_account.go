package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceAccount() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceAccountRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenAccountSchema, "name"),
	}
}

func datasourceAccountRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	name := d.Get("name").(string)

	r, err := client.Accounts.List()
	if err != nil {
		return err
	}

	for _, ac := range r.Accounts {
		if ac.Name == name {
			d.SetId(ac.Id)
			return resourceAccountRead(d, m)
		}
	}

	return fmt.Errorf("account %s not found", name)
}
