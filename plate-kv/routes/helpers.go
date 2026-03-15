package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"plain/kv/internal/plate"
)

type commandRequest struct {
	Command string `json:"command"`
	Args    []any  `json:"args"`
}

func handleCommand(deps *plate.Dependencies, allowed map[string]struct{}, prependKey bool) http.HandlerFunc {
	return plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request commandRequest
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		command := strings.ToUpper(strings.TrimSpace(request.Command))
		if _, ok := allowed[command]; !ok {
			return plate.NewAPIError(http.StatusBadRequest, "unsupported_command", fmt.Sprintf("command %s is not allowed on this endpoint", command))
		}
		args := request.Args
		if prependKey {
			key, err := plate.PathValue(r, "key")
			if err != nil {
				return err
			}
			args = append([]any{key}, args...)
		}
		result, err := plate.ExecuteCommand(r.Context(), deps, plateID, command, args...)
		if err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"result": result})
		return nil
	})
}

func parseCommandMatrix(r *http.Request) ([][]any, error) {
	if r.Body == nil {
		return nil, plate.NewAPIError(http.StatusBadRequest, "invalid_json", "request body is required")
	}
	var raw [][]any
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return nil, plate.NewAPIError(http.StatusBadRequest, "invalid_json", err.Error())
	}
	return raw, nil
}

func mustCommands(commands ...string) map[string]struct{} {
	allowed := make(map[string]struct{}, len(commands))
	for _, command := range commands {
		allowed[strings.ToUpper(command)] = struct{}{}
	}
	return allowed
}

func queryInt64(r *http.Request, key string, fallback int64) (int64, error) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, plate.NewAPIError(http.StatusBadRequest, "invalid_query", fmt.Sprintf("invalid %s", key))
	}
	return parsed, nil
}

func queryInt(r *http.Request, key string, fallback int) (int, error) {
	value, err := queryInt64(r, key, int64(fallback))
	if err != nil {
		return 0, err
	}
	return int(value), nil
}

func execute(r *http.Request, deps *plate.Dependencies, plateID string, command string, args ...any) (any, error) {
	return plate.ExecuteCommand(r.Context(), deps, plateID, command, args...)
}

func writeResult(w http.ResponseWriter, result any) {
	plate.WriteOK(w, http.StatusOK, map[string]any{"result": result})
}

func requiredString(value string, name string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", plate.NewAPIError(http.StatusBadRequest, "invalid_request", fmt.Sprintf("%s is required", name))
	}
	return trimmed, nil
}

func requireNonEmptyStrings(values []string, name string) error {
	if len(values) == 0 {
		return plate.NewAPIError(http.StatusBadRequest, "invalid_request", fmt.Sprintf("%s must not be empty", name))
	}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", fmt.Sprintf("%s must not contain empty values", name))
		}
	}
	return nil
}

func queryString(r *http.Request, key string, fallback string) string {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return fallback
	}
	return value
}

func queryRequiredString(r *http.Request, key string) (string, error) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return "", plate.NewAPIError(http.StatusBadRequest, "invalid_query", fmt.Sprintf("%s is required", key))
	}
	return value, nil
}

func queryOptionalBool(r *http.Request, key string) *bool {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return nil
	}
	parsed := plate.QueryBool(r, key)
	return &parsed
}

func redisValueString(value any) (string, error) {
	switch typed := value.(type) {
	case nil:
		return "", plate.NewAPIError(http.StatusBadRequest, "invalid_request", "value is required")
	case string:
		return typed, nil
	case json.Number:
		return typed.String(), nil
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32), nil
	case int:
		return strconv.Itoa(typed), nil
	case int8:
		return strconv.FormatInt(int64(typed), 10), nil
	case int16:
		return strconv.FormatInt(int64(typed), 10), nil
	case int32:
		return strconv.FormatInt(int64(typed), 10), nil
	case int64:
		return strconv.FormatInt(typed, 10), nil
	case uint:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint64:
		return strconv.FormatUint(typed, 10), nil
	case bool:
		if typed {
			return "true", nil
		}
		return "false", nil
	default:
		payload, err := json.Marshal(typed)
		if err != nil {
			return "", plate.NewAPIError(http.StatusBadRequest, "invalid_request", fmt.Sprintf("failed to encode value: %v", err))
		}
		return string(payload), nil
	}
}

func redisValueStrings(values []any) ([]any, error) {
	args := make([]any, 0, len(values))
	for _, value := range values {
		stringValue, err := redisValueString(value)
		if err != nil {
			return nil, err
		}
		args = append(args, stringValue)
	}
	return args, nil
}

func stringMapPairs(values map[string]any) ([]any, error) {
	if len(values) == 0 {
		return nil, plate.NewAPIError(http.StatusBadRequest, "invalid_request", "value must contain at least one field")
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	pairs := make([]any, 0, len(values)*2)
	for _, key := range keys {
		value, err := redisValueString(values[key])
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, key, value)
	}
	return pairs, nil
}

func hashPairsFromValue(value any) ([]any, error) {
	switch typed := value.(type) {
	case map[string]any:
		return stringMapPairs(typed)
	case string:
		trimmed := strings.TrimSpace(typed)
		if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
			var object map[string]any
			if err := json.Unmarshal([]byte(trimmed), &object); err != nil {
				return nil, plate.NewAPIError(http.StatusBadRequest, "invalid_request", "value must be a valid JSON object")
			}
			return stringMapPairs(object)
		}
	}
	return nil, plate.NewAPIError(http.StatusBadRequest, "invalid_request", "value must be an object or a stringified JSON object when field is omitted")
}
