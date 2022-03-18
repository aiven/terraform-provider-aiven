---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_kafka_user Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  The Kafka User resource allows the creation and management of Aiven Kafka Users.
---

# aiven_kafka_user (Resource)

The Kafka User resource allows the creation and management of Aiven Kafka Users.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **project** (String) Identifies the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. This property cannot be changed, doing so forces recreation of the resource.
- **service_name** (String) Specifies the name of the service that this resource belongs to. To set up proper dependencies please refer to this variable as a reference. This property cannot be changed, doing so forces recreation of the resource.
- **username** (String) The actual name of the Kafka User. To set up proper dependencies please refer to this variable as a reference. This property cannot be changed, doing so forces recreation of the resource.

### Optional

- **id** (String) The ID of this resource.
- **password** (String, Sensitive) The password of the Kafka User.

### Read-Only

- **access_cert** (String, Sensitive) Access certificate for the user
- **access_key** (String, Sensitive) Access certificate key for the user
- **type** (String) Type of the user account. Tells whether the user is the primary account or a regular account.

