ClickHouse integration with a Kafka source
==========================================

The goal of the example is to show how to integrate ClickHouse with Kafka topic sources.

Imagine you're ingesting messages from edges IoT devices with measurements of the following form:

.. code-block:: javascript

  {
    "sensor_id": 10000001,
    "ts": "2022-12-01T10:08:24.446369",
    "key": "pressure_pa",
    "value": 101.325
  }


``kafka.tf`` creates a Kafka service and the associated `edge-measurements` topic.
ACLs are not glossed over to focus the example on the integration.

``clickhouse.tf`` creates a ClickHouse service and the Kafka service integration, using the JSON schema presented above.
The result will be a ``service_kafka-gcp-eu`` database containing the ingested messages.

``clickhouse.tf`` also create an ``iot_analytics`` database for downstream transformation and aggregation of the raw data,
as well as analyst and writer roles with different sets of privileges over this database.
