resource "aiven_pg_database" "example" {
  project       = "my-project" // Force new
  service_name  = "my-pg" // Force new
  database_name = "testdb" // Force new

  // OPTIONAL FIELDS
  lc_collate = "en_US.UTF-8" // Force new
  lc_ctype   = "en_US.UTF-8" // Force new
}
