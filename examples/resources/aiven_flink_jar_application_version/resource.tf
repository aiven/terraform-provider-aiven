resource "aiven_flink_jar_application_version" "example" {
  project        = "my-project" // Force new
  service_name   = "my-application" // Force new
  application_id = "foo" // Force new
  source         = "./example.jar"

  /* COMPUTED FIELDS
  application_version_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  created_at             = "2021-01-01T00:00:00Z"
  created_by             = "foo"
  file_info = [{
    file_sha256          = "foo"
    file_size            = 42
    file_status          = "FAILED"
    url                  = "foo"
    verify_error_code    = 1
    verify_error_message = "foo"
  }]
  source_checksum = "foo"
  version         = 42
  */
}
