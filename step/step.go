package step

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bitrise-io/go-steputils/v2/export"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/naveen-bitrise/bitrise-step-get-identity-token/api"
)

const identityTokenKey = "BITRISE_IDENTITY_TOKEN"

type TokenFetcher struct {
	inputParser   stepconf.InputParser
	envRepository env.Repository
	exporter      export.Exporter
	logger        log.Logger
}

func NewTokenFetcher(inputParser stepconf.InputParser, envRepository env.Repository, exporter export.Exporter, logger log.Logger) TokenFetcher {
	return TokenFetcher{
		inputParser:   inputParser,
		envRepository: envRepository,
		exporter:      exporter,
		logger:        logger,
	}
}

func (r TokenFetcher) ProcessConfig() (Config, error) {
	var input Input
	err := r.inputParser.Parse(&input)
	if err != nil {
		return Config{}, err
	}

	stepconf.Print(input)
	r.logger.Println()
	r.logger.EnableDebugLog(input.Verbose)

	return Config{
		BuildURL:   input.BuildURL,
		BuildToken: input.BuildToken,
		Audience:   input.Audience,
	}, nil
}

func (r TokenFetcher) Run(config Config) (Result, error) {
	client := api.NewDefaultAPIClient(config.BuildURL, config.BuildToken, r.logger)

	parameter := api.GetIdentityTokenParameter{
		Audience: config.Audience,
	}
	response, err := client.GetIdentityToken(parameter)
	if err != nil {
		return Result{}, err
	}

	r.logger.Donef("Identity token fetched.")

	// Decode and log token claims for debugging
	if claims, err := decodeJWTPayload(response.Token); err == nil {
		r.logger.Debugf("Token claims:")
		r.logger.Debugf("  workflow: %v", claims["workflow"])
		r.logger.Debugf("  app_slug: %v", claims["app_slug"])
		r.logger.Debugf("  audience: %v", claims["aud"])
		r.logger.Debugf("  subject: %v", claims["sub"])
		r.logger.Debugf("  issuer: %v", claims["iss"])
	} else {
		r.logger.Warnf("Failed to decode token claims: %s", err)
	}

	return Result{
		IdentityToken: response.Token,
	}, nil
}

func (r TokenFetcher) Export(result Result) error {
	r.logger.Printf("The following outputs are exported as environment variables:")

	values := map[string]string{
		identityTokenKey: result.IdentityToken,
	}

	for key, value := range values {
		err := r.exporter.ExportOutput(key, value)
		if err != nil {
			return err
		}

		r.logger.Donef("$%s = %s", key, value)
	}

	return nil
}

// decodeJWTPayload decodes the payload section of a JWT token without verification
func decodeJWTPayload(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT token format: expected 3 parts, got %d", len(parts))
	}

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	// Parse JSON
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	return claims, nil
}
