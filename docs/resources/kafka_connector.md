# Kafka connectors Resource

The Kafka connectors resource allows the creation and management of Aiven Kafka connectors.

## Example Usage

```hcl
resource "aiven_kafka_connector" "kafka-es-con1" {
  project = aiven_project.kafka-con-project1.project
  service_name = aiven_kafka.kafka-service1.service_name
  connector_name = "kafka-es-con1"

  config = {
    "topics" = aiven_kafka_topic.kafka-topic1.topic_name
    "connector.class" : "io.aiven.connect.elasticsearch.ElasticsearchSinkConnector"
    "type.name" = "es-connector"
    "name" = "kafka-es-con1"
    "connection.url" = aiven_elasticsearch.es-service1.service_uri
  }
}
```


* `project` and `service_name`- (Required) define the project and service the Kafka Connectors belongs to. 
They should be defined using reference as shown above to set up dependencies correctly.

* `connector_name`- (Required) is the Kafka connector name.

* `config`- (Required)is the Kafka Connector configuration parameters, where `topics`, `connector.class` and `name` 
are required parameters but the rest of them are connector type specific.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `plugin_author` - Kafka connector author.

* `plugin_class` - Kafka connector Java class.

* `plugin_doc_url` - Kafka connector documentation URL.

* `plugin_title` - Kafka connector title.

* `plugin_type` - Kafka connector type.

* `plugin_version` - Kafka connector version.

* `task` - List of tasks of a connector, each element contains `connector` 
(Related connector name) and `task` (Task id / number).