---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_governance_access Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Request access to an Apache Kafka topic in Aiven for Apache Kafka® Governance. Governance https://aiven.io/docs/products/kafka/howto/governance helps you manage your Kafka clusters securely and efficiently through structured policies, roles, and processes. You can manage approval workflows using Terraform and GitHub Actions https://aiven.io/docs/products/kafka/howto/terraform-governance-approvals.
  This resource is in the beta stage and may change without notice. Set
  the PROVIDER_AIVEN_ENABLE_BETA environment variable to use the resource.
---

# aiven_governance_access (Resource)

Request access to an Apache Kafka topic in Aiven for Apache Kafka® Governance. [Governance](https://aiven.io/docs/products/kafka/howto/governance) helps you manage your Kafka clusters securely and efficiently through structured policies, roles, and processes. You can [manage approval workflows using Terraform and GitHub Actions](https://aiven.io/docs/products/kafka/howto/terraform-governance-approvals). 

**This resource is in the beta stage and may change without notice.** Set
the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the resource.

## Example Usage

```terraform
resource "aiven_governance_access" "example_access" {
 organization_id = data.aiven_organization.main.id
 access_name     = "example-topic-access"
 access_type     = "KAFKA"

 access_data {
   project      = data.aiven_project.example_project.project
   service_name = aiven_kafka.example_kafka.service_name

   acls {
     resource_name   = "example-topic"
     resource_type   = "Topic"
     operation       = "Read"
     permission_type = "ALLOW"
     host            = "*"
   }
 }

 owner_user_group_id = aiven_organization_user_group.example.group_id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `access_data` (Block List, Min: 1, Max: 1) Details of the access. Changing this property forces recreation of the resource. (see [below for nested schema](#nestedblock--access_data))
- `access_name` (String) The name to describe the access. Maximum length: `54`. Changing this property forces recreation of the resource.
- `access_type` (String) The type of access. The possible value is `KAFKA`. Changing this property forces recreation of the resource.
- `organization_id` (String) The ID of the organization. Changing this property forces recreation of the resource.

### Optional

- `owner_user_group_id` (String) The ID of the user group that owns the access. Maximum length: `54`. Changing this property forces recreation of the resource.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.
- `susbcription_id` (String) The ID of the access.

<a id="nestedblock--access_data"></a>
### Nested Schema for `access_data`

Required:

- `acls` (Block Set, Min: 1, Max: 10) The permissions granted to the assigned service user. Maximum length: `54`. Changing this property forces recreation of the resource. (see [below for nested schema](#nestedblock--access_data--acls))
- `project` (String) The name of the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `service_name` (String) The name of the service that this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.

Optional:

- `username` (String) The name for the new service user given access. If not provided, the name is automatically generated. Maximum length: `54`. Changing this property forces recreation of the resource.

<a id="nestedblock--access_data--acls"></a>
### Nested Schema for `access_data.acls`

Required:

- `operation` (String) The action that will be allowed for the service user. The possible values are `Read` and `Write`. Changing this property forces recreation of the resource.
- `permission_type` (String) Explicitly allows or denies the action for the service user on the specified resource. The possible value is `ALLOW`. Changing this property forces recreation of the resource.
- `resource_name` (String) The name of the resource the permission applies to, such as the topic name or group ID in the Kafka service. Maximum length: `256`. Changing this property forces recreation of the resource.
- `resource_type` (String) The type of resource. The possible value is `Topic`. Changing this property forces recreation of the resource.

Optional:

- `host` (String) The IP address from which a principal is allowed or denied access to the resource. Use `*` for all hosts. Maximum length: `256`. Changing this property forces recreation of the resource.

Read-Only:

- `id` (String) The ACL ID.
- `pattern_type` (String) Pattern used to match specified resources. The possible value is `LITERAL`.
- `principal` (String) Identities in `user:name` format that the permissions apply to.



<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `default` (String)
- `delete` (String)
- `read` (String)
- `update` (String)
