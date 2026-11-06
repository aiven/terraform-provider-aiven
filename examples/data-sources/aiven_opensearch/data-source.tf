data "aiven_opensearch" "example" {
  project      = "my-project"
  service_name = "my-opensearch"

  /* COMPUTED FIELDS
  cmk_id                = "12345678-1234-1234-1234-12345678abcd"
  project_vpc_id        = "my-project/1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  additional_disk_space = "30GiB"
  cloud_name            = "google-europe-west1"
  components {
    privatelink_connection_id   = "foo"
    component                   = "foo"
    connection_uri              = "foo"
    host                        = "foo"
    kafka_authentication_method = "certificate"
    kafka_ssl_ca                = "letsencrypt"
    port                        = 42
    route                       = "dynamic"
    ssl                         = true
    usage                       = "disaster_recovery"
  }
  disk_space_cap             = "foo"
  disk_space_default         = "foo"
  disk_space_step            = "foo"
  disk_space_used            = "foo"
  maintenance_window_dow     = "monday"
  maintenance_window_enabled = true
  maintenance_window_time    = "10:00:00"
  opensearch {
    opensearch_dashboards_uri = "foo"
    password                  = "password123"
    uris                      = ["foo"]
    username                  = "foo"
  }
  plan             = "startup-4"
  service_host     = "foo"
  service_password = "password123"
  service_port     = 42
  service_type     = "opensearch"
  service_uri      = "foo"
  service_username = "foo"
  state            = "POWEROFF"
  static_ips       = ["ip350351b76da"]
  tag {
    key   = "foo"
    value = "foo"
  }
  tech_emails {
    email = "jane@example.com"
  }
  termination_protection = false
  */
}
