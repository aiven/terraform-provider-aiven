# Use Debezium as a source connector to integrate Aiven for PostgreSQL® and Aiven for Apache Kafka®

This example uses Debezium to capture changes in a PostgreSQL database. It creates a PostgreSQL service, a Kafka service, and an Aiven for Apache Kafka Connect service. The Kafka and Kafka Connect services are deployed in Google Cloud, and the PostgreSQL service is deployed in Azure. It also enables and configures a Debezium source connector for the Kafka Connect service to monitor for table changes and produce messages with information about the change to a Kafka topic.
