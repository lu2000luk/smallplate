// Endpoints contained in this file:
// POST /{plateID}/streams/command
// POST /{plateID}/streams/{key}/command
// POST /{plateID}/streams/add
// GET /{plateID}/streams/length/{key}
// GET /{plateID}/streams/range/{key}
// GET /{plateID}/streams/reverse/{key}
// POST /{plateID}/streams/read
// POST /{plateID}/streams/trim
// POST /{plateID}/streams/delete
// GET /{plateID}/streams/info/{key}
// POST /{plateID}/streams/groups/create
// POST /{plateID}/streams/groups/read
// POST /{plateID}/streams/groups/ack
// GET /{plateID}/streams/groups/pending/{key}/{group}
// POST /{plateID}/streams/groups/claim
// POST /{plateID}/streams/groups/autoclaim
package routes

import (
	"net/http"
	"strconv"
	"strings"

	"plain/kv/internal/plate"
)

func registerStreams(mux *http.ServeMux, deps *plate.Dependencies) {
	allowed := mustCommands("XADD", "XLEN", "XRANGE", "XREVRANGE", "XREAD", "XTRIM", "XDEL", "XINFO", "XGROUP", "XREADGROUP", "XACK", "XPENDING", "XCLAIM", "XAUTOCLAIM")
	mux.HandleFunc("POST /{plateID}/streams/command", handleCommand(deps, allowed, false))
	mux.HandleFunc("POST /{plateID}/streams/{key}/command", handleCommand(deps, allowed, true))
	mux.HandleFunc("POST /{plateID}/streams/add", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key    string         `json:"key"`
			ID     string         `json:"id"`
			Values map[string]any `json:"values"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		pairs, err := stringMapPairs(request.Values)
		if err != nil {
			return err
		}
		entryID := strings.TrimSpace(request.ID)
		if entryID == "" {
			entryID = "*"
		}
		args := append([]any{key, entryID}, pairs...)
		result, err := execute(r, deps, plateID, "XADD", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/streams/length/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "XLEN", key)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/streams/range/{key}", streamRangeHandler(deps, "XRANGE"))
	mux.HandleFunc("GET /{plateID}/streams/reverse/{key}", streamRangeHandler(deps, "XREVRANGE"))
	mux.HandleFunc("POST /{plateID}/streams/read", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Streams []struct {
				Key string `json:"key"`
				ID  string `json:"id"`
			} `json:"streams"`
			Count   *int64 `json:"count"`
			BlockMS *int64 `json:"block_ms"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		if len(request.Streams) == 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "streams must not be empty")
		}
		args := []any{}
		if request.Count != nil {
			args = append(args, "COUNT", strconv.FormatInt(*request.Count, 10))
		}
		if request.BlockMS != nil {
			args = append(args, "BLOCK", strconv.FormatInt(*request.BlockMS, 10))
		}
		args = append(args, "STREAMS")
		for _, stream := range request.Streams {
			key, err := requiredString(stream.Key, "stream key")
			if err != nil {
				return err
			}
			args = append(args, key)
		}
		for _, stream := range request.Streams {
			id := strings.TrimSpace(stream.ID)
			if id == "" {
				id = "$"
			}
			args = append(args, id)
		}
		result, err := execute(r, deps, plateID, "XREAD", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/streams/trim", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key         string `json:"key"`
			Strategy    string `json:"strategy"`
			Threshold   any    `json:"threshold"`
			Approximate bool   `json:"approximate"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		strategy := strings.ToUpper(strings.TrimSpace(request.Strategy))
		if strategy != "MAXLEN" && strategy != "MINID" {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "strategy must be maxlen or minid")
		}
		threshold, err := redisValueString(request.Threshold)
		if err != nil {
			return err
		}
		args := []any{key, strategy}
		if request.Approximate {
			args = append(args, "~")
		}
		args = append(args, threshold)
		result, err := execute(r, deps, plateID, "XTRIM", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/streams/delete", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key string   `json:"key"`
			IDs []string `json:"ids"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		if err := requireNonEmptyStrings(request.IDs, "ids"); err != nil {
			return err
		}
		args := []any{key}
		for _, id := range request.IDs {
			args = append(args, id)
		}
		result, err := execute(r, deps, plateID, "XDEL", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/streams/info/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		section := strings.ToUpper(queryString(r, "section", "STREAM"))
		if section != "STREAM" && section != "GROUPS" && section != "CONSUMERS" {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_query", "section must be stream, groups, or consumers")
		}
		args := []any{section, key}
		if section == "CONSUMERS" {
			group, err := queryRequiredString(r, "group")
			if err != nil {
				return err
			}
			args = append(args, group)
		}
		result, err := execute(r, deps, plateID, "XINFO", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/streams/groups/create", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key      string `json:"key"`
			Group    string `json:"group"`
			ID       string `json:"id"`
			MKStream bool   `json:"mkstream"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		group, err := requiredString(request.Group, "group")
		if err != nil {
			return err
		}
		streamID := strings.TrimSpace(request.ID)
		if streamID == "" {
			streamID = "$"
		}
		args := []any{"CREATE", key, group, streamID}
		if request.MKStream {
			args = append(args, "MKSTREAM")
		}
		result, err := execute(r, deps, plateID, "XGROUP", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/streams/groups/read", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Group    string `json:"group"`
			Consumer string `json:"consumer"`
			Streams  []struct {
				Key string `json:"key"`
				ID  string `json:"id"`
			} `json:"streams"`
			Count   *int64 `json:"count"`
			BlockMS *int64 `json:"block_ms"`
			NoAck   bool   `json:"noack"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		group, err := requiredString(request.Group, "group")
		if err != nil {
			return err
		}
		consumer, err := requiredString(request.Consumer, "consumer")
		if err != nil {
			return err
		}
		if len(request.Streams) == 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "streams must not be empty")
		}
		args := []any{"GROUP", group, consumer}
		if request.Count != nil {
			args = append(args, "COUNT", strconv.FormatInt(*request.Count, 10))
		}
		if request.BlockMS != nil {
			args = append(args, "BLOCK", strconv.FormatInt(*request.BlockMS, 10))
		}
		if request.NoAck {
			args = append(args, "NOACK")
		}
		args = append(args, "STREAMS")
		for _, stream := range request.Streams {
			key, err := requiredString(stream.Key, "stream key")
			if err != nil {
				return err
			}
			args = append(args, key)
		}
		for _, stream := range request.Streams {
			id := strings.TrimSpace(stream.ID)
			if id == "" {
				id = ">"
			}
			args = append(args, id)
		}
		result, err := execute(r, deps, plateID, "XREADGROUP", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/streams/groups/ack", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key   string   `json:"key"`
			Group string   `json:"group"`
			IDs   []string `json:"ids"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		group, err := requiredString(request.Group, "group")
		if err != nil {
			return err
		}
		if err := requireNonEmptyStrings(request.IDs, "ids"); err != nil {
			return err
		}
		args := []any{key, group}
		for _, id := range request.IDs {
			args = append(args, id)
		}
		result, err := execute(r, deps, plateID, "XACK", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/streams/groups/pending/{key}/{group}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		group, err := plate.PathValue(r, "group")
		if err != nil {
			return err
		}
		args := []any{key, group}
		if start := strings.TrimSpace(r.URL.Query().Get("start")); start != "" {
			end := queryString(r, "end", "+")
			count := queryString(r, "count", "10")
			args = append(args, start, end, count)
			if consumer := strings.TrimSpace(r.URL.Query().Get("consumer")); consumer != "" {
				args = append(args, consumer)
			}
		}
		result, err := execute(r, deps, plateID, "XPENDING", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/streams/groups/claim", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key       string   `json:"key"`
			Group     string   `json:"group"`
			Consumer  string   `json:"consumer"`
			MinIdleMS int64    `json:"min_idle_ms"`
			IDs       []string `json:"ids"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		group, err := requiredString(request.Group, "group")
		if err != nil {
			return err
		}
		consumer, err := requiredString(request.Consumer, "consumer")
		if err != nil {
			return err
		}
		if err := requireNonEmptyStrings(request.IDs, "ids"); err != nil {
			return err
		}
		args := []any{key, group, consumer, strconv.FormatInt(request.MinIdleMS, 10)}
		for _, id := range request.IDs {
			args = append(args, id)
		}
		result, err := execute(r, deps, plateID, "XCLAIM", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/streams/groups/autoclaim", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key       string `json:"key"`
			Group     string `json:"group"`
			Consumer  string `json:"consumer"`
			MinIdleMS int64  `json:"min_idle_ms"`
			Start     string `json:"start"`
			Count     *int64 `json:"count"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		group, err := requiredString(request.Group, "group")
		if err != nil {
			return err
		}
		consumer, err := requiredString(request.Consumer, "consumer")
		if err != nil {
			return err
		}
		start, err := requiredString(request.Start, "start")
		if err != nil {
			return err
		}
		args := []any{key, group, consumer, strconv.FormatInt(request.MinIdleMS, 10), start}
		if request.Count != nil {
			args = append(args, "COUNT", strconv.FormatInt(*request.Count, 10))
		}
		result, err := execute(r, deps, plateID, "XAUTOCLAIM", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
}

func streamRangeHandler(deps *plate.Dependencies, command string) http.HandlerFunc {
	return plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		startFallback := "-"
		endFallback := "+"
		if command == "XREVRANGE" {
			startFallback = "+"
			endFallback = "-"
		}
		start := queryString(r, "start", startFallback)
		end := queryString(r, "end", endFallback)
		args := []any{key, start, end}
		if count := strings.TrimSpace(r.URL.Query().Get("count")); count != "" {
			args = append(args, "COUNT", count)
		}
		result, err := execute(r, deps, plateID, command, args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	})
}
