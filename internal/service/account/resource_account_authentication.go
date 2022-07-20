package account

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var aivenAccountAuthenticationSchema = map[string]*schema.Schema{
	"account_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The unique id of the account.",
	},
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The name of the account authentication.",
	},
	"type": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice([]string{"internal", "saml"}, false),
		Description:  schemautil.Complex("The account authentication type.").PossibleValues("internal", "saml").Build(),
	},
	"enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: schemautil.Complex("Status of account authentication method.").DefaultValue(false).Build(),
	},
	"auto_join_team_id": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Team ID",
	},
	"saml_certificate": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "SAML Certificate",
	},
	"saml_digest_algorithm": {
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "sha256",
		Description: "Digest algorithm. This is an advanced option that typically does not need to be set.",
	},
	"saml_field_mapping": {
		Type:        schema.TypeSet,
		MaxItems:    1,
		Optional:    true,
		Description: "Map IdP fields",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"email": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Field name for user email",
				},
				"first_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Field name for user's first name",
				},
				"identity": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Field name for user's identity. This field must always exist in responses, and must be immutable and unique. Contents of this field are used to identify the user. Using user ID (such as unix user id) is highly recommended, as email address may change, requiring relinking user to Aiven user.",
				},
				"last_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Field name for user's last name",
				},
				"real_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Field name for user's full name. If specified, first_name and last_name mappings are ignored",
				},
			},
		},
	},
	"saml_idp_login_allowed": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Set to 'true' to enable IdP initiated login",
	},
	"saml_idp_url": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "SAML Idp URL",
	},
	"saml_signature_algorithm": {
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "rsa-sha256",
		Description: "Signature algorithm. This is an advanced option that typically does not need to be set.",
	},
	"saml_variant": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "SAML server variant",
	},
	"saml_entity_id": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "SAML Entity id",
	},
	"authentication_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Account authentication id",
	},
	"saml_acs_url": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "SAML Assertion Consumer Service URL",
	},
	"saml_metadata_url": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "SAML Metadata URL",
	},
	"create_time": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Time of creation",
	},
	"update_time": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Time of last update",
	},
}

func ResourceAccountAuthentication() *schema.Resource {
	return &schema.Resource{
		Description:   "The Account Authentication resource allows the creation and management of an Aiven Account Authentications.",
		CreateContext: resourceAccountAuthenticationCreate,
		ReadContext:   resourceAccountAuthenticationRead,
		UpdateContext: resourceAccountAuthenticationUpdate,
		DeleteContext: resourceAccountAuthenticationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			Enabled:                d.Get("enabled").(bool),
			Name:                   d.Get("name").(string),
			Type:                   d.Get("type").(string),
			AutoJoinTeamId:         d.Get("auto_join_team_id").(string),
			SAMLCertificate:        d.Get("saml_certificate").(string),
			SAMLDigestAlgorithm:    d.Get("saml_digest_algorithm").(string),
			SAMLFieldMapping:       readSAMLFieldMappingFromSchema(d),
			SAMLIdpLoginAllowed:    d.Get("saml_idp_login_allowed").(bool),
			SAMLIdpUrl:             d.Get("saml_idp_url").(string),
			SAMLSignatureAlgorithm: d.Get("saml_signature_algorithm").(string),
			SAMLVariant:            d.Get("saml_variant").(string),
			SAMLEntity:             d.Get("saml_entity_id").(string),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(
		r.AuthenticationMethod.AccountId,
		r.AuthenticationMethod.Id))

	return resourceAccountAuthenticationRead(ctx, d, m)
}

func resourceAccountAuthenticationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountId, authId, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.AccountAuthentications.Get(accountId, authId)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
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
	if err := d.Set("auto_join_team_id", r.AuthenticationMethod.AutoJoinTeamId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_certificate", r.AuthenticationMethod.SAMLCertificate); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_digest_algorithm", r.AuthenticationMethod.SAMLDigestAlgorithm); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_field_mapping", r.AuthenticationMethod.SAMLFieldMapping); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_idp_login_allowed", r.AuthenticationMethod.SAMLIdpLoginAllowed); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_idp_url", r.AuthenticationMethod.SAMLIdpUrl); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_signature_algorithm", r.AuthenticationMethod.SAMLSignatureAlgorithm); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_variant", r.AuthenticationMethod.SAMLVariant); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_entity_id", r.AuthenticationMethod.SAMLEntity); err != nil {
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
	accountId, authId, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.AccountAuthentications.Update(accountId, aiven.AccountAuthenticationMethod{
		Id:                     authId,
		Enabled:                d.Get("enabled").(bool),
		Name:                   d.Get("name").(string),
		Type:                   d.Get("type").(string),
		AutoJoinTeamId:         d.Get("auto_join_team_id").(string),
		SAMLCertificate:        d.Get("saml_certificate").(string),
		SAMLDigestAlgorithm:    d.Get("saml_digest_algorithm").(string),
		SAMLFieldMapping:       readSAMLFieldMappingFromSchema(d),
		SAMLIdpLoginAllowed:    d.Get("saml_idp_login_allowed").(bool),
		SAMLIdpUrl:             d.Get("saml_idp_url").(string),
		SAMLSignatureAlgorithm: d.Get("saml_signature_algorithm").(string),
		SAMLVariant:            d.Get("saml_variant").(string),
		SAMLEntity:             d.Get("saml_entity_id").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(
		r.AuthenticationMethod.AccountId,
		r.AuthenticationMethod.Id))

	return resourceAccountAuthenticationRead(ctx, d, m)
}

func resourceAccountAuthenticationDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountId, teamId, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.AccountAuthentications.Delete(accountId, teamId)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func readSAMLFieldMappingFromSchema(d *schema.ResourceData) *aiven.SAMLFieldMapping {
	set := d.Get("saml_field_mapping").(*schema.Set).List()
	if len(set) == 0 {
		return nil
	}

	r := aiven.SAMLFieldMapping{}
	for _, v := range set {
		cv := v.(map[string]interface{})

		r.Email = cv["email"].(string)
		r.FirstName = cv["first_name"].(string)
		r.Identity = cv["identity"].(string)
		r.LastName = cv["last_name"].(string)
		r.RealName = cv["real_name"].(string)
	}

	return &r
}
