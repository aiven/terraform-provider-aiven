resource "aiven_clickhouse" "bar" {
  project                 = "example-project"
  cloud_name              = "google-europe-west1"
  plan                    = "startup-8"
  service_name            = "example-service"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}
			
resource "aiven_clickhouse_role" "foo" {
  service_name = aiven_clickhouse.bar.service_name
  project      = aiven_clickhouse.bar.project
  role         = "writer"
}
