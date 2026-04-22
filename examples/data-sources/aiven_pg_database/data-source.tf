data "aiven_pg_database" "example" {
  project       = "my-project"
  service_name  = "my-pg"
  database_name = "testdb"

  /* COMPUTED FIELDS
  lc_collate = "en_US.UTF-8"
  lc_ctype   = "en_US.UTF-8"
  */
}
