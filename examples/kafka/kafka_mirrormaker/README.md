
# Replicate data from one Apache Kafka® cluster to another using MirrorMaker 2

[Aiven for Apache Kafka® MirrorMaker 2](https://aiven.io/docs/products/kafka/kafka-mirrormaker/get-started)
is a fully managed distributed Apache Kafka® data replication utility that lets you sync your topic data across Aiven for Apache Kafka® clusters.
With MirrorMaker 2 you can define replication flows to keep a set of topics in sync between multiple Kafka clusters. This is useful for disaster recovery and isolating data for compliance reasons.

A single MirrorMaker 2 cluster can run multiple replication flows, and it has a mechanism for preventing replication cycles. This example sets up:

- a source Kafka service with a topic
- a target Kafka service with a topic
- a MirrorMaker 2 service
- two service integrations between the Kafka clusters and MirrorMaker 2
- a replication flow to move all the topics from the source to the target clusters

In the Mirrormaker 2 configuration, setting `topics` to the wildcard `".*"` means that all the topics from the source cluster will be replicated to the target cluster. Since the flow is unidirectional, `topic-b` will only be present in the target cluster, where it was originally created.

[Run the example files](https://registry.terraform.io/providers/aiven/aiven/latest/docs/guides/examples) to create the service with autoscaling and log into the Aiven Console to see the service and integration.

In the target Kafka cluster you can see:

- the topic named `topic-b`
- some internal MirrorMaker 2 topics starting with prefix `mm2`
- a heartbeat topic for the source Kafka cluster named `source.heartbeats`
- the replicated topic, `topic-a`, prefixed with the source Kafka cluster alias source
