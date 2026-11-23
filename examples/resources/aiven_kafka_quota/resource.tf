resource "aiven_kafka_quota" "example" {
  project      = "my-project" // Force new
  service_name = "my-kafka" // Force new

  // OPTIONAL FIELDS
  client_id          = "abc123" // Force new
  user               = "alice" // Force new
  consumer_byte_rate = 1073741824
  producer_byte_rate = 1073741824
  request_percentage = 100
}
