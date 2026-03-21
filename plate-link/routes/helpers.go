package routes

import (
	"net/http"
	"strconv"
	"strings"

	"plate/link/internal/plate"
)

func requiredString(value string, name string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", plate.NewAPIError(http.StatusBadRequest, "invalid_request", name+" is required")
	}
	return trimmed, nil
}

func optionalInt64(value *int64, name string) error {
	if value == nil {
		return nil
	}
	if *value < 0 {
		return plate.NewAPIError(http.StatusBadRequest, "invalid_request", name+" must be >= 0")
	}
	return nil
}

func parseTail(r *http.Request) []string {
	raw := strings.TrimSpace(r.PathValue("tail"))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, "/")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func parseInt64String(value string) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}
