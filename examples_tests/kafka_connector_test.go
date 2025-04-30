//go:build all || examples

package examples

import (
	"context"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/suite"
)

type KafkaConnectorTestSuite struct {
	BaseTestSuite
}

func TestKafkaConnectorTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaConnectorTestSuite))
}

func (s *KafkaConnectorTestSuite) TestKafkaConnectorOS() {
	// Given
	withPrefix := examplesRandPrefix()
	kafkaServiceName := withPrefix("kafka")
	kafkaConnectorName := withPrefix("kafka-connector")
	kafkaTopicName := withPrefix("kafka-topic")
	osServiceName := withPrefix("os")
	opts := s.withDefaults(&terraform.Options{
		TerraformDir: "../examples/kafka_connectors/os_sink",
		Vars: map[string]interface{}{
			"aiven_token":          s.config.Token,
			"avn_project":          s.config.Project,
			"kafka_name":           kafkaServiceName,
			"kafka_connector_name": kafkaConnectorName,
			"kafka_topic_name":     kafkaTopicName,
			"os_name":              osServiceName,
		},
	})

	// When
	defer terraform.Destroy(s.T(), opts)
	terraform.Apply(s.T(), opts)

	// Then
	ctx := context.Background()

	kafkaService, err := s.client.Services.Get(ctx, s.config.Project, kafkaServiceName)
	s.NoError(err)
	s.Equal("kafka", kafkaService.Type)
	s.Equal("business-4", kafkaService.Plan)
	s.Equal("google-europe-west1", kafkaService.CloudName)

	kafkaConnector, err := s.client.KafkaConnectors.GetByName(ctx, s.config.Project, kafkaServiceName, kafkaConnectorName)
	s.NoError(err)
	s.Equal(kafkaConnector.Name, kafkaConnectorName)

	osService, err := s.client.Services.Get(ctx, s.config.Project, osServiceName)
	s.NoError(err)
	s.Equal("opensearch", osService.Type)
	s.Equal("startup-4", osService.Plan)
	s.Equal("google-europe-west1", osService.CloudName)
}
