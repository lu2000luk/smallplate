package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
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
