# Resource Kafka ACL Resource

The Resource Kafka ACL resource allows the creation and management of ACLs for an Aiven Kafka service.

## Example Usage

```hcl
resource "aiven_kafka_acl" "mytestacl" {
    project = "${aiven_project.myproject.project}"
    service_name = "${aiven_service.myservice.service_name}"
    topic = "<TOPIC_NAME_PATTERN>"
    permission = "admin"
    username = "<USERNAME_PATTERN>"
}
```

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the ACL belongs to.
They should be defined using reference as shown above to set up dependencies correctly.
These properties cannot be changed once the service is created. Doing so will result in
the topic being deleted and new one created instead.

* `topic` - (Required) is a topic name pattern the ACL entry matches to.

* `permission` - (Required) is the level of permission the matching users are given to the matching
topics (admin, read, readwrite, write).

* `username` - (Required) is a username pattern the ACL entry matches to.

Aiven ID format when importing existing resource: `<project_name>/<service_name>/<acl_id>`.
The ACL ID is not directly visible in the Aiven web console.