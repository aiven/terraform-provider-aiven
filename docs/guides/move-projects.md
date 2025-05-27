---
page_title: "Move projects"
---

# Move projects

You can move a project to another [organization or organizational unit](https://aiven.io/docs/platform/concepts/orgs-units-projects) by updating the `aiven_project` or `aiven_organization_project` resource.

Services in the project continue running during the move.

## Move a project within an organization

Users with the organization admin or project admin [role](https://aiven.io/docs/platform/concepts/permissions#organization-roles-and-permissions) can move projects within an organization.

To move a project to another organizational unit in your organization, change the `parent_id` of the project resource. The following example shows an `aiven_project` resource with the `parent_id`
of the destination unit:

```hcl
resource "aiven_organizational_unit" "destination_unit" {
  name = "dest-unit"
  parent_id = data.aiven_organization.main.id
}

resource "aiven_project" "example_project" {
  project    = "example-project"
  parent_id = aiven_organizational_unit.destination_unit.id
}
```

You can also move a project out of an organizational unit and put it directly under the organization. The following example shows an `aiven_organization_project` resource
with the `parent_id` of the organization:

```hcl
resource "aiven_organization_project" "example_project" {
  project_id       = "example-project"
  organization_id  = data.aiven_organization.main.id
  parent_id        = data.aiven_organization.main.id
  billing_group_id = aiven_billing_group.main.id
}
```

## Move a project to another organization

To move a project to a different organization, you must be an [organization admin](https://aiven.io/docs/platform/concepts/permissions#organization-roles-and-permissions) of both organizations.
All users with permission to access the project lose the permissions when you move it to a different organization unless they are members of the target organization.

~> **Note**
To move an `aiven_project` resource to another organization, change the `parent_id` to the ID of the destination organization.

To move an `aiven_organization_project` resource to another organization, you have to change the:

* `organization_id` to the ID of the destination organization
* `parent_id` to the ID of the destination organization or organizational unit
* `billing_group_id` to the ID of one of the billing groups in the destination organization

-> **Tip**
You can get the billing group ID [in the Aiven Console](https://aiven.io/docs/platform/reference/get-resource-IDs) or by using
the `aiven_organization_billing_group_list` data source.

In the following example the project was moved to an organizational unit in another organization:

```hcl
# Destination organization
data "aiven_organization" "dest_org" {
  name = "Destination organization"
}

# Destination unit
data "aiven_organizational_unit" "dest_unit" {
  name      = "Example organizational unit"
  parent_id = data.aiven_organization.dest_org.id
}

# Billing groups in destination organization
data "aiven_organization_billing_group_list" "billing_groups" {
  organization_id = data.aiven_organization.dest_org.id
}

# Project
resource "aiven_organization_project" "example_project" {
  project_id       = "example-project"
  organization_id  = data.aiven_organization.dest_org.id
  parent_id        = data.aiven_organizational_unit.dest_unit.id
  billing_group_id = one([for bg in data.aiven_organization_billing_group_list.billing_groups.billing_groups : bg.billing_group_id if bg.billing_group_name == "Default billing group"])
}
```
