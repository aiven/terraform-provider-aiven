data "aiven_aws_privatelink" "example" {
  project      = "my-project"
  service_name = "foo"

  /* COMPUTED FIELDS
  aws_service_id   = "foo"
  aws_service_name = "my-aws-service-name"
  principals       = ["arn:aws:iam::012345678901:root"]
  state            = "active"
  */
}
