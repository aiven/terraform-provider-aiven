# Account Team Data Source

The Account Team data source provides information about the existing Account Team.

## Example Usage

```hcl
data "aiven_account_team" "account_team1" {
    account_id = aiven_account.<ACCOUNT_RESOURCE>.account_id
    name = "account_team1"
}
```

## Argument Reference

* `name` - (Required) defines an account team name.

* `account_id` - (Required) is a unique account id.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `team_id` - is an auto-generated unique account team id.

* `create_time` - time of creation.

* `update_time` - time of last update.