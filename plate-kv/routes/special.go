// Endpoints contained in this file:
// POST /{plateID}/bitmaps/command
// POST /{plateID}/bitmaps/{key}/command
// POST /{plateID}/geo/command
// POST /{plateID}/geo/{key}/command
package routes

import (
	"net/http"

	"plain/kv/internal/plate"
)

func registerSpecial(mux *http.ServeMux, deps *plate.Dependencies) {
	bitmapAllowed := mustCommands("SETBIT", "GETBIT", "BITCOUNT", "BITOP", "BITPOS", "BITFIELD")
	geoAllowed := mustCommands("GEOADD", "GEOPOS", "GEODIST", "GEOSEARCH", "GEOSEARCHSTORE")
	mux.HandleFunc("POST /{plateID}/bitmaps/command", handleCommand(deps, bitmapAllowed, false))
	mux.HandleFunc("POST /{plateID}/bitmaps/{key}/command", handleCommand(deps, bitmapAllowed, true))
	mux.HandleFunc("POST /{plateID}/geo/command", handleCommand(deps, geoAllowed, false))
	mux.HandleFunc("POST /{plateID}/geo/{key}/command", handleCommand(deps, geoAllowed, true))
}
