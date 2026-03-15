// Endpoints contained in this file:
// POST /{plateID}/hashes/command
// POST /{plateID}/hashes/{key}/command
package routes

import (
	"net/http"

	"plain/kv/internal/plate"
)

func registerHashes(mux *http.ServeMux, deps *plate.Dependencies) {
	allowed := mustCommands("HSET", "HGET", "HMGET", "HGETALL", "HDEL", "HEXISTS", "HINCRBY", "HINCRBYFLOAT", "HKEYS", "HVALS", "HLEN", "HSETNX", "HRANDFIELD")
	mux.HandleFunc("POST /{plateID}/hashes/command", handleCommand(deps, allowed, false))
	mux.HandleFunc("POST /{plateID}/hashes/{key}/command", handleCommand(deps, allowed, true))
}
