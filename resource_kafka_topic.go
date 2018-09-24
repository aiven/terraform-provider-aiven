// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package main

import (
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

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

		Schema: map[string]*schema.Schema{
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
		},
	}
}

func resourceKafkaTopicCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	topicName := d.Get("topic_name").(string)
	partitions := d.Get("partitions").(int)
	replication := d.Get("replication").(int)

	err := client.KafkaTopics.Create(
		project,
		serviceName,
		aiven.CreateKafkaTopicRequest{
			CleanupPolicy:         optionalStringPointer(d, "cleanup_policy"),
			MinimumInSyncReplicas: optionalIntPointer(d, "minimum_in_sync_replicas"),
			Partitions:            &partitions,
			Replication:           &replication,
			RetentionBytes:        optionalIntPointer(d, "retention_bytes"),
			RetentionHours:        optionalIntPointer(d, "retention_hours"),
			TopicName:             topicName,
		},
	)
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
	topic, err := client.KafkaTopics.Get(project, serviceName, topicName)
	if err != nil {
		return err
	}

	d.Set("project", project)
	d.Set("service_name", serviceName)
	d.Set("topic_name", topic.TopicName)
	d.Set("state", topic.State)
	d.Set("partitions", len(topic.Partitions))
	d.Set("replication", topic.Replication)
	d.Set("cleanup_policy", topic.CleanupPolicy)
	d.Set("minimum_in_sync_replicas", topic.MinimumInSyncReplicas)
	d.Set("retention_bytes", topic.RetentionBytes)
	d.Set("retention_hours", topic.RetentionHours)

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
	_, err := client.KafkaTopics.Get(projectName, serviceName, topicName)
	return resourceExists(err)
}

func resourceKafkaTopicState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("Invalid identifier %v, expected <project_name>/<service_name>/<topic_name>", d.Id())
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
		return fmt.Errorf("Error waiting for Aiven Kafka topic to be ACTIVE: %s", err)
	}

	return nil
}
