// Endpoints contained in this file:
// POST /{plateID}/zsets/command
// POST /{plateID}/zsets/{key}/command
// POST /{plateID}/zsets/add
// POST /{plateID}/zsets/remove
// GET /{plateID}/zsets/score/{key}/{member}
// GET /{plateID}/zsets/rank/{key}/{member}
// GET /{plateID}/zsets/range/{key}
// GET /{plateID}/zsets/count/{key}
// GET /{plateID}/zsets/count-by-score/{key}
// GET /{plateID}/zsets/count-by-lex/{key}
// POST /{plateID}/zsets/increment
// POST /{plateID}/zsets/pop-min
// POST /{plateID}/zsets/pop-max
// GET /{plateID}/zsets/random/{key}
// POST /{plateID}/zsets/scores
// POST /{plateID}/zsets/union
// POST /{plateID}/zsets/intersect
// POST /{plateID}/zsets/diff
// POST /{plateID}/zsets/range/store
package routes

import (
	"net/http"
	"strconv"
	"strings"

	"plain/kv/internal/plate"
)

func registerZSets(mux *http.ServeMux, deps *plate.Dependencies) {
	allowed := mustCommands("ZADD", "ZREM", "ZSCORE", "ZRANK", "ZREVRANK", "ZRANGE", "ZCARD", "ZCOUNT", "ZLEXCOUNT", "ZINCRBY", "ZPOPMIN", "ZPOPMAX", "ZRANDMEMBER", "ZUNIONSTORE", "ZINTERSTORE", "ZDIFFSTORE", "ZRANGESTORE", "ZMSCORE")
	mux.HandleFunc("POST /{plateID}/zsets/command", handleCommand(deps, allowed, false))
	mux.HandleFunc("POST /{plateID}/zsets/{key}/command", handleCommand(deps, allowed, true))
	mux.HandleFunc("POST /{plateID}/zsets/add", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key     string `json:"key"`
			Members []struct {
				Member string `json:"member"`
				Score  any    `json:"score"`
			} `json:"members"`
			NX bool `json:"nx"`
			XX bool `json:"xx"`
			CH bool `json:"ch"`
			IN bool `json:"incr"`
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
		if request.NX && request.XX {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "nx and xx cannot both be true")
		}
		args := []any{key}
		if request.NX {
			args = append(args, "NX")
		}
		if request.XX {
			args = append(args, "XX")
		}
		if request.CH {
			args = append(args, "CH")
		}
		if request.IN {
			args = append(args, "INCR")
			if len(request.Members) != 1 {
				return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "incr requires exactly one member")
			}
		}
		for _, member := range request.Members {
			name, err := requiredString(member.Member, "member")
			if err != nil {
				return err
			}
			score, err := redisValueString(member.Score)
			if err != nil {
				return err
			}
			args = append(args, score, name)
		}
		result, err := execute(r, deps, plateID, "ZADD", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/zsets/remove", zsetMemberListHandler(deps, "ZREM"))
	mux.HandleFunc("GET /{plateID}/zsets/score/{key}/{member}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		member, err := plate.PathValue(r, "member")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "ZSCORE", key, member)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/zsets/rank/{key}/{member}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		member, err := plate.PathValue(r, "member")
		if err != nil {
			return err
		}
		command := "ZRANK"
		if strings.EqualFold(queryString(r, "order", "asc"), "desc") {
			command = "ZREVRANK"
		}
		result, err := execute(r, deps, plateID, command, key, member)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/zsets/range/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
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
		args := []any{key, strconv.FormatInt(start, 10), strconv.FormatInt(stop, 10)}
		if strings.EqualFold(queryString(r, "order", "asc"), "desc") {
			args = append(args, "REV")
		}
		if plate.QueryBool(r, "with_scores") {
			args = append(args, "WITHSCORES")
		}
		result, err := execute(r, deps, plateID, "ZRANGE", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/zsets/count/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "ZCARD", key)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/zsets/count-by-score/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		min, err := queryRequiredString(r, "min")
		if err != nil {
			return err
		}
		max, err := queryRequiredString(r, "max")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "ZCOUNT", key, min, max)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/zsets/count-by-lex/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		min, err := queryRequiredString(r, "min")
		if err != nil {
			return err
		}
		max, err := queryRequiredString(r, "max")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "ZLEXCOUNT", key, min, max)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/zsets/increment", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key    string `json:"key"`
			Member string `json:"member"`
			Amount any    `json:"amount"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		member, err := requiredString(request.Member, "member")
		if err != nil {
			return err
		}
		amount, err := redisValueString(request.Amount)
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "ZINCRBY", key, amount, member)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/zsets/pop-min", zsetPopHandler(deps, "ZPOPMIN"))
	mux.HandleFunc("POST /{plateID}/zsets/pop-max", zsetPopHandler(deps, "ZPOPMAX"))
	mux.HandleFunc("GET /{plateID}/zsets/random/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		args := []any{key}
		if count := strings.TrimSpace(r.URL.Query().Get("count")); count != "" {
			args = append(args, count)
		}
		if plate.QueryBool(r, "with_scores") {
			args = append(args, "WITHSCORES")
		}
		result, err := execute(r, deps, plateID, "ZRANDMEMBER", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/zsets/scores", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key     string   `json:"key"`
			Members []string `json:"members"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		if err := requireNonEmptyStrings(request.Members, "members"); err != nil {
			return err
		}
		args := []any{key}
		for _, member := range request.Members {
			args = append(args, member)
		}
		result, err := execute(r, deps, plateID, "ZMSCORE", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/zsets/union", zsetStoreHandler(deps, "ZUNIONSTORE"))
	mux.HandleFunc("POST /{plateID}/zsets/intersect", zsetStoreHandler(deps, "ZINTERSTORE"))
	mux.HandleFunc("POST /{plateID}/zsets/diff", zsetStoreHandler(deps, "ZDIFFSTORE"))
	mux.HandleFunc("POST /{plateID}/zsets/range/store", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Source      string `json:"source"`
			Destination string `json:"destination"`
			Start       int64  `json:"start"`
			Stop        int64  `json:"stop"`
			By          string `json:"by"`
			Reverse     bool   `json:"reverse"`
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
		args := []any{destination, source, strconv.FormatInt(request.Start, 10), strconv.FormatInt(request.Stop, 10)}
		if request.Reverse {
			args = append(args, "REV")
		}
		switch strings.ToUpper(strings.TrimSpace(request.By)) {
		case "SCORE":
			args = append(args, "BYSCORE")
		case "LEX":
			args = append(args, "BYLEX")
		case "", "RANK":
		default:
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "by must be rank, score, or lex")
		}
		result, err := execute(r, deps, plateID, "ZRANGESTORE", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
}

func zsetMemberListHandler(deps *plate.Dependencies, command string) http.HandlerFunc {
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

func zsetPopHandler(deps *plate.Dependencies, command string) http.HandlerFunc {
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

func zsetStoreHandler(deps *plate.Dependencies, command string) http.HandlerFunc {
	return plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Keys        []string `json:"keys"`
			Destination string   `json:"destination"`
			Weights     []any    `json:"weights"`
			Aggregate   string   `json:"aggregate"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		if err := requireNonEmptyStrings(request.Keys, "keys"); err != nil {
			return err
		}
		destination, err := requiredString(request.Destination, "destination")
		if err != nil {
			return err
		}
		args := []any{destination, strconv.Itoa(len(request.Keys))}
		for _, key := range request.Keys {
			args = append(args, key)
		}
		if len(request.Weights) > 0 {
			if len(request.Weights) != len(request.Keys) {
				return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "weights must match keys length")
			}
			weights, err := redisValueStrings(request.Weights)
			if err != nil {
				return err
			}
			args = append(args, "WEIGHTS")
			args = append(args, weights...)
		}
		if aggregate := strings.ToUpper(strings.TrimSpace(request.Aggregate)); aggregate != "" {
			if aggregate != "SUM" && aggregate != "MIN" && aggregate != "MAX" {
				return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "aggregate must be sum, min, or max")
			}
			args = append(args, "AGGREGATE", aggregate)
		}
		result, err := execute(r, deps, plateID, command, args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	})
}
