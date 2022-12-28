//go:build all || examples

package examples

import (
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
	pgNameEU := withPrefix("pg-eu")
	pgNameUS := withPrefix("pg-us")
	pgNameAS := withPrefix("pg-as")
	opts := s.withDefaults(&terraform.Options{
		TerraformDir: "../examples/postgres",
		Vars: map[string]interface{}{
			"avn_token":        s.config.Token,
			"avn_project":      s.config.Project,
			"postgres_eu_name": pgNameEU,
			"postgres_us_name": pgNameUS,
			"postgres_as_name": pgNameAS,
		},
	})

	// When
	defer terraform.Destroy(s.T(), opts)
	terraform.Apply(s.T(), opts)

	// Then
	pgEU, err := s.client.Services.Get(s.config.Project, pgNameEU)
	s.NoError(err)
	s.Equal("pg", pgEU.Type)
	s.Equal("startup-4", pgEU.Plan)
	s.Equal("aws-eu-west-2", pgEU.CloudName)

	pgUS, err := s.client.Services.Get(s.config.Project, pgNameUS)
	s.NoError(err)
	s.Equal("pg", pgUS.Type)
	s.Equal("business-8", pgUS.Plan)
	s.Equal("do-nyc", pgUS.CloudName)

	pgAS, err := s.client.Services.Get(s.config.Project, pgNameAS)
	s.NoError(err)
	s.Equal("pg", pgAS.Type)
	s.Equal("business-8", pgAS.Plan)
	s.Equal("google-asia-southeast1", pgAS.CloudName)
}
