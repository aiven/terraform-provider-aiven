resource "aiven_azure_privatelink" "foo" {
  project      = data.aiven_project.foo.project
  service_name = aiven_kafka.bar.service_name

  user_subscription_ids = [
    "xxxxxx-xxxx-xxxx-xxxx-xxxxxxxx"
  ]
}
