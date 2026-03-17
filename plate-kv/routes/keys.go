// Endpoints contained in this file:
// GET /{plateID}/keys/{key}
// POST /{plateID}/keys/command
// POST /{plateID}/keys/{key}/command
// GET /{plateID}/scan
// POST /{plateID}/scan/hashes/{key}
// POST /{plateID}/scan/sets/{key}
// POST /{plateID}/scan/zsets/{key}
// DELETE /{plateID}/keys/exact/{key}
// POST /{plateID}/keys/delete
// POST /{plateID}/keys/ttl
// DELETE /{plateID}/keys/ttl
// POST /{plateID}/keys/{key}/rename
// POST /{plateID}/keys/{key}/copy
// DELETE /{plateID}/keys/{pattern}
package routes

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"

	"plain/kv/internal/plate"
)

func registerKeys(mux *http.ServeMux, deps *plate.Dependencies) {
	mux.HandleFunc("GET /{plateID}/keys/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		prefixedKey := plate.PrefixKey(plateID, key)
		ctx := r.Context()
		typ, err := deps.Redis.Type(ctx, prefixedKey).Result()
		if err != nil {
			return err
		}
		exists, err := deps.Redis.Exists(ctx, prefixedKey).Result()
		if err != nil {
			return err
		}
		response := map[string]any{
			"key":    key,
			"exists": exists == 1,
			"type":   typ,
		}
		if exists == 1 {
			ttl, err := deps.Redis.PTTL(ctx, prefixedKey).Result()
			if err != nil {
				return err
			}
			response["ttl_ms"] = ttl.Milliseconds()
			if typ == "string" {
				value, err := deps.Redis.Get(ctx, prefixedKey).Result()
				if err == nil {
					response["value"] = value
				}
			}
		}
		plate.WriteOK(w, http.StatusOK, response)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/keys/command", handleCommand(deps, mustCommands(
		"GET", "SET", "MGET", "MSET", "DEL", "UNLINK", "EXISTS", "TYPE", "RENAME", "COPY",
		"EXPIRE", "PEXPIRE", "EXPIREAT", "PEXPIREAT", "TTL", "PTTL", "PERSIST", "GETEX", "GETDEL",
		"SCAN",
	), false))
	mux.HandleFunc("POST /{plateID}/keys/{key}/command", handleCommand(deps, mustCommands(
		"GET", "SET", "DEL", "UNLINK", "EXISTS", "TYPE", "EXPIRE", "PEXPIRE", "EXPIREAT", "PEXPIREAT", "TTL", "PTTL", "PERSIST", "GETEX", "GETDEL",
	), true))
	mux.HandleFunc("GET /{plateID}/scan", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		cursor, err := queryInt64(r, "cursor", 0)
		if err != nil {
			return err
		}
		count, err := queryInt64(r, "count", 100)
		if err != nil {
			return err
		}
		pattern := r.URL.Query().Get("match")
		args := []any{strconv.FormatUint(uint64(cursor), 10), "MATCH", pattern, "COUNT", strconv.FormatInt(count, 10)}
		if keyType := r.URL.Query().Get("type"); keyType != "" {
			args = append(args, "TYPE", keyType)
		}
		result, err := plate.ExecuteCommand(r.Context(), deps, plateID, "SCAN", args...)
		if err != nil {
			return err
		}
		array, _ := result.([]any)
		keys := []string{}
		nextCursor := "0"
		if len(array) == 2 {
			nextCursor, _ = array[0].(string)
			rawKeys, _ := array[1].([]any)
			for _, item := range rawKeys {
				if key, ok := item.(string); ok {
					keys = append(keys, plate.UnprefixKey(plateID, key))
				}
			}
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"cursor": nextCursor, "keys": keys, "done": nextCursor == "0"})
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/scan/hashes/{key}", scanCollectionHandler(deps, "HSCAN"))
	mux.HandleFunc("POST /{plateID}/scan/sets/{key}", scanCollectionHandler(deps, "SSCAN"))
	mux.HandleFunc("POST /{plateID}/scan/zsets/{key}", scanCollectionHandler(deps, "ZSCAN"))
	mux.HandleFunc("DELETE /{plateID}/keys/exact/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "DEL", key)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/keys/delete", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
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
		result, err := execute(r, deps, plateID, "DEL", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/keys/ttl", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key   string `json:"key"`
			TTLMS int64  `json:"ttl_ms"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		if request.TTLMS <= 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "ttl_ms must be greater than zero")
		}
		result, err := execute(r, deps, plateID, "PEXPIRE", key, strconv.FormatInt(request.TTLMS, 10))
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("DELETE /{plateID}/keys/ttl", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key string `json:"key"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "PERSIST", key)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/keys/{key}/rename", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		var request struct {
			Destination string `json:"destination"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		destination, err := requiredString(request.Destination, "destination")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "RENAME", key, destination)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/keys/{key}/copy", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		var request struct {
			Destination string `json:"destination"`
			Replace     bool   `json:"replace"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		destination, err := requiredString(request.Destination, "destination")
		if err != nil {
			return err
		}
		args := []any{key, destination}
		if request.Replace {
			args = append(args, "REPLACE")
		}
		result, err := execute(r, deps, plateID, "COPY", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("DELETE /{plateID}/keys/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		pattern, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		deleted, err := deleteByPattern(r.Context(), deps.Redis, plateID, pattern)
		if err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"deleted": deleted, "pattern": pattern})
		return nil
	}))
}

func scanCollectionHandler(deps *plate.Dependencies, command string) http.HandlerFunc {
	return plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		var request struct {
			Cursor uint64 `json:"cursor"`
			Match  string `json:"match"`
			Count  int64  `json:"count"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		args := []any{key, strconv.FormatUint(request.Cursor, 10)}
		if strings.TrimSpace(request.Match) != "" {
			args = append(args, "MATCH", request.Match)
		}
		if request.Count > 0 {
			args = append(args, "COUNT", strconv.FormatInt(request.Count, 10))
		}
		result, err := plate.ExecuteCommand(r.Context(), deps, plateID, command, args...)
		if err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"result": result})
		return nil
	})
}

func deleteByPattern(ctx context.Context, client *redis.Client, plateID string, pattern string) (int64, error) {
	match := plate.PrefixPattern(plateID, pattern)
	var cursor uint64
	var total int64
	for {
		keys, next, err := client.Scan(ctx, cursor, match, 500).Result()
		if err != nil {
			return 0, err
		}
		if len(keys) > 0 {
			deleted, err := client.Unlink(ctx, keys...).Result()
			if err != nil {
				return 0, err
			}
			total += deleted
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return total, nil
}
