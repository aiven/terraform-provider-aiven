# Account Team Member Data Source

The Account Team Member  data source provides information about the existing Aiven Account Team Member.

## Example Usage

```hcl
data "aiven_account_team_member" "foo" {
  account_id = aiven_account.<ACCOUNT_RESOURCE>.account_id
  team_id = aiven_account_team.<TEAM_RESOURCE>.team_id
  user_email = "user+1@example.com"
}
```

## Argument Reference

* `account_id` - (Required) is a unique account id.

* `team_id` - (Required) is an account team id.

* `user_email` - (Required) is a user email address that first will be invited, and after accepting an invitation,
he or she becomes a member of a team.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `invited_by_user_email` - team invited by user email.

* `accepted` - is a boolean flag that determines whether an invitation was accepted or not by the user. 
`false` value means that the invitation was sent to the user but not yet accepted. 
`true` means that the user accepted the invitation and now a member of an account team.
 
* `create_time` - time of creation.