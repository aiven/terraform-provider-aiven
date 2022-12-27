package examples

import (
	"fmt"
	"os"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/suite"
)

type envConfig struct {
	Token        string `envconfig:"AIVEN_TOKEN" required:"true"`
	Project      string `envconfig:"AIVEN_PROJECT_NAME" required:"true"`
	ProviderPath string `envconfig:"AIVEN_PROVIDER_PATH" required:"true"`
}

// BaseTestSuite use for example tests
type BaseTestSuite struct {
	suite.Suite
	config       *envConfig
	client       *aiven.Client
	tfConfigPath string
}

func (s *BaseTestSuite) SetupSuite() {
	err := s.setupSuite()
	if err != nil {
		s.Fail(err.Error())
	}
}

func (s *BaseTestSuite) setupSuite() error {
	s.config = new(envConfig)
	err := envconfig.Process("", s.config)
	if err != nil {
		return err
	}

	// Writes terraform config which forces to use dev provider
	tfConfigPath, err := newTFConfig(s.config.ProviderPath)
	if err != nil {
		return err
	}
	s.tfConfigPath = tfConfigPath

	// Uses client to validates resources
	client, err := newClient(s.config.Token)
	if err != nil {
		return err
	}
	s.client = client
	return nil
}

func (s *BaseTestSuite) TearDownSuite() {
	// Uncomment when fixed https://github.com/stretchr/testify/issues/934
	//_ = os.Remove(s.tfConfigPath)
}

// withDefaults adds default options for terraform test
func (s *BaseTestSuite) withDefaults(opts *terraform.Options) *terraform.Options {
	// No need to use lock file for dev build
	opts.Lock = false

	// Forces to use dev build of provider
	if opts.EnvVars == nil {
		opts.EnvVars = make(map[string]string)
	}
	opts.EnvVars["TF_CLI_CONFIG_FILE"] = s.tfConfigPath
	return terraform.WithDefaultRetryableErrors(s.T(), opts)
}

// randName Returns randomized name for a given format specifier randName("foo-%s")
func randName(format string) string {
	return fmt.Sprintf(format, strings.ToLower(random.UniqueId()))
}

func randNameGen(format string) func() string {
	return func() string {
		return randName(format)
	}
}

func newClient(token string) (*aiven.Client, error) {
	client, err := aiven.NewTokenClient(token, "terraform-provider-aiven/0.11+compatible/dev")
	if err != nil {
		return nil, err
	}
	return client, nil
}

// configTemplate forces terraform use dev build
// https://www.terraform.io/cli/config/config-file#development-overrides-for-provider-developers
const configTemplate = `provider_installation {
  dev_overrides {
    "aiven/aiven" = "%s"
  }
  direct {}
}`

func newTFConfig(providerPath string) (string, error) {
	f, err := os.CreateTemp("", "config-*.tfrc")
	if err != nil {
		return "", err
	}

	c := fmt.Sprintf(configTemplate, providerPath)
	_, err = f.Write([]byte(c))
	if err != nil {
		return "", err
	}

	return f.Name(), nil
}
