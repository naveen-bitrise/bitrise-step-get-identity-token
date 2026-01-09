package step

import (
	"testing"

	"github.com/bitrise-io/go-steputils/v2/export"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/naveen-bitrise/bitrise-step-get-identity-token/step/mocks"
	"github.com/stretchr/testify/assert"
)

func TestConfigParsing(t *testing.T) {
	config := Config{
		BuildURL:   "build-url",
		BuildToken: stepconf.Secret("build-token"),
		Audience:   "audience",
	}

	mockEnvRepository := mocks.NewRepository(t)
	mockEnvRepository.On("Get", "build_url").Return(config.BuildURL)
	mockEnvRepository.On("Get", "build_api_token").Return(string(config.BuildToken))
	mockEnvRepository.On("Get", "audience").Return(config.Audience)
	mockEnvRepository.On("Get", "verbose").Return("false")

	inputParser := stepconf.NewInputParser(mockEnvRepository)
	exporter := export.NewExporter(mocks.NewFactory(t))
	sut := NewTokenFetcher(inputParser, mockEnvRepository, exporter, log.NewLogger())

	receivedConfig, err := sut.ProcessConfig()
	assert.NoError(t, err)
	assert.Equal(t, config, receivedConfig)

	mockEnvRepository.AssertExpectations(t)
}

func TestExport(t *testing.T) {
	result := Result{
		IdentityToken: "token",
	}

	cmd := testCommand()
	mockFactory := mocks.NewFactory(t)
	mockFactory.On("Create", "envman", mockParameters("BITRISE_IDENTITY_TOKEN", result.IdentityToken), (*command.Opts)(nil)).Return(cmd)

	mockEnvRepository := mocks.NewRepository(t)
	inputParser := stepconf.NewInputParser(mockEnvRepository)
	exporter := export.NewExporter(mockFactory)
	sut := NewTokenFetcher(inputParser, mockEnvRepository, exporter, log.NewLogger())

	err := sut.Export(result)
	assert.NoError(t, err)

	mockEnvRepository.AssertExpectations(t)
}

func testCommand() command.Command {
	factory := command.NewFactory(env.NewRepository())
	return factory.Create("pwd", []string{}, nil)
}

func mockParameters(key, value string) []string {
	return []string{"add", "--key", key, "--value", value}
}
