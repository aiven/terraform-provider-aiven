# Enable tiered storage for Aiven for Apache Kafka®

[Aiven for Apache Kafka tiered storage](https://aiven.io/docs/products/kafka/howto/kafka-tiered-storage-get-started) optimizes resources by storing recent, frequently accessed data on faster local disks.
Tiered storage in Apache Kafka means tiering data by retention: older data is moved to remote storage, while recent data stays on the broker. When you enable tiered storage for a topic,
you specify the limit for the local retention in either time or partition size. When this limit is reached, the local copy is marked for deletion and the data is then accessible from the remote storage.

In this example you:

- Create an Aiven for Apache Kafka service with tiered storage enabled
- Create a Kafka topic
- Enable tiered storage on the topic using the `remote_storage_enable` attribute
- Set the local retention limit by time using `local_retention_ms`
- Set the size of the index using `segment_bytes`

For this example, the index size is very small to speed up closing the segments and moving them to tiered storage. This means you can quickly see tiered storage in action. It's not recommended to use such a small value for `segment_bytes`
for production systems because it will cause too many open files and performance issues. More information on the configuration options is available in the
[`aiven_kafka_topic` resource documentation](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/kafka_topic).

To create the service and topic with tiered storage, [run the example files](https://registry.terraform.io/providers/aiven/aiven/latest/docs/guides/examples) and log into the Aiven Console to see them.

## Load data

To simlulate streaming data to your topic and see it move to tiered storage, you can use the [sample data generator](https://aiven.io/docs/products/kafka/howto/generate-sample-data) in the Aiven Console.
The sample data generator produces realistic test messages to a topic. It's designed for quick onboarding with no client configuration required.

The Kafka logs have entries for data that is sent to the remote storage and the tiered storage details show remote storage used by topics.

## View the changes in the Aiven Console

To view your Aiven for Kafka service and topic:

1. In the [Aiven Console](https://console.aiven.io), go to your project.
2. Click the name of the service to open it.
3. To view the topic, click **Topics**.
4. To view tiered storage details and settings, click **Tiered storage**.
