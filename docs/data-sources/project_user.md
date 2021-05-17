# Project User Data Source

The Project User data source provides information about the existing Aiven Project User.

## Example Usage

```hcl
data "aiven_project_user" "mytestuser" {
    project = aiven_project.myproject.project
    email = "john.doe@example.com"
}
```

## Argument Reference

* `project` - (Required) defines the project the user is a member of.

* `email` - (Required) identifies the email address of the user.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `member_type` - (Required) defines the access level the user has to the project.

* `accepted` - is a computed property tells whether the user has accepted the request to join
the project; adding user to a project sends an invitation to the target user and the
actual membership is only created once the user accepts the invitation. This property
cannot be set, only read.

Aiven ID format when importing existing resource: `<project_name>/<email>`