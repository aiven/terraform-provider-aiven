resource "aiven_aws_privatelink" "example" {
  project      = "my-project" // Force new
  service_name = "foo" // Force new
  principals   = ["arn:aws:iam::012345678901:root"]

  /* COMPUTED FIELDS
  aws_service_id   = "foo"
  aws_service_name = "my-aws-service-name"
  state            = "active"
  */
}
