//go:build all || examples

package examples

import (
	"context"
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
	// Given
	withPrefix := examplesRandPrefix()
	mysqlName := withPrefix("mysql")
	opts := s.withDefaults(&terraform.Options{
		TerraformDir: "../examples/mysql",
		Vars: map[string]interface{}{
			"aiven_token":        s.config.Token,
			"aiven_project_name": s.config.Project,
			"mysql_name":         mysqlName,
			"mysql_username":     "username" + uniqueID(),
			"mysql_password":     "password" + uniqueID(),
		},
	})

	// When
	defer terraform.Destroy(s.T(), opts)
	terraform.Apply(s.T(), opts)

	// Then
	ctx := context.Background()

	mysql, err := s.client.Services.Get(ctx, s.config.Project, mysqlName)
	s.NoError(err)
	s.Equal("mysql", mysql.Type)
	s.Equal("business-4", mysql.Plan)
	s.Equal("google-europe-west1", mysql.CloudName)
	s.Equal("monday", mysql.MaintenanceWindow.DayOfWeek)
	s.Equal("10:00:00", mysql.MaintenanceWindow.TimeOfDay)
}
