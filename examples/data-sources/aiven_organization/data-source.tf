data "aiven_organization" "example" {
  // REQUIRED EXACTLY ONE
  id   = "org1a23f456789"
  name = "Aiven Ltd"

  /* COMPUTED FIELDS
  create_time = "2021-01-01T00:00:00Z"
  update_time = "2021-01-01T00:00:00Z"
  */
}
