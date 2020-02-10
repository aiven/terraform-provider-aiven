package aiven

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceAccountTeamProject() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceAccountTeamProjectRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenAccountTeamProjectSchema, "account_id", "team_id", "project_name"),
	}
}

func datasourceAccountTeamProjectRead(d *schema.ResourceData, m interface{}) error {
	accountId := d.Get("account_id").(string)
	teamId := d.Get("team_id").(string)
	projectName := d.Get("project_name").(string)

	d.SetId(buildResourceID(accountId, teamId, projectName))

	return resourceAccountTeamProjectRead(d, m)
}
