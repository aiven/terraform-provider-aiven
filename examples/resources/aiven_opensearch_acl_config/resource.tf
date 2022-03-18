data "aiven_project" "foo" {
		  project = "example_project"
}
		
resource "aiven_opensearch" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "example_service_name"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_service_user" "foo" {
  service_name = aiven_opensearch.bar.service_name
  project      = data.aiven_project.foo.project
  username     = "user-example"
}

resource "aiven_opensearch_acl_config" "foo" {
  project      = data.aiven_project.foo.project
  service_name = aiven_opensearch.bar.service_name
  enabled      = true
  extended_acl = false
}
