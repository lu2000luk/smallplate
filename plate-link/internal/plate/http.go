package plate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/redis/go-redis/v9"
)

type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Envelope struct {
	OK    bool `json:"ok"`
	Data  any  `json:"data,omitempty"`
	Error any  `json:"error,omitempty"`
	Meta  any  `json:"meta,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

func NewAPIError(status int, code string, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		Error("failed to encode response:", err)
	}
}

func WriteOK(w http.ResponseWriter, status int, data any) {
	WriteJSON(w, status, Envelope{OK: true, Data: data})
}

func WriteError(w http.ResponseWriter, err error) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		WriteJSON(w, apiErr.Status, Envelope{OK: false, Error: apiErr})
		return
	}
	if errors.Is(err, redis.Nil) {
		WriteJSON(w, http.StatusNotFound, Envelope{OK: false, Error: &APIError{Code: "not_found", Message: "resource not found"}})
		return
	}
	WriteJSON(w, http.StatusInternalServerError, Envelope{OK: false, Error: &APIError{Code: "internal_error", Message: err.Error()}})
}

func WriteCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func HandleCORS(w http.ResponseWriter, r *http.Request) bool {
	WriteCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func DecodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return NewAPIError(http.StatusBadRequest, "invalid_json", "request body is required")
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return NewAPIError(http.StatusBadRequest, "invalid_json", err.Error())
	}
	return nil
}

func Authenticated(deps *Dependencies, next func(http.ResponseWriter, *http.Request, string) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				Error("panic while handling request:", recovered)
				WriteError(w, NewAPIError(http.StatusInternalServerError, "panic", "internal server error"))
			}
		}()

		plateID := strings.TrimSpace(r.PathValue("plateID"))
		if plateID == "" {
			WriteError(w, NewAPIError(http.StatusBadRequest, "missing_plate_id", "plate id is required"))
			return
		}

		authKey := NormalizeAuthorizationHeader(r.Header.Get("Authorization"))
		if authKey == "" {
			WriteError(w, NewAPIError(http.StatusUnauthorized, "missing_authorization", "authorization header is required"))
			return
		}

		authCtx, cancelAuth := context.WithTimeout(r.Context(), deps.Config.RedisOpTimeout)
		defer cancelAuth()
		allowed, err := deps.Manager.Authorize(authCtx, plateID, authKey)
		if err != nil {
			WriteError(w, NewAPIError(http.StatusServiceUnavailable, "manager_unavailable", err.Error()))
			return
		}
		if !allowed {
			WriteError(w, NewAPIError(http.StatusUnauthorized, "invalid_authorization", "authorization denied"))
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), deps.Config.RedisOpTimeout)
		defer cancel()
		if err := next(w, r.WithContext(ctx), plateID); err != nil {
			WriteError(w, err)
		}
	}
}

func NormalizeAuthorizationHeader(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(trimmed), "bearer ") {
		return strings.TrimSpace(trimmed[7:])
	}
	return trimmed
}

func PathValue(r *http.Request, name string) (string, error) {
	raw := strings.TrimSpace(r.PathValue(name))
	if raw == "" {
		return "", NewAPIError(http.StatusBadRequest, "missing_path_value", fmt.Sprintf("missing path value %q", name))
	}
	decoded, err := url.PathUnescape(raw)
	if err != nil {
		return "", NewAPIError(http.StatusBadRequest, "invalid_path_value", err.Error())
	}
	return decoded, nil
}
