---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_organization_user Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  The Organization User data source provides information about the existing Aiven Organization User.
---

# aiven_organization_user (Data Source)

The Organization User data source provides information about the existing Aiven Organization User.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `organization_id` (String) The unique organization ID. This property cannot be changed, doing so forces recreation of the resource.
- `user_email` (String) This is a user email address that first will be invited, and after accepting an invitation, they become a member of the organization. This property cannot be changed, doing so forces recreation of the resource.

### Read-Only

- `accepted` (Boolean) This is a boolean flag that determines whether an invitation was accepted or not by the user. `false` value means that the invitation was sent to the user but not yet accepted. `true` means that the user accepted the invitation and now a member of an organization.
- `create_time` (String) Time of creation
- `id` (String) The ID of this resource.
- `invited_by` (String) The email address of the user who sent an invitation to the user.