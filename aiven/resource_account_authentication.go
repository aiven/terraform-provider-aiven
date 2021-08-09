package aiven

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var aivenAccountAuthenticationSchema = map[string]*schema.Schema{
	"account_id": {
		Type:        schema.TypeString,
		Description: "Account id",
		Required:    true,
	},
	"type": {
		Type:         schema.TypeString,
		Description:  "Account authentication id",
		Required:     true,
		ValidateFunc: validation.StringInSlice([]string{"internal", "saml"}, false),
	},
	"name": {
		Type:        schema.TypeString,
		Description: "Account team name",
		Required:    true,
	},
	"enabled": {
		Type:        schema.TypeBool,
		Description: "Status of account authentication method",
		Optional:    true,
		Default:     false,
	},
	"saml_certificate": {
		Type:        schema.TypeString,
		Description: "SAML Certificate",
		Optional:    true,
	},
	"saml_idp_url": {
		Type:        schema.TypeString,
		Description: "SAML Idp URL",
		Optional:    true,
	},
	"saml_entity_id": {
		Type:        schema.TypeString,
		Description: "SAML Entity id",
		Optional:    true,
	},
	"saml_acs_url": {
		Type:        schema.TypeString,
		Description: "SAML Assertion Consumer Service URL",
		Optional:    true,
		Computed:    true,
	},
	"saml_metadata_url": {
		Type:        schema.TypeString,
		Description: "SAML Metadata URL",
		Optional:    true,
		Computed:    true,
	},
	"authentication_id": {
		Type:        schema.TypeString,
		Description: "Account authentication id",
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

func resourceAccountAuthentication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAccountAuthenticationCreate,
		ReadContext:   resourceAccountAuthenticationRead,
		UpdateContext: resourceAccountAuthenticationUpdate,
		DeleteContext: resourceAccountAuthenticationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAccountAuthenticationState,
		},

		Schema: aivenAccountAuthenticationSchema,
	}
}

func resourceAccountAuthenticationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountId := d.Get("account_id").(string)

	r, err := client.AccountAuthentications.Create(
		accountId,
		aiven.AccountAuthenticationMethod{
			Enabled:         d.Get("enabled").(bool),
			Name:            d.Get("name").(string),
			Type:            d.Get("type").(string),
			SAMLCertificate: d.Get("saml_certificate").(string),
			SAMLIdpUrl:      d.Get("saml_idp_url").(string),
			SAMLEntity:      d.Get("saml_entity_id").(string),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildResourceID(
		r.AuthenticationMethod.AccountId,
		r.AuthenticationMethod.Id))

	return resourceAccountAuthenticationRead(ctx, d, m)
}

func resourceAccountAuthenticationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountId, authId := splitResourceID2(d.Id())
	r, err := client.AccountAuthentications.Get(accountId, authId)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("account_id", r.AuthenticationMethod.AccountId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", r.AuthenticationMethod.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", r.AuthenticationMethod.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled", r.AuthenticationMethod.Enabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_certificate", r.AuthenticationMethod.SAMLCertificate); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_idp_url", r.AuthenticationMethod.SAMLCertificate); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_entity_id", r.AuthenticationMethod.SAMLCertificate); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("authentication_id", r.AuthenticationMethod.Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_acs_url", r.AuthenticationMethod.SAMLAcsUrl); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_metadata_url", r.AuthenticationMethod.SAMLMetadataUrl); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("create_time", r.AuthenticationMethod.CreateTime.String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("update_time", r.AuthenticationMethod.UpdateTime.String()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAccountAuthenticationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	accountId, authId := splitResourceID2(d.Id())

	r, err := client.AccountAuthentications.Update(accountId, aiven.AccountAuthenticationMethod{
		Id:              authId,
		Enabled:         d.Get("enabled").(bool),
		Name:            d.Get("name").(string),
		Type:            d.Get("type").(string),
		SAMLCertificate: d.Get("saml_certificate").(string),
		SAMLIdpUrl:      d.Get("saml_idp_url").(string),
		SAMLEntity:      d.Get("saml_entity_id").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildResourceID(
		r.AuthenticationMethod.AccountId,
		r.AuthenticationMethod.Id))

	return resourceAccountAuthenticationRead(ctx, d, m)
}

func resourceAccountAuthenticationDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountId, teamId := splitResourceID2(d.Id())

	err := client.AccountAuthentications.Delete(accountId, teamId)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAccountAuthenticationState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	di := resourceAccountAuthenticationRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get account authentication %v", di)
	}

	return []*schema.ResourceData{d}, nil
}
