package account

import (
	"context"
	"fmt"
	"strings"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/accountauthentication"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
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
		ValidateFunc: validation.StringInSlice(accountauthentication.AuthenticationMethodTypeChoices(), false),
		Description:  userconfig.Desc("The account authentication type.").PossibleValuesString(accountauthentication.AuthenticationMethodTypeChoices()...).Build(),
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
		Description:   "Creates and manages an authentication method.",
		CreateContext: common.WithGenClient(resourceAccountAuthenticationCreate),
		ReadContext:   common.WithGenClient(resourceAccountAuthenticationRead),
		UpdateContext: common.WithGenClient(resourceAccountAuthenticationUpdate),
		DeleteContext: common.WithGenClient(resourceAccountAuthenticationDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema:             aivenAccountAuthenticationSchema,
		DeprecationMessage: "This resource is deprecated. Use the Aiven Console instead. View the documentation for more information: https://aiven.io/docs/platform/howto/saml/add-identity-providers",
	}
}

func resourceAccountAuthenticationCreate(_ context.Context, _ *schema.ResourceData, _ avngen.Client) error {
	return fmt.Errorf("creating account authentication is unsupported")
}

func resourceAccountAuthenticationRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	accountID, authID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.AccountAuthenticationMethodGet(ctx, accountID, authID)
	if err != nil {
		return err
	}

	if err = d.Set("account_id", resp.AccountId); err != nil {
		return err
	}
	if err = d.Set("name", resp.AuthenticationMethodName); err != nil {
		return err
	}
	if err = d.Set("type", resp.AuthenticationMethodType); err != nil {
		return err
	}
	if err = d.Set("enabled", resp.AuthenticationMethodEnabled); err != nil {
		return err
	}
	if err = d.Set("auto_join_team_id", resp.AutoJoinTeamId); err != nil {
		return err
	}
	if err = d.Set("saml_certificate", resp.SamlCertificate); err != nil {
		return err
	}
	if err = d.Set("saml_digest_algorithm", resp.SamlDigestAlgorithm); err != nil {
		return err
	}
	if err = d.Set("saml_field_mapping", flattenSAMLFieldMapping(resp.SamlFieldMapping)); err != nil {
		return err
	}
	if err = d.Set("saml_idp_login_allowed", resp.SamlIdpLoginAllowed); err != nil {
		return err
	}
	if err = d.Set("saml_idp_url", resp.SamlIdpUrl); err != nil {
		return err
	}
	if err = d.Set("saml_signature_algorithm", resp.SamlSignatureAlgorithm); err != nil {
		return err
	}
	if err = d.Set("saml_variant", resp.SamlVariant); err != nil {
		return err
	}
	if err = d.Set("saml_entity_id", resp.SamlEntityId); err != nil {
		return err
	}
	if err = d.Set("authentication_id", resp.AuthenticationMethodId); err != nil {
		return err
	}
	if err = d.Set("saml_acs_url", resp.SamlAcsUrl); err != nil {
		return err
	}
	if err = d.Set("saml_metadata_url", resp.SamlMetadataUrl); err != nil {
		return err
	}
	if err = d.Set("create_time", resp.CreateTime.String()); err != nil {
		return err
	}
	if err = d.Set("update_time", resp.UpdateTime.String()); err != nil {
		return err
	}

	return nil
}

func flattenSAMLFieldMapping(s *accountauthentication.SamlFieldMappingOut) []map[string]interface{} {
	if s == nil {
		return make([]map[string]interface{}, 0)
	}

	v := make([]map[string]interface{}, 0)
	return append(v, map[string]interface{}{
		"email":      s.Email,
		"first_name": s.FirstName,
		"identity":   s.Identity,
		"last_name":  s.LastName,
		"real_name":  s.RealName,
	})
}

func resourceAccountAuthenticationUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	accountID, authID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	req := accountauthentication.AccountAuthenticationMethodUpdateIn{
		AuthenticationMethodName: util.NilIfZero(d.Get("name").(string)),
		AutoJoinTeamId:           util.NilIfZero(d.Get("auto_join_team_id").(string)),
		SamlCertificate:          util.NilIfZero(strings.TrimSpace(d.Get("saml_certificate").(string))),
		SamlDigestAlgorithm:      accountauthentication.SamlDigestAlgorithmType(d.Get("saml_digest_algorithm").(string)),
		SamlFieldMapping:         readSAMLFieldMappingFromSchema(d),
		SamlIdpUrl:               util.NilIfZero(d.Get("saml_idp_url").(string)),
		SamlSignatureAlgorithm:   accountauthentication.SamlSignatureAlgorithmType(d.Get("saml_signature_algorithm").(string)),
		SamlVariant:              accountauthentication.SamlVariantType(d.Get("saml_variant").(string)),
		SamlEntityId:             util.NilIfZero(d.Get("saml_entity_id").(string)),
	}

	// Handle booleans separately to distinguish between not set and false
	if enabled, ok := d.GetOk("enabled"); ok {
		req.AuthenticationMethodEnabled = util.ToPtr(enabled.(bool))
	}

	if allowed, ok := d.GetOk("saml_idp_login_allowed"); ok {
		req.SamlIdpLoginAllowed = util.ToPtr(allowed.(bool))
	}

	_, err = client.AccountAuthenticationMethodUpdate(ctx, accountID, authID, &req)
	if err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(accountID, authID))

	return resourceAccountAuthenticationRead(ctx, d, client)
}

func resourceAccountAuthenticationDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	accountID, teamID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	err = client.AccountAuthenticationMethodDelete(ctx, accountID, teamID)
	if common.IsCritical(err) {
		return err
	}

	return nil
}

func readSAMLFieldMappingFromSchema(d *schema.ResourceData) *accountauthentication.SamlFieldMappingIn {
	set := d.Get("saml_field_mapping").(*schema.Set).List()
	if len(set) == 0 {
		return nil
	}

	r := accountauthentication.SamlFieldMappingIn{}
	for _, v := range set {
		cv := v.(map[string]interface{})

		r.Email = util.NilIfZero(cv["email"].(string))
		r.FirstName = util.NilIfZero(cv["first_name"].(string))
		r.Identity = util.NilIfZero(cv["identity"].(string))
		r.LastName = util.NilIfZero(cv["last_name"].(string))
		r.RealName = util.NilIfZero(cv["real_name"].(string))
	}

	return &r
}
