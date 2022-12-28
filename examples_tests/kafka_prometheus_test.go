//go:build all || examples

package examples

import (
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/suite"
)

type KafkaPrometheusTestSuite struct {
	BaseTestSuite
}

func TestKafkaPrometheusTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaPrometheusTestSuite))
}

func (s *KafkaPrometheusTestSuite) TestKafkaPrometheus() {
	s.T().Parallel()

	// Given
	kafkaServiceName := randName("test-examples-kafka-%s")
	prometheusEndpointName := randName("test-prometheus-endpoint-%s")
	opts := s.withDefaults(&terraform.Options{
		TerraformDir: "../examples/kafka_prometheus",
		Vars: map[string]interface{}{
			"avn_token":                s.config.Token,
			"avn_project":              s.config.Project,
			"kafka_name":               kafkaServiceName,
			"prometheus_endpoint_name": prometheusEndpointName,
			"prometheus_username":      randName("username%s"),
			"prometheus_password":      randName("password%s"),
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

	endpoints, err := s.client.ServiceIntegrationEndpoints.List(s.config.Project)
	s.NoError(err)
	foundEndpoints := make([]*aiven.ServiceIntegrationEndpoint, 0)
	for _, e := range endpoints {
		if e.EndpointType == "prometheus" && e.EndpointName == prometheusEndpointName {
			foundEndpoints = append(foundEndpoints, e)
		}
	}
	s.Len(foundEndpoints, 1)

	integrations, err := s.client.ServiceIntegrations.List(s.config.Project, kafkaServiceName)
	s.NoError(err)
	foundIntegrations := 0
	for _, i := range integrations {
		if i.IntegrationType == "prometheus" && *i.DestinationEndpointID == foundEndpoints[0].EndpointID {
			foundIntegrations++
		}
	}
	s.Equal(1, foundIntegrations)
}
