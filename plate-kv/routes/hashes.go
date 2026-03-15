// Endpoints contained in this file:
// POST /{plateID}/hashes/command
// POST /{plateID}/hashes/{key}/command
// POST /{plateID}/hashes/set
// GET /{plateID}/hashes/get/{key}
// GET /{plateID}/hashes/get/{key}/{field}
// POST /{plateID}/hashes/get
// DELETE /{plateID}/hashes/delete/{key}
// DELETE /{plateID}/hashes/delete/{key}/{field}
// POST /{plateID}/hashes/delete
// GET /{plateID}/hashes/exists/{key}/{field}
// GET /{plateID}/hashes/keys/{key}
// GET /{plateID}/hashes/values/{key}
// GET /{plateID}/hashes/length/{key}
// POST /{plateID}/hashes/increment
// POST /{plateID}/hashes/set-if-absent
// GET /{plateID}/hashes/random/{key}
package routes

import (
	"net/http"
	"strconv"

	"plain/kv/internal/plate"
)

func registerHashes(mux *http.ServeMux, deps *plate.Dependencies) {
	allowed := mustCommands("HSET", "HGET", "HMGET", "HGETALL", "HDEL", "HEXISTS", "HINCRBY", "HINCRBYFLOAT", "HKEYS", "HVALS", "HLEN", "HSETNX", "HRANDFIELD")
	mux.HandleFunc("POST /{plateID}/hashes/command", handleCommand(deps, allowed, false))
	mux.HandleFunc("POST /{plateID}/hashes/{key}/command", handleCommand(deps, allowed, true))
	mux.HandleFunc("POST /{plateID}/hashes/set", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key   string `json:"key"`
			Field string `json:"field"`
			Value any    `json:"value"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		args := []any{key}
		if request.Field != "" {
			field, err := requiredString(request.Field, "field")
			if err != nil {
				return err
			}
			value, err := redisValueString(request.Value)
			if err != nil {
				return err
			}
			args = append(args, field, value)
		} else {
			pairs, err := hashPairsFromValue(request.Value)
			if err != nil {
				return err
			}
			args = append(args, pairs...)
		}
		result, err := execute(r, deps, plateID, "HSET", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/hashes/get/{key}/{field}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		field, err := plate.PathValue(r, "field")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "HGET", key, field)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/hashes/get/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "HGETALL", key)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/hashes/get", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key    string   `json:"key"`
			Fields []string `json:"fields"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		if err := requireNonEmptyStrings(request.Fields, "fields"); err != nil {
			return err
		}
		args := []any{key}
		for _, field := range request.Fields {
			args = append(args, field)
		}
		result, err := execute(r, deps, plateID, "HMGET", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("DELETE /{plateID}/hashes/delete/{key}/{field}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		field, err := plate.PathValue(r, "field")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "HDEL", key, field)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("DELETE /{plateID}/hashes/delete/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		if plate.QueryBool(r, "allow_delete_key") {
			result, err := execute(r, deps, plateID, "DEL", key)
			if err != nil {
				return err
			}
			writeResult(w, result)
			return nil
		}
		return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "field is required; use allow_delete_key=true to delete the whole key")
	}))
	mux.HandleFunc("POST /{plateID}/hashes/delete", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key    string   `json:"key"`
			Fields []string `json:"fields"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		if err := requireNonEmptyStrings(request.Fields, "fields"); err != nil {
			return err
		}
		args := []any{key}
		for _, field := range request.Fields {
			args = append(args, field)
		}
		result, err := execute(r, deps, plateID, "HDEL", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/hashes/exists/{key}/{field}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		field, err := plate.PathValue(r, "field")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "HEXISTS", key, field)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/hashes/keys/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "HKEYS", key)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/hashes/values/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "HVALS", key)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/hashes/length/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "HLEN", key)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/hashes/increment", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key    string `json:"key"`
			Field  string `json:"field"`
			Amount any    `json:"amount"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		field, err := requiredString(request.Field, "field")
		if err != nil {
			return err
		}
		amount, err := redisValueString(request.Amount)
		if err != nil {
			return err
		}
		command := "HINCRBY"
		if looksFloat(amount) {
			command = "HINCRBYFLOAT"
		}
		result, err := execute(r, deps, plateID, command, key, field, amount)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/hashes/set-if-absent", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key   string `json:"key"`
			Field string `json:"field"`
			Value any    `json:"value"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		field, err := requiredString(request.Field, "field")
		if err != nil {
			return err
		}
		value, err := redisValueString(request.Value)
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "HSETNX", key, field, value)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/hashes/random/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		count, err := queryInt64(r, "count", 1)
		if err != nil {
			return err
		}
		args := []any{key, strconv.FormatInt(count, 10)}
		if plate.QueryBool(r, "with_values") {
			args = append(args, "WITHVALUES")
		}
		result, err := execute(r, deps, plateID, "HRANDFIELD", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
}
