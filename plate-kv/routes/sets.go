// Endpoints contained in this file:
// POST /{plateID}/sets/command
// POST /{plateID}/sets/{key}/command
// POST /{plateID}/sets/add
// POST /{plateID}/sets/remove
// GET /{plateID}/sets/members/{key}
// POST /{plateID}/sets/contains
// GET /{plateID}/sets/count/{key}
// GET /{plateID}/sets/random/{key}
// POST /{plateID}/sets/pop
// POST /{plateID}/sets/union
// POST /{plateID}/sets/intersect
// POST /{plateID}/sets/diff
package routes

import (
	"net/http"
	"strconv"
	"strings"

	"plain/kv/internal/plate"
)

func registerSets(mux *http.ServeMux, deps *plate.Dependencies) {
	allowed := mustCommands("SADD", "SREM", "SMEMBERS", "SISMEMBER", "SMISMEMBER", "SCARD", "SRANDMEMBER", "SPOP", "SUNION", "SINTER", "SDIFF", "SUNIONSTORE", "SINTERSTORE", "SDIFFSTORE")
	mux.HandleFunc("POST /{plateID}/sets/command", handleCommand(deps, allowed, false))
	mux.HandleFunc("POST /{plateID}/sets/{key}/command", handleCommand(deps, allowed, true))
	mux.HandleFunc("POST /{plateID}/sets/add", setMembersHandler(deps, "SADD"))
	mux.HandleFunc("POST /{plateID}/sets/remove", setMembersHandler(deps, "SREM"))
	mux.HandleFunc("GET /{plateID}/sets/members/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "SMEMBERS", key)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/sets/contains", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key     string   `json:"key"`
			Member  any      `json:"member"`
			Members []string `json:"members"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		if len(request.Members) > 0 {
			if err := requireNonEmptyStrings(request.Members, "members"); err != nil {
				return err
			}
			args := []any{key}
			for _, member := range request.Members {
				args = append(args, member)
			}
			result, err := execute(r, deps, plateID, "SMISMEMBER", args...)
			if err != nil {
				return err
			}
			writeResult(w, result)
			return nil
		}
		member, err := redisValueString(request.Member)
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "SISMEMBER", key, member)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/sets/count/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "SCARD", key)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/sets/random/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		count := r.URL.Query().Get("count")
		args := []any{key}
		if count != "" {
			args = append(args, count)
		}
		result, err := execute(r, deps, plateID, "SRANDMEMBER", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/sets/pop", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key   string `json:"key"`
			Count *int64 `json:"count"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		args := []any{key}
		if request.Count != nil {
			if *request.Count <= 0 {
				return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "count must be greater than zero")
			}
			args = append(args, strconv.FormatInt(*request.Count, 10))
		}
		result, err := execute(r, deps, plateID, "SPOP", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/sets/union", setOperationHandler(deps, "SUNION", "SUNIONSTORE"))
	mux.HandleFunc("POST /{plateID}/sets/intersect", setOperationHandler(deps, "SINTER", "SINTERSTORE"))
	mux.HandleFunc("POST /{plateID}/sets/diff", setOperationHandler(deps, "SDIFF", "SDIFFSTORE"))
}

func setMembersHandler(deps *plate.Dependencies, command string) http.HandlerFunc {
	return plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key     string `json:"key"`
			Members []any  `json:"members"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		if len(request.Members) == 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "members must not be empty")
		}
		members, err := redisValueStrings(request.Members)
		if err != nil {
			return err
		}
		args := append([]any{key}, members...)
		result, err := execute(r, deps, plateID, command, args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	})
}

func setOperationHandler(deps *plate.Dependencies, readCommand string, storeCommand string) http.HandlerFunc {
	return plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Keys        []string `json:"keys"`
			Destination string   `json:"destination"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		if err := requireNonEmptyStrings(request.Keys, "keys"); err != nil {
			return err
		}
		args := make([]any, 0, len(request.Keys)+1)
		if strings.TrimSpace(request.Destination) != "" {
			args = append(args, request.Destination)
		}
		for _, key := range request.Keys {
			args = append(args, key)
		}
		command := readCommand
		if strings.TrimSpace(request.Destination) != "" {
			command = storeCommand
		}
		result, err := execute(r, deps, plateID, command, args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	})
}
