resource "aiven_cmk" "example" {
  project      = "my-project" // Force new
  cmk_provider = "aws" // Force new
  resource     = "my-resource" // Force new

  // OPTIONAL FIELDS
  default_cmk = false

  /* COMPUTED FIELDS
  cmk_id     = "foo"
  created_at = "2021-01-01T00:00:00Z"
  status     = "current"
  updated_at = "2021-01-01T00:00:00Z"
  */
}
