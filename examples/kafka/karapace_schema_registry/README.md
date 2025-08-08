# Manage Aiven for Apache Kafka® with the open source Karapace schema registry

Karapace is Aiven's [open source HTTP API interface and schema registry](https://aiven.io/docs/products/kafka/karapace).

This example creates an Aiven for Apache Kafka® service and a Kafka topic. It sets up Karapace on the Kafka service by enabling the schema registry and the REST API.
The REST API feature lets you produce and consume messages over HTTP. It also enables automatic topic creation, which allows you to send messages to topics
that don't already exist on the Kafka cluster.

[Run the example files](https://registry.terraform.io/providers/aiven/aiven/latest/docs/guides/examples) to set up Karapace and log into the Aiven Console to see the service.

More information on using Karapace is available in the [Karapace documentation](https://www.karapace.io/quickstart) and [Aiven docs](https://aiven.io/docs/products/kafka/karapace).
