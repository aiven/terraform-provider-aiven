// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/pkg/cache"
	"github.com/hashicorp/terraform/helper/schema"
)

var aivenKafkaTopicSchema = map[string]*schema.Schema{
	"project": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Project to link the kafka topic to",
		ForceNew:    true,
	},
	"service_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Service to link the kafka topic to",
		ForceNew:    true,
	},
	"topic_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Topic name",
		ForceNew:    true,
	},
	"partitions": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Number of partitions to create in the topic",
	},
	"replication": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Replication factor for the topic",
	},
	"retention_bytes": {
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     -1,
		Description: "Retention bytes",
	},
	"retention_hours": {
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     72,
		Description: "Retention period (hours)",
	},
	"minimum_in_sync_replicas": {
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     1,
		Description: "Minimum required nodes in-sync replicas (ISR) to produce to a partition",
	},
	"cleanup_policy": {
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "delete",
		Description: "Topic cleanup policy. Allowed values: delete, compact",
		ForceNew:    true,
	},
}

func resourceKafkaTopic() *schema.Resource {
	return &schema.Resource{
		Create: resourceKafkaTopicCreate,
		Read:   resourceKafkaTopicRead,
		Update: resourceKafkaTopicUpdate,
		Delete: resourceKafkaTopicDelete,
		Exists: resourceKafkaTopicExists,
		Importer: &schema.ResourceImporter{
			State: resourceKafkaTopicState,
		},

		Schema: aivenKafkaTopicSchema,
	}
}

func resourceKafkaTopicCreate(d *schema.ResourceData, m interface{}) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	topicName := d.Get("topic_name").(string)
	partitions := d.Get("partitions").(int)
	replication := d.Get("replication").(int)

	createRequest := aiven.CreateKafkaTopicRequest{
		CleanupPolicy:         optionalStringPointer(d, "cleanup_policy"),
		MinimumInSyncReplicas: optionalIntPointer(d, "minimum_in_sync_replicas"),
		Partitions:            &partitions,
		Replication:           &replication,
		RetentionBytes:        optionalIntPointer(d, "retention_bytes"),
		RetentionHours:        optionalIntPointer(d, "retention_hours"),
		TopicName:             topicName,
	}
	w := &KafkaTopicCreateWaiter{
		Client:        m.(*aiven.Client),
		Project:       project,
		ServiceName:   serviceName,
		CreateRequest: createRequest,
	}
	_, err := w.Conf().WaitForState()
	if err != nil {
		return err
	}

	err = resourceKafkaTopicWait(d, m)

	if err != nil {
		return err
	}

	d.SetId(buildResourceID(project, serviceName, topicName))

	return resourceKafkaTopicRead(d, m)
}

func resourceKafkaTopicRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project, serviceName, topicName := splitResourceID3(d.Id())
	topic, err := cache.TopicCache{}.Read(project, serviceName, topicName, client)
	if err != nil {
		return err
	}

	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("topic_name", topic.TopicName); err != nil {
		return err
	}
	if err := d.Set("partitions", len(topic.Partitions)); err != nil {
		return err
	}
	if err := d.Set("replication", topic.Replication); err != nil {
		return err
	}
	if err := d.Set("cleanup_policy", topic.CleanupPolicy); err != nil {
		return err
	}
	if err := d.Set("minimum_in_sync_replicas", topic.MinimumInSyncReplicas); err != nil {
		return err
	}
	if err := d.Set("retention_bytes", topic.RetentionBytes); err != nil {
		return err
	}
	if err := d.Set("retention_hours", topic.RetentionHours); err != nil {
		return err
	}

	return nil
}

func resourceKafkaTopicUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	partitions := d.Get("partitions").(int)
	projectName, serviceName, topicName := splitResourceID3(d.Id())
	err := client.KafkaTopics.Update(
		projectName,
		serviceName,
		topicName,
		aiven.UpdateKafkaTopicRequest{
			MinimumInSyncReplicas: optionalIntPointer(d, "minimum_in_sync_replicas"),
			Partitions:            &partitions,
			Replication:           optionalIntPointer(d, "replication"),
			RetentionBytes:        optionalIntPointer(d, "retention_bytes"),
			RetentionHours:        optionalIntPointer(d, "retention_hours"),
		},
	)
	if err != nil {
		return err
	}

	err = resourceKafkaTopicWait(d, m)
	if err != nil {
		return err
	}

	return resourceKafkaTopicRead(d, m)
}

func resourceKafkaTopicDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName, topicName := splitResourceID3(d.Id())
	return client.KafkaTopics.Delete(projectName, serviceName, topicName)
}

func resourceKafkaTopicExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	projectName, serviceName, topicName := splitResourceID3(d.Id())
	_, err := cache.TopicCache{}.Read(projectName, serviceName, topicName, client)
	return resourceExists(err)
}

func resourceKafkaTopicState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<topic_name>", d.Id())
	}

	err := resourceKafkaTopicRead(d, m)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourceKafkaTopicWait(d *schema.ResourceData, m interface{}) error {
	w := &KafkaTopicChangeWaiter{
		Client:      m.(*aiven.Client),
		Project:     d.Get("project").(string),
		ServiceName: d.Get("service_name").(string),
		TopicName:   d.Get("topic_name").(string),
	}

	_, err := w.Conf().WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for Aiven Kafka topic to be ACTIVE: %s", err)
	}

	return nil
}
