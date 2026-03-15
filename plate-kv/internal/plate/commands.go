package plate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

type CommandSpec struct {
	Name             string
	AllowBatch       bool
	AllowTransaction bool
	Rewrite          func(string, []any) ([]any, error)
}

func NewCommandRegistry() map[string]CommandSpec {
	registry := make(map[string]CommandSpec)
	register := func(name string, spec CommandSpec) {
		spec.Name = name
		registry[name] = spec
	}

	for _, name := range []string{"GET", "SET", "TYPE", "TTL", "PTTL", "PERSIST", "GETEX", "GETDEL", "INCR", "DECR", "INCRBY", "DECRBY", "INCRBYFLOAT", "APPEND", "STRLEN", "SETRANGE", "GETRANGE", "HSET", "HGET", "HMGET", "HGETALL", "HDEL", "HEXISTS", "HINCRBY", "HINCRBYFLOAT", "HKEYS", "HVALS", "HLEN", "HSETNX", "HRANDFIELD", "LPUSH", "RPUSH", "LPOP", "RPOP", "LLEN", "LRANGE", "LINDEX", "LPOS", "LSET", "LINSERT", "LREM", "LTRIM", "SADD", "SREM", "SMEMBERS", "SISMEMBER", "SMISMEMBER", "SCARD", "SRANDMEMBER", "SPOP", "ZADD", "ZREM", "ZSCORE", "ZRANK", "ZREVRANK", "ZRANGE", "ZCARD", "ZCOUNT", "ZLEXCOUNT", "ZINCRBY", "ZPOPMIN", "ZPOPMAX", "ZRANDMEMBER", "ZMSCORE", "SETBIT", "GETBIT", "BITCOUNT", "BITPOS", "BITFIELD", "GEOADD", "GEOPOS", "GEODIST", "GEOSEARCH", "XLEN", "XRANGE", "XREVRANGE", "XTRIM", "XDEL", "XACK", "XPENDING", "XCLAIM", "XAUTOCLAIM"} {
		register(name, simpleKeySpec(name, true, true))
	}

	for _, name := range []string{"EXPIRE", "PEXPIRE", "EXPIREAT", "PEXPIREAT"} {
		register(name, simpleKeySpec(name, true, true))
	}
	for _, name := range []string{"MGET", "DEL", "UNLINK", "EXISTS", "SUNION", "SINTER", "SDIFF"} {
		register(name, multiKeySpec(name, 0, true, true))
	}
	register("MSET", CommandSpec{AllowBatch: true, AllowTransaction: true, Rewrite: rewriteMSet})
	register("RENAME", pairKeySpec("RENAME", 0, 1, true, true))
	register("COPY", CommandSpec{AllowBatch: true, AllowTransaction: true, Rewrite: rewriteCopy})
	register("LMOVE", pairKeySpec("LMOVE", 0, 1, true, true))
	register("SUNIONSTORE", CommandSpec{AllowBatch: true, AllowTransaction: true, Rewrite: rewriteStoreSetOp})
	register("SINTERSTORE", CommandSpec{AllowBatch: true, AllowTransaction: true, Rewrite: rewriteStoreSetOp})
	register("SDIFFSTORE", CommandSpec{AllowBatch: true, AllowTransaction: true, Rewrite: rewriteStoreSetOp})
	register("ZUNIONSTORE", CommandSpec{AllowBatch: true, AllowTransaction: true, Rewrite: rewriteZStore})
	register("ZINTERSTORE", CommandSpec{AllowBatch: true, AllowTransaction: true, Rewrite: rewriteZStore})
	register("ZDIFFSTORE", CommandSpec{AllowBatch: true, AllowTransaction: true, Rewrite: rewriteZStore})
	register("ZRANGESTORE", pairKeySpec("ZRANGESTORE", 0, 1, true, true))
	register("BITOP", CommandSpec{AllowBatch: true, AllowTransaction: true, Rewrite: rewriteBitOp})
	register("GEOSEARCHSTORE", CommandSpec{AllowBatch: true, AllowTransaction: true, Rewrite: rewriteGeoSearchStore})
	register("SCAN", CommandSpec{AllowBatch: true, AllowTransaction: false, Rewrite: rewriteScan})
	register("HSCAN", simpleKeySpec("HSCAN", true, false))
	register("SSCAN", simpleKeySpec("SSCAN", true, false))
	register("ZSCAN", simpleKeySpec("ZSCAN", true, false))
	register("PUBLISH", CommandSpec{AllowBatch: false, AllowTransaction: false, Rewrite: rewritePublish})
	register("XADD", simpleKeySpec("XADD", true, true))
	register("XINFO", CommandSpec{AllowBatch: true, AllowTransaction: true, Rewrite: rewriteXInfo})
	register("XGROUP", CommandSpec{AllowBatch: true, AllowTransaction: true, Rewrite: rewriteXGroup})
	register("XREAD", CommandSpec{AllowBatch: false, AllowTransaction: false, Rewrite: rewriteXRead})
	register("XREADGROUP", CommandSpec{AllowBatch: false, AllowTransaction: false, Rewrite: rewriteXReadGroup})
	return registry
}

func ExecuteCommand(ctx context.Context, deps *Dependencies, plateID string, command string, args ...any) (any, error) {
	name := strings.ToUpper(strings.TrimSpace(command))
	spec, ok := deps.CommandRegistry[name]
	if !ok {
		return nil, NewAPIError(http.StatusBadRequest, "unsupported_command", "unsupported command")
	}
	rewritten, err := spec.Rewrite(plateID, args)
	if err != nil {
		return nil, err
	}
	result, err := deps.Redis.Do(ctx, append([]any{name}, rewritten...)...).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return NormalizeResult(result), nil
}

func Pipeline(ctx context.Context, deps *Dependencies, plateID string, raw [][]any) ([]any, error) {
	pipe := deps.Redis.Pipeline()
	cmds := make([]*redis.Cmd, 0, len(raw))
	for _, entry := range raw {
		if len(entry) == 0 {
			return nil, NewAPIError(http.StatusBadRequest, "invalid_command", "empty command in pipeline")
		}
		name, err := stringArg(entry[0])
		if err != nil {
			return nil, err
		}
		upper := strings.ToUpper(name)
		spec, ok := deps.CommandRegistry[upper]
		if !ok || !spec.AllowBatch {
			return nil, NewAPIError(http.StatusBadRequest, "unsupported_command", fmt.Sprintf("command %s is not allowed in pipeline", upper))
		}
		rewritten, err := spec.Rewrite(plateID, entry[1:])
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, pipe.Do(ctx, append([]any{upper}, rewritten...)...))
	}
	_, _ = pipe.Exec(ctx)
	results := make([]any, 0, len(cmds))
	for _, cmd := range cmds {
		if err := cmd.Err(); err != nil && err != redis.Nil {
			results = append(results, map[string]any{"error": err.Error()})
			continue
		}
		results = append(results, NormalizeResult(cmd.Val()))
	}
	return results, nil
}

func Transaction(ctx context.Context, deps *Dependencies, plateID string, raw [][]any) ([]any, error) {
	txPipe := deps.Redis.TxPipeline()
	cmds := make([]*redis.Cmd, 0, len(raw))
	for _, entry := range raw {
		if len(entry) == 0 {
			return nil, NewAPIError(http.StatusBadRequest, "invalid_command", "empty command in transaction")
		}
		name, err := stringArg(entry[0])
		if err != nil {
			return nil, err
		}
		upper := strings.ToUpper(name)
		spec, ok := deps.CommandRegistry[upper]
		if !ok || !spec.AllowTransaction {
			return nil, NewAPIError(http.StatusBadRequest, "unsupported_command", fmt.Sprintf("command %s is not allowed in transaction", upper))
		}
		rewritten, err := spec.Rewrite(plateID, entry[1:])
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, txPipe.Do(ctx, append([]any{upper}, rewritten...)...))
	}
	_, _ = txPipe.Exec(ctx)
	results := make([]any, 0, len(cmds))
	for _, cmd := range cmds {
		if err := cmd.Err(); err != nil && err != redis.Nil {
			results = append(results, map[string]any{"error": err.Error()})
			continue
		}
		results = append(results, NormalizeResult(cmd.Val()))
	}
	return results, nil
}

func simpleKeySpec(name string, allowBatch bool, allowTransaction bool) CommandSpec {
	return CommandSpec{
		Name:             name,
		AllowBatch:       allowBatch,
		AllowTransaction: allowTransaction,
		Rewrite: func(plateID string, args []any) ([]any, error) {
			return rewriteAtPositions(plateID, args, 0)
		},
	}
}

func multiKeySpec(name string, start int, allowBatch bool, allowTransaction bool) CommandSpec {
	return CommandSpec{
		Name:             name,
		AllowBatch:       allowBatch,
		AllowTransaction: allowTransaction,
		Rewrite: func(plateID string, args []any) ([]any, error) {
			indices := make([]int, 0, len(args)-start)
			for i := start; i < len(args); i++ {
				indices = append(indices, i)
			}
			return rewriteAtPositions(plateID, args, indices...)
		},
	}
}

func pairKeySpec(name string, first int, second int, allowBatch bool, allowTransaction bool) CommandSpec {
	return CommandSpec{
		Name:             name,
		AllowBatch:       allowBatch,
		AllowTransaction: allowTransaction,
		Rewrite: func(plateID string, args []any) ([]any, error) {
			return rewriteAtPositions(plateID, args, first, second)
		},
	}
}

func rewriteAtPositions(plateID string, args []any, positions ...int) ([]any, error) {
	if len(args) == 0 {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_command", "missing command arguments")
	}
	rewritten := append([]any(nil), args...)
	for _, position := range positions {
		if position >= len(rewritten) {
			return nil, NewAPIError(http.StatusBadRequest, "invalid_command", "missing key argument")
		}
		key, err := stringArg(rewritten[position])
		if err != nil {
			return nil, err
		}
		rewritten[position] = PrefixKey(plateID, key)
	}
	return rewritten, nil
}

func rewriteMSet(plateID string, args []any) ([]any, error) {
	if len(args) == 0 || len(args)%2 != 0 {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_command", "MSET expects key/value pairs")
	}
	rewritten := append([]any(nil), args...)
	for i := 0; i < len(rewritten); i += 2 {
		key, err := stringArg(rewritten[i])
		if err != nil {
			return nil, err
		}
		rewritten[i] = PrefixKey(plateID, key)
	}
	return rewritten, nil
}

func rewriteCopy(plateID string, args []any) ([]any, error) {
	rewritten, err := rewriteAtPositions(plateID, args, 0, 1)
	if err != nil {
		return nil, err
	}
	for index := 2; index < len(rewritten); index++ {
		arg, ok := rewritten[index].(string)
		if ok && strings.EqualFold(arg, "DB") {
			return nil, NewAPIError(http.StatusBadRequest, "invalid_command", "COPY with DB is not allowed")
		}
	}
	return rewritten, nil
}

func rewriteStoreSetOp(plateID string, args []any) ([]any, error) {
	return multiKeySpec("store", 0, true, true).Rewrite(plateID, args)
}

func rewriteZStore(plateID string, args []any) ([]any, error) {
	if len(args) < 2 {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_command", "missing zset store arguments")
	}
	destination, err := stringArg(args[0])
	if err != nil {
		return nil, err
	}
	numKeys, err := intArg(args[1])
	if err != nil || numKeys <= 0 {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_command", "invalid numkeys")
	}
	if len(args) < 2+numKeys {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_command", "missing zset source keys")
	}
	rewritten := append([]any(nil), args...)
	rewritten[0] = PrefixKey(plateID, destination)
	for index := 0; index < numKeys; index++ {
		key, err := stringArg(rewritten[2+index])
		if err != nil {
			return nil, err
		}
		rewritten[2+index] = PrefixKey(plateID, key)
	}
	return rewritten, nil
}

func rewriteBitOp(plateID string, args []any) ([]any, error) {
	if len(args) < 3 {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_command", "BITOP expects operation, destination, and sources")
	}
	rewritten := append([]any(nil), args...)
	destination, err := stringArg(rewritten[1])
	if err != nil {
		return nil, err
	}
	rewritten[1] = PrefixKey(plateID, destination)
	for index := 2; index < len(rewritten); index++ {
		key, err := stringArg(rewritten[index])
		if err != nil {
			return nil, err
		}
		rewritten[index] = PrefixKey(plateID, key)
	}
	return rewritten, nil
}

func rewriteGeoSearchStore(plateID string, args []any) ([]any, error) {
	return rewriteAtPositions(plateID, args, 0, 1)
}

func rewriteScan(plateID string, args []any) ([]any, error) {
	if len(args) == 0 {
		args = []any{"0"}
	}
	rewritten := append([]any(nil), args...)
	insertedMatch := false
	for index := 1; index < len(rewritten); index++ {
		flag, ok := rewritten[index].(string)
		if !ok {
			continue
		}
		if strings.EqualFold(flag, "MATCH") && index+1 < len(rewritten) {
			pattern, err := stringArg(rewritten[index+1])
			if err != nil {
				return nil, err
			}
			rewritten[index+1] = PrefixPattern(plateID, pattern)
			insertedMatch = true
			break
		}
	}
	if !insertedMatch {
		rewritten = append(rewritten, "MATCH", PrefixPattern(plateID, "*"))
	}
	return rewritten, nil
}

func rewritePublish(plateID string, args []any) ([]any, error) {
	return rewriteAtPositions(plateID, args, 0)
}

func rewriteXInfo(plateID string, args []any) ([]any, error) {
	if len(args) < 2 {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_command", "XINFO requires a subcommand and key")
	}
	rewritten := append([]any(nil), args...)
	keyIndex := 1
	key, err := stringArg(rewritten[keyIndex])
	if err != nil {
		return nil, err
	}
	rewritten[keyIndex] = PrefixKey(plateID, key)
	return rewritten, nil
}

func rewriteXGroup(plateID string, args []any) ([]any, error) {
	if len(args) < 2 {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_command", "XGROUP requires more arguments")
	}
	rewritten := append([]any(nil), args...)
	key, err := stringArg(rewritten[1])
	if err != nil {
		return nil, err
	}
	rewritten[1] = PrefixKey(plateID, key)
	return rewritten, nil
}

func rewriteXRead(plateID string, args []any) ([]any, error) {
	return rewriteStreamRead(plateID, args, "XREAD")
}

func rewriteXReadGroup(plateID string, args []any) ([]any, error) {
	return rewriteStreamRead(plateID, args, "XREADGROUP")
}

func rewriteStreamRead(plateID string, args []any, command string) ([]any, error) {
	rewritten := append([]any(nil), args...)
	streamsIndex := -1
	for index, value := range rewritten {
		flag, ok := value.(string)
		if ok && strings.EqualFold(flag, "STREAMS") {
			streamsIndex = index
			break
		}
	}
	if streamsIndex == -1 || streamsIndex == len(rewritten)-1 {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_command", command+" requires STREAMS")
	}
	remaining := len(rewritten) - (streamsIndex + 1)
	if remaining%2 != 0 {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_command", command+" requires matching stream keys and ids")
	}
	keyCount := remaining / 2
	for index := 0; index < keyCount; index++ {
		keyPosition := streamsIndex + 1 + index
		key, err := stringArg(rewritten[keyPosition])
		if err != nil {
			return nil, err
		}
		rewritten[keyPosition] = PrefixKey(plateID, key)
	}
	return rewritten, nil
}

func stringArg(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case fmt.Stringer:
		return typed.String(), nil
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), nil
	case int:
		return strconv.Itoa(typed), nil
	case int64:
		return strconv.FormatInt(typed, 10), nil
	case json.Number:
		return typed.String(), nil
	default:
		return "", NewAPIError(http.StatusBadRequest, "invalid_argument", fmt.Sprintf("expected string argument, got %T", value))
	}
}

func intArg(value any) (int, error) {
	switch typed := value.(type) {
	case int:
		return typed, nil
	case int64:
		return int(typed), nil
	case float64:
		return int(typed), nil
	case string:
		parsed, err := strconv.Atoi(typed)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return 0, err
		}
		return int(parsed), nil
	default:
		return 0, fmt.Errorf("invalid int argument %T", value)
	}
}
