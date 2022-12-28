//go:build all || examples

package examples

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/suite"
)

type MysqlTestSuite struct {
	BaseTestSuite
}

func TestMysqlTestSuite(t *testing.T) {
	suite.Run(t, new(MysqlTestSuite))
}

func (s *MysqlTestSuite) TestMysql() {
	s.T().Parallel()

	// Given
	withPrefix := examplesRandPrefix()
	mysqlName := withPrefix("mysql")
	opts := s.withDefaults(&terraform.Options{
		TerraformDir: "../examples/mysql",
		Vars: map[string]interface{}{
			"avn_token":      s.config.Token,
			"avn_project":    s.config.Project,
			"mysql_name":     mysqlName,
			"mysql_username": "username" + uniqueID(),
			"mysql_password": "password" + uniqueID(),
		},
	})

	// When
	defer terraform.Destroy(s.T(), opts)
	terraform.Apply(s.T(), opts)

	// Then
	mysql, err := s.client.Services.Get(s.config.Project, mysqlName)
	s.NoError(err)
	s.Equal("mysql", mysql.Type)
	s.Equal("business-4", mysql.Plan)
	s.Equal("google-europe-west1", mysql.CloudName)
	s.Equal("monday", mysql.MaintenanceWindow.DayOfWeek)
	s.Equal("10:00:00", mysql.MaintenanceWindow.TimeOfDay)
}
