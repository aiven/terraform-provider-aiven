---
page_title: "aiven_service_component Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  The Service Component data source provides information about the existing Aiven service Component.
---
# aiven_service_component (Data Source)
The Service Component data source provides information about the existing Aiven service Component.

Service components can be defined to get the connection info for specific service. Services may support multiple different access routes (VPC peering and public access), have additional components or support various authentication methods. Each of these may be represented by different DNS name or TCP port and the specific component to match can be selected by specifying appropriate filters as shown below.

## Example Usage
```terraform
data "aiven_service_component" "sc1" {
  project                     = aiven_kafka.project1.project
  service_name                = aiven_kafka.service1.service_name
  component                   = "kafka"
  route                       = "dynamic"
  kafka_authentication_method = "certificate"
}
```
<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `component` (String) Service component name
- `project` (String) Project name

### Optional

- `kafka_authentication_method` (String) Kafka authentication method. This is a value specific to the 'kafka' service component. The possible values are `certificate` and `sasl`.
- `route` (String) Network access route. The possible values are `dynamic`, `private`, `privatelink` and `public`.
- `service_name` (String) Service name
- `ssl` (Boolean) Whether the endpoint is encrypted or accepts plaintext. By default endpoints are always encrypted and this property is only included for service components that may disable encryption
- `usage` (String) DNS usage name. The possible values are `disaster_recovery`, `primary` and `replica`.

### Read-Only

- `host` (String) DNS name for connecting to the service component
- `id` (String) The ID of this resource.
- `kafka_ssl_ca` (String) Kafka certificate used. The possible values are `letsencrypt` and `project_ca`.
- `port` (Number) Port number for connecting to the service component
