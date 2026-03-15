// Endpoints contained in this file:
// POST /{plateID}/streams/command
// POST /{plateID}/streams/{key}/command
package routes

import (
	"net/http"

	"plain/kv/internal/plate"
)

func registerStreams(mux *http.ServeMux, deps *plate.Dependencies) {
	allowed := mustCommands("XADD", "XLEN", "XRANGE", "XREVRANGE", "XREAD", "XTRIM", "XDEL", "XINFO", "XGROUP", "XREADGROUP", "XACK", "XPENDING", "XCLAIM", "XAUTOCLAIM")
	mux.HandleFunc("POST /{plateID}/streams/command", handleCommand(deps, allowed, false))
	mux.HandleFunc("POST /{plateID}/streams/{key}/command", handleCommand(deps, allowed, true))
}
