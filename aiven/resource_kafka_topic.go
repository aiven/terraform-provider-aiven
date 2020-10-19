// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"
	"time"
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
		Type:         schema.TypeInt,
		Optional:     true,
		Description:  "Retention period (hours)",
		ValidateFunc: validation.IntAtLeast(-1),
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			// When a retention hours field is not set to any value and consequently is null (empty string).
			// Allow ignoring those.
			return new == ""
		},
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
	"termination_protection": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
		Description: `It is a Terraform client-side deletion protection, which prevents a Kafka 
			topic from being deleted. It is recommended to enable this for any production Kafka 
			topic containing critical data.`,
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
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Read:   schema.DefaultTimeout(4 * time.Minute),
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

	timeout := d.Timeout(schema.TimeoutCreate)
	_, err := w.Conf(timeout).WaitForState()
	if err != nil {
		return err
	}

	d.SetId(buildResourceID(project, serviceName, topicName))

	return resourceKafkaTopicRead(d, m)
}

func resourceKafkaTopicRead(d *schema.ResourceData, m interface{}) error {
	project, serviceName, topicName := splitResourceID3(d.Id())
	topic, err := getTopic(d, m)
	if err != nil {
		return err
	}

	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("topic_name", topicName); err != nil {
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

	if topic.RetentionHours != nil {
		// it could be -1, which means infinite retention
		if *topic.RetentionHours >= -1 {
			if err := d.Set("retention_hours", *topic.RetentionHours); err != nil {
				return err
			}
		}
	}

	if err := d.Set("termination_protection", d.Get("termination_protection")); err != nil {
		return err
	}

	return nil
}

func getTopic(d *schema.ResourceData, m interface{}) (aiven.KafkaTopic, error) {
	project, serviceName, topicName := splitResourceID3(d.Id())

	w := &KafkaTopicAvailabilityWaiter{
		Client:      m.(*aiven.Client),
		Project:     project,
		ServiceName: serviceName,
		TopicName:   topicName,
	}

	timeout := d.Timeout(schema.TimeoutRead)
	topic, err := w.Conf(timeout).WaitForState()
	if err != nil {
		return aiven.KafkaTopic{}, fmt.Errorf("error waiting for Aiven Kafka topic to be ACTIVE: %s", err)
	}

	return topic.(aiven.KafkaTopic), nil
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

	return resourceKafkaTopicRead(d, m)
}

func resourceKafkaTopicDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName, topicName := splitResourceID3(d.Id())

	if d.Get("termination_protection").(bool) {
		return fmt.Errorf("cannot delete kafka topic when termination_protection is enabled")
	}

	return client.KafkaTopics.Delete(projectName, serviceName, topicName)
}

func resourceKafkaTopicExists(d *schema.ResourceData, m interface{}) (bool, error) {
	_, err := getTopic(d, m)

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
