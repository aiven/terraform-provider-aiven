package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceAccountAuthentication() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceAccountAuthenticationRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenAccountAuthenticationSchema, "account_id", "name"),
	}
}

func datasourceAccountAuthenticationRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	name := d.Get("name").(string)
	accountId := d.Get("account_id").(string)

	r, err := client.AccountAuthentications.List(accountId)
	if err != nil {
		return err
	}

	for _, a := range r.AuthenticationMethods {
		if a.Name == name {
			d.SetId(buildResourceID(a.AccountId, a.Id))
			return resourceAccountAuthenticationRead(d, m)
		}
	}

	return fmt.Errorf("account authentication %s not found", name)
}
