//go:build all || examples

package examples

import (
	"context"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/suite"
)

type PostgresTestSuite struct {
	BaseTestSuite
}

func TestPostgresTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}

func (s *PostgresTestSuite) TestPostgres() {
	s.T().Parallel()

	// Given
	withPrefix := examplesRandPrefix()
	pgName := withPrefix("pg")
	opts := s.withDefaults(&terraform.Options{
		TerraformDir: "../examples/postgres",
		Vars: map[string]interface{}{
			"aiven_token":           s.config.Token,
			"aiven_project_name":    s.config.Project,
			"postgres_service_name": pgName,
		},
	})

	// When
	defer terraform.Destroy(s.T(), opts)
	terraform.Apply(s.T(), opts)

	// Then
	ctx := context.Background()

	pg, err := s.client.Services.Get(ctx, s.config.Project, pgName)
	s.NoError(err)
	s.Equal("pg", pg.Type)
	s.Equal("startup-4", pg.Plan)
	s.Equal("aws-eu-west-2", pg.CloudName)
}
