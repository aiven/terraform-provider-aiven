data "aiven_project_vpc" "example" {
  // LOOKUP — exactly one of: `project_vpc_id` or `cloud_name`
  project        = "my-project"
  project_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  // cloud_name  = "aws-eu-central-1"

  /* COMPUTED FIELDS
  network_cidr = "192.168.6.0/24"
  state        = "ACTIVE"
  */
}
