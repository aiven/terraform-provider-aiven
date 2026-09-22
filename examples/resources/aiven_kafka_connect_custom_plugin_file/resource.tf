resource "aiven_kafka_connect_custom_plugin_file" "example" {
  organization_id = "org1a23f456789" // Force new
  plugin_name     = "my-custom-connectors" // Force new
  plugin_version  = "2.7.14" // Force new
  source          = "./my-custom-connectors-2.7.14.jar"

  // OPTIONAL FIELDS
  content_type     = "application/java-archive" // Force new
  file_description = "Fixed memory leak in the MQTT source connector."
  service_type     = "kafka_connect" // Force new

  /* COMPUTED FIELDS
  plugin_file_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  plugin_id      = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  created_at     = "2021-01-01T00:00:00Z"
  created_by     = "foo"
  file_sha256    = "foo"
  file_size      = 42
  file_status    = "FAILED"
  plugin_classes = [{
    author            = "foo"
    created_at        = "2021-01-01T00:00:00Z"
    created_by        = "foo"
    description       = "example description"
    doc_url           = "foo"
    plugin_class      = "foo"
    plugin_class_type = "sink"
    title             = "foo"
    updated_at        = "2021-01-01T00:00:00Z"
    updated_by        = "foo"
  }]
  source_checksum      = "foo"
  updated_at           = "2021-01-01T00:00:00Z"
  updated_by           = "foo"
  verify_error_code    = 42
  verify_error_message = "foo"
  */
}
