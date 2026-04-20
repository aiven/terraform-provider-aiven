data "aiven_flink_application" "example" {
  project      = "my-project"
  service_name = "my-application"

  // REQUIRED EXACTLY ONE
  application_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  // name        = "TestJob"

  /* COMPUTED FIELDS
  created_at = "2021-01-01T00:00:00Z"
  created_by = "foo"
  updated_at = "2021-01-01T00:00:00Z"
  updated_by = "foo"
  */
}
