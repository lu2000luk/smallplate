// Endpoints contained in this file:
// POST /{plateID}/strings/command
// POST /{plateID}/strings/{key}/command
package routes

import (
	"net/http"

	"plain/kv/internal/plate"
)

func registerStrings(mux *http.ServeMux, deps *plate.Dependencies) {
	allowed := mustCommands("GET", "SET", "MGET", "MSET", "GETEX", "GETDEL", "INCR", "DECR", "INCRBY", "DECRBY", "INCRBYFLOAT", "APPEND", "STRLEN", "SETRANGE", "GETRANGE")
	mux.HandleFunc("POST /{plateID}/strings/command", handleCommand(deps, allowed, false))
	mux.HandleFunc("POST /{plateID}/strings/{key}/command", handleCommand(deps, allowed, true))
}
