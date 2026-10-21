resource "aiven_flink_jar_application_deployment" "example" {
  project        = "my-project" // Force new
  service_name   = "my-application" // Force new
  application_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d" // Force new
  version_id     = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d" // Force new

  // OPTIONAL FIELDS
  entry_class        = "com.example.MyFlinkJob" // Force new
  parallelism        = 1 // Force new
  program_args       = ["example-argument"] // Force new
  restart_enabled    = true // Force new
  starting_savepoint = "path/to/savepoint" // Force new

  /* COMPUTED FIELDS
  deployment_id  = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  job_id         = "foo"
  created_at     = "2021-01-01T00:00:00Z"
  created_by     = "foo"
  error_msg      = "foo"
  last_savepoint = "foo"
  status         = "CANCELED"
  */
}
