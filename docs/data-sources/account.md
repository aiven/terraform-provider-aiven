# Account Data Source

The Account data source provides information about the existing Aiven Account.

## Example Usage

```hcl
data "aiven_account" "account1" {
    name = "<ACCOUNT_NAME>"
}
```

## Argument Reference

* `name` - (Required) defines an account name.

## Attribute Reference

The following attributes are exported:

* `account_id` - is an auto-generated unique account id.

* `owner_team_id` - is an owner team id.

* `tenant_id` - is a tenant id.

* `create_time` - time of creation.

* `update_time` - time of last update.