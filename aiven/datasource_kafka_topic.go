// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceKafkaTopic() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceKafkaTopicRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenKafkaTopicSchema, "project", "service_name", "topic_name"),
	}
}

func datasourceKafkaTopicRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	topicName := d.Get("topic_name").(string)

	topic, err := client.KafkaTopics.Get(projectName, serviceName, topicName)
	if err != nil {
		return err
	}

	d.SetId(buildResourceID(projectName, serviceName, topicName))
	d.Set("project", projectName)
	d.Set("service_name", serviceName)
	d.Set("topic_name", topicName)
	d.Set("state", topic.State)
	d.Set("partitions", len(topic.Partitions))
	d.Set("replication", topic.Replication)
	d.Set("cleanup_policy", topic.CleanupPolicy)
	d.Set("minimum_in_sync_replicas", topic.MinimumInSyncReplicas)
	d.Set("retention_bytes", topic.RetentionBytes)
	d.Set("retention_hours", topic.RetentionHours)

	return nil
}
