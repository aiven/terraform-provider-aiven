data "aiven_organization" "example" {
  name = "my-organization"
}

resource "aiven_kafka_connect_custom_plugin_file" "example" {
  organization_id    = data.aiven_organization.example.id
  plugin_name        = "my-custom-connectors"
  plugin_version     = "2.7.14"
  service_type       = "kafka_connect"
  source             = "./my-custom-connectors-2.7.14.jar"
  plugin_description = "My custom connector bundle for stream processing."
  file_description   = "Initial release."
}
