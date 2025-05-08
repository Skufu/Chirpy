package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	apiKeyPrefix := headers.Get("Authorization")
	if apiKeyPrefix == "" {
		return "", errors.New("no API key found")
	}

	if !strings.HasPrefix(apiKeyPrefix, "ApiKey ") {
		return "", errors.New("invalid API key format")
	}

	apiKey := strings.TrimPrefix(apiKeyPrefix, "ApiKey ")
	return apiKey, nil
}
