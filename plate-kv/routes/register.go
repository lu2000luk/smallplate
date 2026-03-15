package routes

import (
	"net/http"

	"plain/kv/internal/plate"
)

func Register(mux *http.ServeMux, deps *plate.Dependencies) {
	registerKeys(mux, deps)
	registerStrings(mux, deps)
	registerHashes(mux, deps)
	registerLists(mux, deps)
	registerSets(mux, deps)
	registerZSets(mux, deps)
	registerSpecial(mux, deps)
	registerStreams(mux, deps)
	registerBatch(mux, deps)
	registerPubSub(mux, deps)
	registerInfo(mux, deps)
	registerJSON(mux, deps)
}
