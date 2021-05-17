# Service User Resource

The Service User resource allows the creation and management of Aiven Service Users.

## Example Usage

```hcl
resource "aiven_service_user" "myserviceuser" {
  project = aiven_project.myproject.project
  service_name = aiven_service.myservice.service_name
  username = "<USERNAME>"
}
```

~> **Note** The service user resource is not supported for Aiven Grafana services.

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the user belongs to. They should be defined
  using reference as shown above to set up dependencies correctly.

* `username` - (Required) is the actual name of the user account.

* `password` - (Optional) is the password of the service user (not applicable for all services), the Terraform user can
  set that.

* `redis_acl_categories` - (Optional) Redis specific field, defines command category rules.

* `redis_acl_commands` - (Optional) Redis specific field, defines rules for individual commands.

* `redis_acl_keys` - (Optional) Redis specific field, defines key access rules.

## Attribute Reference

Service users have several computed properties that cannot be set, only read:

* `access_cert` - is the access certificate of the user (not applicable for all services).

* `access_key` - is the access key of the user (not applicable for all services).

* `type` - tells whether the user is primary account or regular account.

Aiven ID format when importing existing resource: `<project_name>/<service_name>/<username>`