// Endpoints contained in this file:
// POST /{plateID}/lists/command
// POST /{plateID}/lists/{key}/command
// POST /{plateID}/lists/left/push
// POST /{plateID}/lists/right/push
// POST /{plateID}/lists/left/pop
// POST /{plateID}/lists/right/pop
// GET /{plateID}/lists/range/{key}
// GET /{plateID}/lists/item/{key}/{index}
// GET /{plateID}/lists/length/{key}
// POST /{plateID}/lists/position
// POST /{plateID}/lists/set
// POST /{plateID}/lists/insert
// POST /{plateID}/lists/remove
// POST /{plateID}/lists/trim
// POST /{plateID}/lists/move
package routes

import (
	"net/http"
	"strconv"
	"strings"

	"plain/kv/internal/plate"
)

func registerLists(mux *http.ServeMux, deps *plate.Dependencies) {
	allowed := mustCommands("LPUSH", "RPUSH", "LPOP", "RPOP", "LLEN", "LRANGE", "LINDEX", "LPOS", "LSET", "LINSERT", "LREM", "LTRIM", "LMOVE")
	mux.HandleFunc("POST /{plateID}/lists/command", handleCommand(deps, allowed, false))
	mux.HandleFunc("POST /{plateID}/lists/{key}/command", handleCommand(deps, allowed, true))
	mux.HandleFunc("POST /{plateID}/lists/left/push", listPushHandler(deps, "LPUSH"))
	mux.HandleFunc("POST /{plateID}/lists/right/push", listPushHandler(deps, "RPUSH"))
	mux.HandleFunc("POST /{plateID}/lists/left/pop", listPopHandler(deps, "LPOP"))
	mux.HandleFunc("POST /{plateID}/lists/right/pop", listPopHandler(deps, "RPOP"))
	mux.HandleFunc("GET /{plateID}/lists/range/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		start, err := queryInt64(r, "start", 0)
		if err != nil {
			return err
		}
		stop, err := queryInt64(r, "stop", -1)
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "LRANGE", key, strconv.FormatInt(start, 10), strconv.FormatInt(stop, 10))
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/lists/item/{key}/{index}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		index, err := plate.PathValue(r, "index")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "LINDEX", key, index)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/lists/length/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "LLEN", key)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/lists/position", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key    string `json:"key"`
			Value  any    `json:"value"`
			Rank   *int64 `json:"rank"`
			Count  *int64 `json:"count"`
			MaxLen *int64 `json:"maxlen"`
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
		if request.Rank != nil {
			args = append(args, "RANK", strconv.FormatInt(*request.Rank, 10))
		}
		if request.Count != nil {
			args = append(args, "COUNT", strconv.FormatInt(*request.Count, 10))
		}
		if request.MaxLen != nil {
			args = append(args, "MAXLEN", strconv.FormatInt(*request.MaxLen, 10))
		}
		result, err := execute(r, deps, plateID, "LPOS", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/lists/set", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key   string `json:"key"`
			Index int64  `json:"index"`
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
		result, err := execute(r, deps, plateID, "LSET", key, strconv.FormatInt(request.Index, 10), value)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/lists/insert", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key   string `json:"key"`
			Pivot any    `json:"pivot"`
			Value any    `json:"value"`
			Where string `json:"where"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		where := strings.ToUpper(strings.TrimSpace(request.Where))
		if where != "BEFORE" && where != "AFTER" {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "where must be before or after")
		}
		pivot, err := redisValueString(request.Pivot)
		if err != nil {
			return err
		}
		value, err := redisValueString(request.Value)
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "LINSERT", key, where, pivot, value)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/lists/remove", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key   string `json:"key"`
			Value any    `json:"value"`
			Count int64  `json:"count"`
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
		result, err := execute(r, deps, plateID, "LREM", key, strconv.FormatInt(request.Count, 10), value)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/lists/trim", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key   string `json:"key"`
			Start int64  `json:"start"`
			Stop  int64  `json:"stop"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "LTRIM", key, strconv.FormatInt(request.Start, 10), strconv.FormatInt(request.Stop, 10))
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/lists/move", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Source      string `json:"source"`
			Destination string `json:"destination"`
			From        string `json:"from"`
			To          string `json:"to"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		source, err := requiredString(request.Source, "source")
		if err != nil {
			return err
		}
		destination, err := requiredString(request.Destination, "destination")
		if err != nil {
			return err
		}
		from := strings.ToUpper(strings.TrimSpace(request.From))
		to := strings.ToUpper(strings.TrimSpace(request.To))
		if (from != "LEFT" && from != "RIGHT") || (to != "LEFT" && to != "RIGHT") {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "from and to must be left or right")
		}
		result, err := execute(r, deps, plateID, "LMOVE", source, destination, from, to)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
}

func listPushHandler(deps *plate.Dependencies, command string) http.HandlerFunc {
	return plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key    string `json:"key"`
			Values []any  `json:"values"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		if len(request.Values) == 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "values must not be empty")
		}
		values, err := redisValueStrings(request.Values)
		if err != nil {
			return err
		}
		args := append([]any{key}, values...)
		result, err := execute(r, deps, plateID, command, args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	})
}

func listPopHandler(deps *plate.Dependencies, command string) http.HandlerFunc {
	return plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
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
		result, err := execute(r, deps, plateID, command, args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	})
}
