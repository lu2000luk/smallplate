// Endpoints contained in this file:
// POST /{plateID}/strings/command
// POST /{plateID}/strings/{key}/command
// GET /{plateID}/strings/get/{key}
// POST /{plateID}/strings/get
// POST /{plateID}/strings/set
// POST /{plateID}/strings/set-many
// POST /{plateID}/strings/increment
// POST /{plateID}/strings/decrement
// POST /{plateID}/strings/append
// GET /{plateID}/strings/length/{key}
// GET /{plateID}/strings/range/{key}
// POST /{plateID}/strings/range/set
// POST /{plateID}/strings/get-and-expire
// DELETE /{plateID}/strings/get-and-delete/{key}
package routes

import (
	"net/http"
	"strconv"
	"strings"

	"plain/kv/internal/plate"
)

func registerStrings(mux *http.ServeMux, deps *plate.Dependencies) {
	allowed := mustCommands("GET", "SET", "MGET", "MSET", "GETEX", "GETDEL", "INCR", "DECR", "INCRBY", "DECRBY", "INCRBYFLOAT", "APPEND", "STRLEN", "SETRANGE", "GETRANGE")
	mux.HandleFunc("POST /{plateID}/strings/command", handleCommand(deps, allowed, false))
	mux.HandleFunc("POST /{plateID}/strings/{key}/command", handleCommand(deps, allowed, true))
	mux.HandleFunc("GET /{plateID}/strings/get/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "GET", key)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/strings/get", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Keys []string `json:"keys"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		if err := requireNonEmptyStrings(request.Keys, "keys"); err != nil {
			return err
		}
		args := make([]any, 0, len(request.Keys))
		for _, key := range request.Keys {
			args = append(args, key)
		}
		result, err := execute(r, deps, plateID, "MGET", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/strings/set", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key     string `json:"key"`
			Value   any    `json:"value"`
			TTLMS   *int64 `json:"ttl_ms"`
			NX      bool   `json:"nx"`
			XX      bool   `json:"xx"`
			KeepTTL bool   `json:"keep_ttl"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		value, err := redisValueString(request.Value)
		if err != nil {
			return err
		}
		args := []any{key, value}
		if request.TTLMS != nil {
			if *request.TTLMS < 0 {
				return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "ttl_ms must be zero or greater")
			}
			if *request.TTLMS > 0 {
				args = append(args, "PX", strconv.FormatInt(*request.TTLMS, 10))
			}
		}
		if request.KeepTTL {
			args = append(args, "KEEPTTL")
		}
		if request.NX && request.XX {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "nx and xx cannot both be true")
		}
		if request.NX {
			args = append(args, "NX")
		}
		if request.XX {
			args = append(args, "XX")
		}
		result, err := execute(r, deps, plateID, "SET", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/strings/set-many", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Values map[string]any `json:"values"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		if len(request.Values) == 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "values must contain at least one key")
		}
		pairs, err := stringMapPairs(request.Values)
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "MSET", pairs...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/strings/increment", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key    string `json:"key"`
			Amount any    `json:"amount"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		command := "INCR"
		args := []any{key}
		if request.Amount != nil {
			amount, err := redisValueString(request.Amount)
			if err != nil {
				return err
			}
			if looksFloat(amount) {
				command = "INCRBYFLOAT"
			} else {
				command = "INCRBY"
			}
			args = append(args, amount)
		}
		result, err := execute(r, deps, plateID, command, args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/strings/decrement", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key    string `json:"key"`
			Amount any    `json:"amount"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		command := "DECR"
		args := []any{key}
		if request.Amount != nil {
			amount, err := redisValueString(request.Amount)
			if err != nil {
				return err
			}
			if looksFloat(amount) {
				return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "decrement only supports integer amounts")
			}
			command = "DECRBY"
			args = append(args, amount)
		}
		result, err := execute(r, deps, plateID, command, args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/strings/append", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key   string `json:"key"`
			Value any    `json:"value"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		value, err := redisValueString(request.Value)
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "APPEND", key, value)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/strings/length/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "STRLEN", key)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/strings/range/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		start, err := queryInt64(r, "start", 0)
		if err != nil {
			return err
		}
		end, err := queryInt64(r, "end", -1)
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "GETRANGE", key, strconv.FormatInt(start, 10), strconv.FormatInt(end, 10))
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/strings/range/set", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key    string `json:"key"`
			Offset int64  `json:"offset"`
			Value  any    `json:"value"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		if request.Offset < 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "offset must be zero or greater")
		}
		value, err := redisValueString(request.Value)
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "SETRANGE", key, strconv.FormatInt(request.Offset, 10), value)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/strings/get-and-expire", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key   string `json:"key"`
			TTLMS *int64 `json:"ttl_ms"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		args := []any{key}
		if request.TTLMS != nil {
			if *request.TTLMS < 0 {
				return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "ttl_ms must be zero or greater")
			}
			if *request.TTLMS == 0 {
				args = append(args, "PERSIST")
			} else {
				args = append(args, "PX", strconv.FormatInt(*request.TTLMS, 10))
			}
		}
		result, err := execute(r, deps, plateID, "GETEX", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("DELETE /{plateID}/strings/get-and-delete/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "GETDEL", key)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
}

func looksFloat(value string) bool {
	return strings.ContainsAny(value, ".eE")
}
