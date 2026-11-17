//go:build all || examples

package examples

import (
	"context"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/suite"
)

type PostgresCRDRTestSuite struct {
	BaseTestSuite
}

func TestPostgresCRDRTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresCRDRTestSuite))
}

func (s *PostgresCRDRTestSuite) TestPostgresCRDR() {
	// Given
	withPrefix := examplesRandPrefix()
	primaryName := withPrefix("pg-primary")
	recoveryName := withPrefix("pg-recovery")
	opts := s.withDefaults(&terraform.Options{
		TerraformDir: "../examples/postgres/postgres_crdr",
		Vars: map[string]interface{}{
			"aiven_token":           s.config.Token,
			"aiven_project_name":    s.config.Project,
			"primary_service_name":  primaryName,
			"recovery_service_name": recoveryName,
		},
	})

	// When
	defer terraform.Destroy(s.T(), opts)
	terraform.Apply(s.T(), opts)

	// Then
	ctx := context.Background()

	primary, err := s.client.Services.Get(ctx, s.config.Project, primaryName)
	s.NoError(err)
	s.Equal("pg", primary.Type)
	s.Equal("startup-4", primary.Plan)
	s.Equal("google-europe-west1", primary.CloudName)

	recovery, err := s.client.Services.Get(ctx, s.config.Project, recoveryName)
	s.NoError(err)
	s.Equal("pg", recovery.Type)
	s.Equal("startup-4", recovery.Plan)
	s.Equal("google-europe-west4", recovery.CloudName)

	integrations, err := s.client.ServiceIntegrations.List(ctx, s.config.Project, primaryName)
	s.NoError(err)

	found := 0
	for _, i := range integrations {
		if i.IntegrationType == "disaster_recovery" && *i.DestinationService == recoveryName {
			found++
		}
	}
	s.Equal(1, found)
}
