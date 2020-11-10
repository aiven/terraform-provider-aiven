# Account Team Member Resource

The Account Team Member resource allows the creation and management of an Aiven Account Team Member.

During the creation of `aiven_account_team_member` resource, an email invitation will be sent  
to a user using `user_email` address. If the user accepts an invitation, he or she will become 
a member of the account team. The deletion of `aiven_account_team_member` will not only 
delete the invitation if one was sent but not yet accepted by the user, it will also 
eliminate an account team member if one has accepted an invitation previously.

## Example Usage

```hcl
resource "aiven_account_team_member" "foo" {
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