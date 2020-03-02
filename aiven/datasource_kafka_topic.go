// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceKafkaTopic() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceKafkaTopicRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenKafkaTopicSchema, "project", "service_name", "topic_name"),
	}
}

func datasourceKafkaTopicRead(d *schema.ResourceData, m interface{}) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	topicName := d.Get("topic_name").(string)

	d.SetId(buildResourceID(projectName, serviceName, topicName))

	return resourceKafkaTopicRead(d, m)
}
