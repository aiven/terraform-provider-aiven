---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_alloydbomni_user Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Aiven for AlloyDB Omni service user.
  This resource is in the beta stage and may change without notice. Set
  the PROVIDER_AIVEN_ENABLE_BETA environment variable to use the resource.
---

# aiven_alloydbomni_user (Resource)

Creates and manages an Aiven for AlloyDB Omni service user.

**This resource is in the beta stage and may change without notice.** Set
the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the resource.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project` (String) The name of the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `service_name` (String) The name of the service that this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `username` (String) The name of the service user for this service. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.

### Optional

- `password` (String, Sensitive) The password of the service user.
- `pg_allow_replication` (Boolean) Allows replication. For the default avnadmin user this attribute is required and is always `true`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `access_cert` (String, Sensitive) The access certificate for the servie user.
- `access_key` (String, Sensitive) The access certificate key for the service user.
- `id` (String) The ID of this resource.
- `type` (String) The service user account type, either primary or regular.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `default` (String)
- `delete` (String)
- `read` (String)
- `update` (String)
