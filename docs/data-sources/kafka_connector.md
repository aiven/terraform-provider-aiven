# Kafka connector Data Source

The Kafka connector data source provides information about the existing Aiven Kafka connector.

## Example Usage

```hcl
data "aiven_kafka_connector" "kafka-es-con1" {
    project = aiven_project.kafka-con-project1.project
    service_name = aiven_service.kafka-service1.service_name
    connector_name = "kafka-es-con1"
}
```


* `project` and `service_name`- (Required) define the project and service the Kafka Connectors belongs to. 
They should be defined using reference as shown above to set up dependencies correctly.

* `connector_name`- (Required) is the Kafka connector name.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `config`- is the Kafka Connector configuration parameters, where `topics`, `connector.class` and `name` 
are required parameters but the rest of them are connector type specific.

* `plugin_author` - Kafka connector author.

* `plugin_class` - Kafka connector Java class.

* `plugin_doc_url` - Kafka connector documentation URL.

* `plugin_title` - Kafka connector title.

* `plugin_type` - Kafka connector type.

* `plugin_version` - Kafka connector version.

* `task` - List of tasks of a connector, each element contains `connector` 
(Related connector name) and `task` (Task id / number).