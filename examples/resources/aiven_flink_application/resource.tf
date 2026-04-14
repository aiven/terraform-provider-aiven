resource "aiven_flink_application" "example" {
  project      = "my-project" // Force new
  service_name = "my-application" // Force new
  name         = "TestJob"

  /* COMPUTED FIELDS
  application_id = "foo"
  created_at     = "2021-01-01T00:00:00Z"
  created_by     = "foo"
  updated_at     = "2021-01-01T00:00:00Z"
  updated_by     = "foo"
  */
}
