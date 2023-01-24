package account

import (
	"context"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
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
		Description:  userconfig.Desc("The account authentication type.").PossibleValues("internal", "saml").Build(),
	},
	"enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: userconfig.Desc("Status of account authentication method.").DefaultValue(false).Build(),
	},
	"auto_join_team_id": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Team ID",
	},
	"saml_certificate": {
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "SAML Certificate",
		DiffSuppressFunc: schemautil.TrimSpaceDiffSuppressFunc,
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

	accountID := d.Get("account_id").(string)
	r, err := client.AccountAuthentications.Create(
		accountID,
		aiven.AccountAuthenticationMethodCreate{
			AuthenticationMethodName: d.Get("name").(string),
			AuthenticationMethodType: d.Get("type").(string),
			AutoJoinTeamID:           d.Get("auto_join_team_id").(string),
			SAMLCertificate:          strings.TrimSpace(d.Get("saml_certificate").(string)),
			SAMLDigestAlgorithm:      d.Get("saml_digest_algorithm").(string),
			SAMLEntityID:             d.Get("saml_entity_id").(string),
			SAMLFieldMapping:         readSAMLFieldMappingFromSchema(d),
			SAMLIdpLoginAllowed:      d.Get("saml_idp_login_allowed").(bool),
			SAMLIdpURL:               d.Get("saml_idp_url").(string),
			SAMLSignatureAlgorithm:   d.Get("saml_signature_algorithm").(string),
			SAMLVariant:              d.Get("saml_variant").(string),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(
		accountID,
		r.AuthenticationMethod.AuthenticationMethodID))

	return resourceAccountAuthenticationRead(ctx, d, m)
}

func resourceAccountAuthenticationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountID, authID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.AccountAuthentications.Get(accountID, authID)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("account_id", r.AuthenticationMethod.AccountID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", r.AuthenticationMethod.AuthenticationMethodName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", r.AuthenticationMethod.AuthenticationMethodType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled", r.AuthenticationMethod.AuthenticationMethodEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("auto_join_team_id", r.AuthenticationMethod.AutoJoinTeamID); err != nil {
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
	if err := d.Set("saml_idp_url", r.AuthenticationMethod.SAMLIdpURL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_signature_algorithm", r.AuthenticationMethod.SAMLSignatureAlgorithm); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_variant", r.AuthenticationMethod.SAMLVariant); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_entity_id", r.AuthenticationMethod.SAMLEntityID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("authentication_id", r.AuthenticationMethod.AuthenticationMethodID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_acs_url", r.AuthenticationMethod.SAMLAcsURL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("saml_metadata_url", r.AuthenticationMethod.SAMLMetadataURL); err != nil {
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
	accountID, authID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r := aiven.AccountAuthenticationMethodUpdate{
		AuthenticationMethodEnabled: d.Get("enabled").(bool),
		AuthenticationMethodName:    d.Get("name").(string),
		AutoJoinTeamID:              d.Get("auto_join_team_id").(string),
		SAMLCertificate:             strings.TrimSpace(d.Get("saml_certificate").(string)),
		SAMLDigestAlgorithm:         d.Get("saml_digest_algorithm").(string),
		SAMLFieldMapping:            readSAMLFieldMappingFromSchema(d),
		SAMLIdpLoginAllowed:         d.Get("saml_idp_login_allowed").(bool),
		SAMLIdpURL:                  d.Get("saml_idp_url").(string),
		SAMLSignatureAlgorithm:      d.Get("saml_signature_algorithm").(string),
		SAMLVariant:                 d.Get("saml_variant").(string),
		SAMLEntity:                  d.Get("saml_entity_id").(string),
	}

	_, err = client.AccountAuthentications.Update(accountID, authID, r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(accountID, authID))
	return resourceAccountAuthenticationRead(ctx, d, m)
}

func resourceAccountAuthenticationDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	accountID, teamID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.AccountAuthentications.Delete(accountID, teamID)
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
