# Account Resource

The Account resource allows the creation and management of an Aiven Account.

## Example Usage

```hcl
resource "aiven_account" "account1" {
    name = "<ACCOUNT_NAME>"
}
```

## Argument Reference

* `name` - (Required) defines an account name.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `account_id` - is an auto-generated unique account id.

* `owner_team_id` - is an owner team id.

* `tenant_id` - is a tenant id.

* `create_time` - time of creation.

* `update_time` - time of last update.

Aiven ID format when importing existing resource: `<account_id>`
