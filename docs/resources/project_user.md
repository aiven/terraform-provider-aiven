# Project User Resource

The Project User resource allows the creation and management of Aiven Project Users.

## Example Usage

```hcl
resource "aiven_project_user" "mytestuser" {
    project = "${aiven_project.myproject.project}"
    email = "john.doe@example.com"
    member_type = "admin"
}
```

## Argument Reference

* `project` - (Required) defines the project the user is a member of.

* `email` - (Required) identifies the email address of the user.

* `member_type` - (Required) defines the access level the user has to the project.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `accepted` - is a computed property tells whether the user has accepted the request to join
the project; adding user to a project sends an invitation to the target user and the
actual membership is only created once the user accepts the invitation. This property
cannot be set, only read.

Aiven ID format when importing existing resource: `<project_name>/<email>`