resource "aiven_byoc_aws_provision" "example" {
  organization_id             = data.aiven_organization.main.id
  custom_cloud_environment_id = aiven_byoc_aws_entity.example.custom_cloud_environment_id
  aws_iam_role_arn            = aws_iam_role.aiven_byoc.arn

  depends_on = [
    aws_iam_role_policy_attachment.aiven_byoc,
  ]
}
