# Account Authentication Data Source

The Account Authentication data source provides information about the existing Aiven Account Authentication.

## Example Usage

```hcl
data "aiven_account_authentication" "foo" {
    account_id = aiven_account.<ACCOUNT_RESOURCE>.account_id
    name = "auth-1"
}
```

## Argument Reference

* `account_id` - (Required) is a unique account id.

* `name` - (Required) is an account authentication name.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `type` - is an account authentication type, can be one of `internal` and `saml`.

* `account_id` - is a unique account id.

* `name` - is an account authentication name.

* `type` - is an account authentication type, can be one of `internal` and `saml`.

* `enabled` - defines an authentication method enabled or not. 

* `saml_certificate` - is a SAML Certificate.

* `saml_entity_id` - is a SAML Entity ID.

* `saml_idp_url` - is a SAML Idp URL.

* `saml_acs_url` - is a SAML Assertion Consumer Service URL.

* `saml_metadata_url` - is a SAML Metadata URL.

* `authentication_id` - account authentication id.

* `create_time` - time of creation.

* `update_time` - time of last update.