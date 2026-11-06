resource "aiven_pg" "example" {
  project      = "my-project" // Force new
  service_name = "my-pg" // Force new
  plan         = "startup-4"

  // OPTIONAL FIELDS
  cmk_id                  = "12345678-1234-1234-1234-12345678abcd"
  project_vpc_id          = "my-project/1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  additional_disk_space   = "30GiB"
  cloud_name              = "google-europe-west1"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  static_ips              = ["ip350351b76da"]
  tag {
    key   = "foo"
    value = "foo"
  }
  tech_emails {
    email = "jane@example.com"
  }
  termination_protection = false

  /* COMPUTED FIELDS
  components = [{
    component                   = "foo"
    connection_uri              = "foo"
    host                        = "foo"
    kafka_authentication_method = "certificate"
    kafka_ssl_ca                = "letsencrypt"
    port                        = 42
    privatelink_connection_id   = "foo"
    route                       = "dynamic"
    ssl                         = true
    usage                       = "disaster_recovery"
  }]
  disk_space_cap             = "foo"
  disk_space_default         = "foo"
  disk_space_step            = "foo"
  disk_space_used            = "foo"
  maintenance_window_enabled = true
  pg = [{
    dbname          = "foo"
    host            = "foo"
    max_connections = 42
    params = [{
      database_name = "foo"
      host          = "foo"
      password      = "password123"
      port          = 42
      sslmode       = "foo"
      user          = "foo"
    }]
    password     = "password123"
    port         = 42
    replica_uri  = "foo"
    sslmode      = "foo"
    standby_uris = ["foo"]
    syncing_uris = ["foo"]
    uri          = "foo"
    uris         = ["foo"]
    user         = "foo"
  }]
  service_host     = "foo"
  service_password = "password123"
  service_port     = 42
  service_type     = "pg"
  service_uri      = "foo"
  service_username = "foo"
  state            = "POWEROFF"
  */
}
