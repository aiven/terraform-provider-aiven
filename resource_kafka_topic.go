package main

import (
  "log"

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
				Description: "TBD",
			},
      "retention_hours": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "TBD",
			},
      "minimum_in_sync_replicas": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
        Default:     1,
				Description: "TBD",
			},
      "cleanup_policy": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
        Default:     "delete",
				Description: "TBD",
			},
		},
	}
}

func resourceKafkaTopicCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

  topic := d.Get("topic").(string)
  partitions := d.Get("partitions").(int)
  replication := d.Get("replication").(int)

	err := client.KafkaTopics.Create(
		d.Get("project").(string),
		d.Get("service_name").(string),
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

	d.SetId(topic + "!")

  // TODO: actually wait that the topic has been created

	//return resourceKafkaTopicRead(d, m)
  return nil
}

func resourceKafkaTopicRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

  log.Printf("[DEBUG] reading information for kafka topic: %s", d.Get("topic").(string))

	topic, err := client.KafkaTopics.Get(
		d.Get("project").(string),
		d.Get("service_name").(string),
    d.Get("topic").(string),
	)
	if err != nil {
		return err
	}

	d.Set("name", topic.TopicName)
	d.Set("state", topic.State)
	d.Set("replication", topic.Replication)
  // TODO: add the other fields

	return nil
}

func resourceKafkaTopicUpdate(d *schema.ResourceData, m interface{}) error {
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
