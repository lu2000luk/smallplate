// Endpoints contained in this file:
// POST /{plateID}/lists/command
// POST /{plateID}/lists/{key}/command
package routes

import (
	"net/http"

	"plain/kv/internal/plate"
)

func registerLists(mux *http.ServeMux, deps *plate.Dependencies) {
	allowed := mustCommands("LPUSH", "RPUSH", "LPOP", "RPOP", "LLEN", "LRANGE", "LINDEX", "LPOS", "LSET", "LINSERT", "LREM", "LTRIM", "LMOVE")
	mux.HandleFunc("POST /{plateID}/lists/command", handleCommand(deps, allowed, false))
	mux.HandleFunc("POST /{plateID}/lists/{key}/command", handleCommand(deps, allowed, true))
}
