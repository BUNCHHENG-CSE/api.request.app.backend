package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const standardIntrospectionQuery = `
query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    subscriptionType { name }
    types {
      kind
      name
      description
      fields(includeDeprecated: true) {
        name
        description
        args {
          name
          description
          type {
            kind
            name
            ofType { kind name }
          }
        }
        type {
          kind
          name
          ofType { kind name }
        }
      }
    }
  }
}
`

type GraphQLService interface {
	Introspect(ctx context.Context, endpointURL string, headers map[string]string) (map[string]any, error)
}

type graphQLService struct {
	httpClient *http.Client
}

func NewGraphQLService() GraphQLService {
	return &graphQLService{
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *graphQLService) Introspect(ctx context.Context, endpointURL string, headers map[string]string) (map[string]any, error) {
	if endpointURL == "" {
		return nil, errors.New("GraphQL endpoint URL is required")
	}

	payload := map[string]any{
		"query": standardIntrospectionQuery,
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach GraphQL endpoint: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read GraphQL response: %w", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return nil, errors.New("GraphQL endpoint returned non-JSON response")
	}

	if errorsVal, hasErrors := parsed["errors"]; hasErrors && errorsVal != nil {
		return nil, fmt.Errorf("GraphQL introspection returned errors: %v", errorsVal)
	}

	dataVal, hasData := parsed["data"].(map[string]any)
	if !hasData || dataVal == nil {
		return nil, errors.New("GraphQL response contains no schema data")
	}

	return dataVal, nil
}
