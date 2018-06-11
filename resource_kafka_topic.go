package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jelmersnoeck/aiven"
)

func resourceKafkaTopic() *schema.Resource {
	return &schema.Resource{
		Create: resourceKafkaTopicCreate,
		Read:   resourceKafkaTopicRead,
		Update: resourceKafkaTopicUpdate,
		Delete: resourceKafkaTopicDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"project": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Project to link the kafka topic to",
			},
			"service_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service to link the kafka topic to",
			},
			"topic": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Topic name",
			},
			"partitions": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Number of partitions to create in the topic",
			},
			"replication": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Replication factor for the topic",
			},
			"retention_bytes": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     -1,
				Description: "Retention bytes",
			},
			"retention_hours": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     72,
				Description: "Retention period (hours)",
			},
			"minimum_in_sync_replicas": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Minimum required nodes In Sync Replicas (ISR) to produce to a partition",
			},
			"cleanup_policy": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "delete",
				Description: "Topic cleanup policy. Allowed values: delete, compact",
			},
		},
	}
}

func resourceKafkaTopicCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	service_name := d.Get("service_name").(string)
	topic := d.Get("topic").(string)
	partitions := d.Get("partitions").(int)
	replication := d.Get("replication").(int)

	err := client.KafkaTopics.Create(
		project,
		service_name,
		aiven.CreateKafkaTopicRequest{
			optionalStringPointer(d, "cleanup_policy"),
			optionalIntPointer(d, "minimum_in_sync_replicas"),
			&partitions,
			&replication,
			optionalIntPointer(d, "retention_bytes"),
			optionalIntPointer(d, "retention_hours"),
			topic,
		},
	)
	if err != nil {
		d.SetId("")
		return err
	}

	err = resourceKafkaTopicWait(d, m)

	if err != nil {
		d.SetId("")
		return err
	}

	d.SetId(project + "/" + service_name + "/" + topic)

	return resourceKafkaTopicRead(d, m)
}

func resourceKafkaTopicRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	log.Printf("[DEBUG] reading information for kafka topic: %s", d.Id())

	project, service_name, topic_name := resourceKafkaParseTopicId(d.Id())

	topic, err := client.KafkaTopics.Get(
		project,
		service_name,
		topic_name,
	)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] topic data: %#v", topic)

	d.Set("project", project)
	d.Set("service_name", service_name)
	d.Set("topic", topic.TopicName)
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

	project := d.Get("project").(string)
	service_name := d.Get("service_name").(string)
	topic := d.Get("topic").(string)
	partitions := d.Get("partitions").(int)

	err := client.KafkaTopics.Update(
		project,
		service_name,
		topic,
		aiven.UpdateKafkaTopicRequest{
			optionalIntPointer(d, "minimum_in_sync_replicas"),
			&partitions,
			optionalIntPointer(d, "retention_bytes"),
			optionalIntPointer(d, "retention_hours"),
		},
	)
	if err != nil {
		return err
	}

	err = resourceKafkaTopicWait(d, m)

	if err != nil {
		return err
	}

	return nil
}

func resourceKafkaTopicDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	return client.KafkaTopics.Delete(
		d.Get("project").(string),
		d.Get("service_name").(string),
		d.Get("topic").(string),
	)
}

func resourceKafkaTopicWait(d *schema.ResourceData, m interface{}) error {
	w := &KafkaTopicChangeWaiter{
		Client:      m.(*aiven.Client),
		Project:     d.Get("project").(string),
		ServiceName: d.Get("service_name").(string),
		Topic:       d.Get("topic").(string),
	}

	_, err := w.Conf().WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Aiven Kafka topic to be ACTIVE: %s", err)
	}

	return nil
}

func resourceKafkaParseTopicId(id string) (project, service_name, topic string) {
	s := strings.Split(id, "/")
	return s[0], s[1], s[2]
}
