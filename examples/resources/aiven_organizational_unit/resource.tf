resource "aiven_organizational_unit" "example_unit" {
  name      = "Example organizational unit"
  parent_id = aiven_organization.main.id
}