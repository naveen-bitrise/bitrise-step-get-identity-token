package main

import (
	"os"

	"github.com/bitrise-io/go-steputils/v2/export"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-steputils/v2/stepenv"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	. "github.com/bitrise-io/go-utils/v2/exitcode"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/naveen-bitrise/bitrise-step-get-identity-token/step"
)

func main() {
	exitCode := run()
	os.Exit(int(exitCode))
}

func run() ExitCode {
	logger := log.NewLogger()

	fetcher := createTokenFetcher(logger)
	config, err := fetcher.ProcessConfig()
	if err != nil {
		logger.Errorf("Process config: %s", err)
		return Failure
	}

	result, err := fetcher.Run(config)
	if err != nil {
		logger.Errorf("Run: %s", err)
		return Failure
	}

	if err := fetcher.Export(result); err != nil {
		logger.Errorf("Export outputs: %s", err)
		return Failure
	}

	return Success
}

func createTokenFetcher(logger log.Logger) step.TokenFetcher {
	envRepository := stepenv.NewRepository(env.NewRepository())
	inputParser := stepconf.NewInputParser(envRepository)
	exporter := export.NewExporter(command.NewFactory(envRepository))

	return step.NewTokenFetcher(inputParser, envRepository, exporter, logger)
}
