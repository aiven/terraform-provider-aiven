resource "aiven_flink_jar_application" "example" {
  project      = "my-project" // Force new
  service_name = "my-application" // Force new
  name         = "TestJob"

  /* COMPUTED FIELDS
  application_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  application_versions = [{
    created_at = "2021-01-01T00:00:00Z"
    created_by = "foo"
    file_info = [{
      file_sha256          = "foo"
      file_size            = 42
      file_status          = "FAILED"
      url                  = "foo"
      verify_error_code    = 1
      verify_error_message = "foo"
    }]
    id      = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
    version = 42
  }]
  created_at = "2021-01-01T00:00:00Z"
  created_by = "foo"
  current_deployment = [{
    created_at         = "2021-01-01T00:00:00Z"
    created_by         = "foo"
    entry_class        = "foo"
    error_msg          = "foo"
    id                 = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
    job_id             = "foo"
    last_savepoint     = "foo"
    parallelism        = 42
    program_args       = ["foo"]
    starting_savepoint = "foo"
    status             = "CANCELED"
    version_id         = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  }]
  updated_at = "2021-01-01T00:00:00Z"
  updated_by = "foo"
  */
}
