package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceAccountTeam() *schema.Resource {
	return &schema.Resource{
		Read: datasourceAccountTeamRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenAccountTeamSchema,
			"account_id", "name"),
	}
}

func datasourceAccountTeamRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	name := d.Get("name").(string)
	accountId := d.Get("account_id").(string)

	r, err := client.AccountTeams.List(accountId)
	if err != nil {
		return err
	}

	for _, t := range r.Teams {
		if t.Name == name {
			d.SetId(buildResourceID(t.AccountId, t.Id))
			return resourceAccountTeamRead(d, m)
		}
	}

	return fmt.Errorf("account team %s not found", name)
}
