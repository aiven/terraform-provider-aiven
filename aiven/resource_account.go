package aiven

import (
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var aivenAccountSchema = map[string]*schema.Schema{
	"account_id": {
		Type:        schema.TypeString,
		Description: "Account id",
		Computed:    true,
	},
	"name": {
		Type:        schema.TypeString,
		Description: "Account name",
		Required:    true,
	},
	"owner_team_id": {
		Type:        schema.TypeString,
		Description: "Owner team id",
		Optional:    true,
		Computed:    true,
	},
	"tenant_id": {
		Type:        schema.TypeString,
		Description: "Tenant id",
		Optional:    true,
		Computed:    true,
	},
	"create_time": {
		Type:        schema.TypeString,
		Description: "Time of creation",
		Optional:    true,
		Computed:    true,
	},
	"update_time": {
		Type:        schema.TypeString,
		Description: "Time of last update",
		Optional:    true,
		Computed:    true,
	},
}

func resourceAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccountCreate,
		Read:   resourceAccountRead,
		Update: resourceAccountUpdate,
		Delete: resourceAccountDelete,
		Exists: resourceAccountExists,
		Importer: &schema.ResourceImporter{
			State: resourceAccountState,
		},

		Schema: aivenAccountSchema,
	}
}

func resourceAccountCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)
	name := d.Get("name").(string)

	r, err := client.Accounts.Create(
		aiven.Account{
			Name: name,
		},
	)
	if err != nil {
		return err
	}

	d.SetId(r.Account.Id)

	return resourceAccountRead(d, m)
}

func resourceAccountRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	r, err := client.Accounts.Get(d.Id())
	if err != nil {
		return err
	}

	if err := d.Set("account_id", r.Account.Id); err != nil {
		return err
	}
	if err := d.Set("name", r.Account.Name); err != nil {
		return err
	}
	if err := d.Set("owner_team_id", r.Account.OwnerTeamId); err != nil {
		return err
	}
	if err := d.Set("tenant_id", r.Account.TenantId); err != nil {
		return err
	}
	if err := d.Set("create_time", r.Account.CreateTime.String()); err != nil {
		return err
	}
	if err := d.Set("update_time", r.Account.UpdateTime.String()); err != nil {
		return err
	}

	return nil
}

func resourceAccountUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	r, err := client.Accounts.Update(d.Id(), aiven.Account{
		Name: d.Get("name").(string),
	})
	if err != nil {
		return err
	}

	d.SetId(r.Account.Id)

	return resourceAccountRead(d, m)
}

func resourceAccountDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	err := client.Accounts.Delete(d.Id())
	if err != nil {
		return err
	}

	return nil
}

func resourceAccountExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	_, err := client.Accounts.Get(d.Id())
	if err != nil {
		return false, err
	}

	return resourceExists(err)
}

func resourceAccountState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	err := resourceAccountRead(d, m)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
