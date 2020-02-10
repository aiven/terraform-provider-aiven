package aiven

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceAccountTeamMember() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceAccountTeamMemberRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenAccountTeamMemberSchema, "account_id", "team_id", "user_email"),
	}
}

func datasourceAccountTeamMemberRead(d *schema.ResourceData, m interface{}) error {
	accountId := d.Get("account_id").(string)
	teamId := d.Get("team_id").(string)
	userEmail := d.Get("user_email").(string)

	d.SetId(buildResourceID(accountId, teamId, userEmail))

	return resourceAccountTeamMemberRead(d, m)
}
