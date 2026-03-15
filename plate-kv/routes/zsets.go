// Endpoints contained in this file:
// POST /{plateID}/zsets/command
// POST /{plateID}/zsets/{key}/command
package routes

import (
	"net/http"

	"plain/kv/internal/plate"
)

func registerZSets(mux *http.ServeMux, deps *plate.Dependencies) {
	allowed := mustCommands("ZADD", "ZREM", "ZSCORE", "ZRANK", "ZREVRANK", "ZRANGE", "ZCARD", "ZCOUNT", "ZLEXCOUNT", "ZINCRBY", "ZPOPMIN", "ZPOPMAX", "ZRANDMEMBER", "ZUNIONSTORE", "ZINTERSTORE", "ZDIFFSTORE", "ZRANGESTORE", "ZMSCORE")
	mux.HandleFunc("POST /{plateID}/zsets/command", handleCommand(deps, allowed, false))
	mux.HandleFunc("POST /{plateID}/zsets/{key}/command", handleCommand(deps, allowed, true))
}
