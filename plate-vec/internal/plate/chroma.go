package plate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ChromaClient struct {
	baseURL string
	client  *http.Client
}

func NewChromaClient(baseURL string, timeout time.Duration) *ChromaClient {
	return &ChromaClient{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *ChromaClient) Heartbeat(ctx context.Context) error {
	_, err := c.Request(ctx, http.MethodGet, "/api/v2/heartbeat", nil)
	return err
}

func (c *ChromaClient) Request(ctx context.Context, method string, path string, body any) (map[string]any, error) {
	value, err := c.RequestAny(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return map[string]any{}, nil
	}
	if mapped, ok := value.(map[string]any); ok {
		return mapped, nil
	}
	return map[string]any{"value": value}, nil
}

func (c *ChromaClient) RequestAny(ctx context.Context, method string, path string, body any) (any, error) {
	if c.baseURL == "" {
		return nil, NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", "chroma url is not configured")
	}
	parsed, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", err.Error())
	}

	var payload io.Reader
	if body != nil {
		encoded, encErr := json.Marshal(body)
		if encErr != nil {
			return nil, NewAPIError(http.StatusBadRequest, "invalid_request", encErr.Error())
		}
		payload = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, parsed.String(), payload)
	if err != nil {
		return nil, NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", err.Error())
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", err.Error())
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", err.Error())
	}

	if resp.StatusCode >= 400 {
		return nil, mapChromaError(resp.StatusCode, content)
	}

	if len(content) == 0 {
		return map[string]any{}, nil
	}

	var result any
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", "invalid chroma response")
	}
	return result, nil
}

func mapChromaError(status int, payload []byte) error {
	message := strings.TrimSpace(string(payload))
	if message == "" {
		message = http.StatusText(status)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err == nil {
		if msg, ok := decoded["message"].(string); ok && strings.TrimSpace(msg) != "" {
			message = msg
		}
		if errName, ok := decoded["error"].(string); ok {
			errName = strings.ToLower(strings.TrimSpace(errName))
			switch {
			case strings.Contains(errName, "not_found"):
				return NewAPIError(http.StatusNotFound, "not_found", message)
			case strings.Contains(errName, "invalid"):
				return NewAPIError(http.StatusBadRequest, "invalid_request", message)
			}
		}
	}

	switch status {
	case http.StatusNotFound:
		return NewAPIError(http.StatusNotFound, "not_found", message)
	case http.StatusConflict:
		return NewAPIError(http.StatusConflict, "conflict", message)
	case http.StatusBadRequest:
		return NewAPIError(http.StatusBadRequest, "invalid_request", message)
	case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
		return NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", message)
	default:
		if status >= 500 {
			return NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", message)
		}
		return NewAPIError(http.StatusBadRequest, "invalid_request", message)
	}
}

func EnsureMap(value any) (map[string]any, error) {
	if value == nil {
		return map[string]any{}, nil
	}
	mapped, ok := value.(map[string]any)
	if ok {
		return mapped, nil
	}
	return nil, errors.New("invalid map payload")
}

func AsString(value any) string {
	s, _ := value.(string)
	return strings.TrimSpace(s)
}

func AsInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}

func ExpectKeys(source map[string]any, keys ...string) map[string]any {
	result := make(map[string]any, len(keys))
	for _, key := range keys {
		if value, ok := source[key]; ok {
			result[key] = value
		}
	}
	return result
}

func ValidateProvider(provider string) error {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case string(ProviderOpenAI), string(ProviderOpenRouter):
		return nil
	case "":
		return nil
	default:
		return NewAPIError(http.StatusBadRequest, "embedding_provider_unsupported", fmt.Sprintf("unsupported provider %q", provider))
	}
}
