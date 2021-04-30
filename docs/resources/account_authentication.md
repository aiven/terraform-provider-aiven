# Account Authentication Resource

The Account Authentication resource allows the creation and management of an Aiven Account Authentications.

## Example Usage

```hcl
resource "aiven_account_authentication" "foo" {
    account_id = aiven_account.<ACCOUNT_RESOURCE>.account_id
    name = "auth-1"
    type = "saml"
    enabled = true
    saml_certificate = "---CERTIFICATE---"
    saml_entity_id = "https://example.com/00000"
    saml_idp_url = "https://example.com/sso/saml"
}
```

## Argument Reference

* `account_id` - (Required) is a unique account id.

* `name` - (Required) is an account authentication name.

* `type` - (Required) is an account authentication type, can be one of `internal` and `saml`.

* `account_id` - (Optional) is a unique account id.

* `name` - (Optional) is an account authentication name.

* `type` - (Optional) is an account authentication type, can be one of `internal` and `saml`.

* `enabled` - (Optional) defines an authentication method enabled or not. 

* `saml_certificate` - (Optional) is a SAML Certificate.

* `saml_entity_id` - (Optional) is a SAML Entity ID.

* `saml_idp_url` - (Optional) is a SAML Idp URL.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `saml_acs_url` - is a SAML Assertion Consumer Service URL.

* `saml_metadata_url` - is a SAML Metadata URL.

* `authentication_id` - account authentication id.

* `create_time` - time of creation.

* `update_time` - time of last update.

Aiven ID format when importing existing resource: `<account_id>/<auth_id>`
