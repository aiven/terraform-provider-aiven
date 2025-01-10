package organization

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/applicationuser"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenOrganizationApplicationUserTokenSchema = map[string]*schema.Schema{
	"organization_id": {
		Type:        schema.TypeString,
		Description: "The ID of the organization the application user belongs to.",
		Required:    true,
		ForceNew:    true,
	},
	"user_id": {
		Type:        schema.TypeString,
		Description: "The ID of the application user the token is created for.",
		Required:    true,
		ForceNew:    true,
	},
	"description": {
		Type:        schema.TypeString,
		Description: "Description of the token.",
		Optional:    true,
		ForceNew:    true,
		Default:     "-",
	},
	"full_token": {
		Type:        schema.TypeString,
		Description: "Full token.",
		Computed:    true,
		Sensitive:   true,
	},
	"token_prefix": {
		Type:        schema.TypeString,
		Description: "Prefix of the token.",
		Computed:    true,
	},
	"max_age_seconds": {
		Type:        schema.TypeInt,
		Description: "The number of hours after which a token expires. Default session duration is 10 hours.",
		Optional:    true,
		ForceNew:    true,
	},
	"extend_when_used": {
		Type:        schema.TypeBool,
		Description: "Extends the token session duration when the token is used. Only applicable if a value is set for `max_age_seconds`.",
		Optional:    true,
		ForceNew:    true,
	},
	"scopes": {
		Type:     schema.TypeSet,
		ForceNew: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Description: "Limits access to specific resources by granting read or write privileges to them. For example: `billing:read`. " +
			"Available scopes are: `authentication`, `billing`, `payments` for [payment methods](https://aiven.io/docs/platform/howto/list-billing), " +
			"`privatelink`, `projects`, `services`, `static_ips`, and `user`.",
		Optional: true,
	},
	"currently_active": {
		Type:        schema.TypeBool,
		Description: "True if the API request was made with this token.",
		Computed:    true,
	},
	"create_time": {
		Type:        schema.TypeString,
		Description: "Time when the token was created.",
		Computed:    true,
	},
	"created_manually": {
		Type: schema.TypeBool,
		Description: "True for tokens explicitly created using the `access_tokens` API. False for tokens created " +
			"when a user logs in.",
		Computed: true,
	},
	"expiry_time": {
		Type:        schema.TypeString,
		Description: "Timestamp when the access token will expire unless extended.",
		Computed:    true,
	},
	"last_ip": {
		Type:        schema.TypeString,
		Description: "IP address of the last request made with this token.",
		Computed:    true,
	},
	"last_used_time": {
		Type:        schema.TypeString,
		Description: "Timestamp when the access token was last used.",
		Computed:    true,
	},
	"last_user_agent": {
		Type:        schema.TypeString,
		Description: "User agent of the last request made with this token.",
		Computed:    true,
	},
	"last_user_agent_human_readable": {
		Type:        schema.TypeString,
		Description: "User agent of the last request made with this token in human-readable format.",
		Computed:    true,
	},
}

func ResourceOrganizationApplicationUserToken() *schema.Resource {
	return &schema.Resource{
		Description: userconfig.Desc("Creates and manages an application user token. Review the" +
			" [best practices](https://aiven.io/docs/platform/concepts/application-users#security-best-practices) for securing application users and their tokens.").
			Build(),
		CreateContext: common.WithGenClient(resourceOrganizationApplicationUserTokenCreate),
		ReadContext:   common.WithGenClient(resourceOrganizationApplicationUserTokenRead),
		DeleteContext: common.WithGenClient(resourceOrganizationApplicationUserTokenDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   aivenOrganizationApplicationUserTokenSchema,
	}
}

func resourceOrganizationApplicationUserTokenCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var req applicationuser.ApplicationUserAccessTokenCreateIn
	err := schemautil.ResourceDataGet(d, &req)
	if err != nil {
		return err
	}

	orgID := d.Get("organization_id").(string)
	userID := d.Get("user_id").(string)

	token, err := client.ApplicationUserAccessTokenCreate(ctx, orgID, userID, &req)
	if err != nil {
		return err
	}

	err = schemautil.ResourceDataSet(d, token)
	if err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(orgID, userID, token.TokenPrefix))
	return resourceOrganizationApplicationUserTokenRead(ctx, d, client)
}

func resourceOrganizationApplicationUserTokenRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID, userID, tokenPrefix, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	tokens, err := client.ApplicationUserAccessTokensList(ctx, orgID, userID)
	if err != nil {
		return err
	}

	var token *applicationuser.TokenOut
	for i := range tokens {
		if tokenPrefix == tokens[i].TokenPrefix {
			token = &tokens[i]
			break
		}
	}

	if token == nil {
		return fmt.Errorf("application user token not found")
	}

	err = schemautil.ResourceDataSet(d, token)
	if err != nil {
		return err
	}

	return nil
}

func resourceOrganizationApplicationUserTokenDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID, userID, tokenPrefix, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	err = client.ApplicationUserAccessTokenDelete(ctx, orgID, userID, tokenPrefix)
	if err != nil {
		return fmt.Errorf("failed to delete application user token: %w", err)
	}
	return nil
}
