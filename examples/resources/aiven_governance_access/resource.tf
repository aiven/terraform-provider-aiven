resource "aiven_governance_access" "example" {
  organization_id = "org1a23f456789" // Force new
  access_data {
    acls {
      operation       = "Write"
      permission_type = "ALLOW"
      resource_name   = "events"
      resource_type   = "Topic"

      // OPTIONAL FIELDS
      host = "*"

      /* COMPUTED FIELDS
      id           = "foo"
      pattern_type = "LITERAL"
      principal    = "foo"
      */
    }
    project_name = "project-1"
    service_name = "service-1"

    // OPTIONAL FIELDS
    username = "api3"
  }
  access_name = "My Access" // Force new
  access_type = "KAFKA" // Force new

  // OPTIONAL FIELDS
  owner_user_group_id = "ug22ba494e096" // Force new

  /* COMPUTED FIELDS
  access_id = "foo"
  */
}
