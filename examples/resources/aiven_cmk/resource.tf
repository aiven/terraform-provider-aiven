resource "aiven_cmk" "example" {
  project      = "my-project" // Force new
  cmk_provider = "aws" // Force new
  resource     = "my-resource" // Force new

  // OPTIONAL FIELDS
  default_cmk = false

  /* COMPUTED FIELDS
  cmk_id     = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  created_at = "2021-01-01T00:00:00Z"
  status     = "current"
  updated_at = "2021-01-01T00:00:00Z"
  */
}
