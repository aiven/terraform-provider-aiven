# Enable and configure SASL authentication for Aiven for Apache Kafka®

Aiven for Apache Kafka lets you use Simple Authentication and Security Layer (SASL) over SSL [to secure your data in transit](https://aiven.io/docs/products/kafka/concepts/auth-types).

This example creates a Kafka service with the SASL authentication method enabled and configures SCRAM-SHA-256 as the SASL mechanism. It outputs the port number for SASL authentication using a public CA.

You can also enable the public CA over a PivateLink connection by setting `letsencrypt_sasl_privatelink` to true in the `kafka_user_config`.
