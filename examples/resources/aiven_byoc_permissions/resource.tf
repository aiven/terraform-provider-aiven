resource "aiven_byoc_permissions" "example" {
  organization_id             = data.aiven_organization.main.id
  custom_cloud_environment_id = aiven_byoc_aws_entity.example.custom_cloud_environment_id

  accounts = [data.aiven_organization.main.id]
  projects = ["my-project"]
}
