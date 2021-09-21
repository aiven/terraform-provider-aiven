# Opensearch ACL Config Resource

The Opensearch ACL Config resource allows the configuration of ACL management on an Aiven Opensearch service.

## Example Usage

```hcl
resource "aiven_opensearch_acl_config" "os-acl-config" {
  project = aiven_project.os-project.project
  service_name = aiven_opensearch.os.service_name
  enabled = true
  extended_acl = false
}
```

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the ACL belongs to. They should be defined
  using reference as shown above to set up dependencies correctly.

All other properties except `project` and `service_name` can be changed after creation of the resource and will not
trigger recreation of Opensearch entire ACL's.

* `enabled` - (Optional) enables of disables Opensearch ACL's.

* `extended_acl` - (Optional) Index rules can be applied in a limited fashion to the _mget, _msearch and _bulk APIs
  (and only those) by enabling the ExtendedAcl option for the service. When it is enabled, users can use these APIs as
  long as all operations only target indexes they have been granted access to.
