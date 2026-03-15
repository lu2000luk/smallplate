// Endpoints contained in this file:
// POST /{plateID}/sets/command
// POST /{plateID}/sets/{key}/command
package routes

import (
	"net/http"

	"plain/kv/internal/plate"
)

func registerSets(mux *http.ServeMux, deps *plate.Dependencies) {
	allowed := mustCommands("SADD", "SREM", "SMEMBERS", "SISMEMBER", "SMISMEMBER", "SCARD", "SRANDMEMBER", "SPOP", "SUNION", "SINTER", "SDIFF", "SUNIONSTORE", "SINTERSTORE", "SDIFFSTORE")
	mux.HandleFunc("POST /{plateID}/sets/command", handleCommand(deps, allowed, false))
	mux.HandleFunc("POST /{plateID}/sets/{key}/command", handleCommand(deps, allowed, true))
}
