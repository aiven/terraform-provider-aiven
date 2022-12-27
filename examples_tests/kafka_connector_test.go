//go:build all || examples

package examples

import (
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
	s.T().Parallel()

	// Given
	kafkaServiceName := randName("test-examples-kafka-%s")
	kafkaConnectorName := randName("test-examples-kafka-connector-%s")
	kafkaTopicName := randName("test-examples-kafka-topic-%s")
	osServiceName := randName("test-examples-os-%s")
	opts := s.withDefaults(&terraform.Options{
		TerraformDir: "../examples/kafka_connectors/os_sink",
		Vars: map[string]interface{}{
			"avn_token":            s.config.Token,
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
	kafkaService, err := s.client.Services.Get(s.config.Project, kafkaServiceName)
	s.NoError(err)
	s.Equal("kafka", kafkaService.Type)
	s.Equal("business-4", kafkaService.Plan)
	s.Equal("google-europe-west1", kafkaService.CloudName)

	kafkaTopic, err := s.client.KafkaTopics.Get(s.config.Project, kafkaServiceName, kafkaTopicName)
	s.NoError(err)
	s.Equal(kafkaTopic.TopicName, kafkaTopicName)

	kafkaConnector, err := s.client.KafkaConnectors.GetByName(s.config.Project, kafkaServiceName, kafkaConnectorName)
	s.NoError(err)
	s.Equal(kafkaConnector.Name, kafkaConnectorName)

	osService, err := s.client.Services.Get(s.config.Project, osServiceName)
	s.NoError(err)
	s.Equal("opensearch", osService.Type)
	s.Equal("startup-4", osService.Plan)
	s.Equal("google-europe-west1", osService.CloudName)
}
