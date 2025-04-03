//go:build all || examples

package examples

import (
	"context"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/suite"
)

type KafkaConnectTestSuite struct {
	BaseTestSuite
}

func TestKafkaConnectTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaConnectTestSuite))
}

func (s *KafkaConnectTestSuite) TestKafkaConnect() {
	s.T().Parallel()

	// Given
	withPrefix := examplesRandPrefix()
	kafkaServiceName := withPrefix("kafka")
	kafkaConnectName := withPrefix("kafka-connect")
	opts := s.withDefaults(&terraform.Options{
		TerraformDir: "../examples/kafka_connect",
		Vars: map[string]interface{}{
			"aiven_token":        s.config.Token,
			"aiven_project_name": s.config.Project,
			"kafka_service_name": kafkaServiceName,
			"kafka_connect_name": kafkaConnectName,
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
	s.Equal("startup-2", kafkaService.Plan)
	s.Equal("google-europe-west1", kafkaService.CloudName)

	kafkaConnect, err := s.client.Services.Get(ctx, s.config.Project, kafkaConnectName)
	s.NoError(err)
	s.Equal("kafka_connect", kafkaConnect.Type)
	s.Equal("startup-4", kafkaConnect.Plan)
	s.Equal("google-europe-west1", kafkaConnect.CloudName)

	integrations, err := s.client.ServiceIntegrations.List(ctx, s.config.Project, kafkaServiceName)
	s.NoError(err)

	// We don't have integration ID here
	found := 0
	for _, i := range integrations {
		if i.IntegrationType == "kafka_connect" && *i.DestinationService == kafkaConnectName {
			found++
		}
	}
	s.Equal(1, found)
}
