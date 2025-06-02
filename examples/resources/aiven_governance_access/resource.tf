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
