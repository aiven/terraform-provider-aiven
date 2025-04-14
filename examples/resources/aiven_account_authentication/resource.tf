resource "aiven_account_authentication" "foo" {
  account_id       = aiven_account.ACCOUNT_RESOURCE.account_id
  name             = "auth-1"
  type             = "saml"
  enabled          = true
  saml_certificate = "---CERTIFICATE---"
  saml_entity_id   = "https://example.com/00000"
  saml_idp_url     = "https://example.com/sso/saml"
}
