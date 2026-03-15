// Endpoints contained in this file:
// POST /{plateID}/pipeline
// POST /{plateID}/transaction
package routes

import (
	"net/http"

	"plain/kv/internal/plate"
)

func registerBatch(mux *http.ServeMux, deps *plate.Dependencies) {
	mux.HandleFunc("POST /{plateID}/pipeline", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		commands, err := parseCommandMatrix(r)
		if err != nil {
			return err
		}
		results, err := plate.Pipeline(r.Context(), deps, plateID, commands)
		if err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, results)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/transaction", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		commands, err := parseCommandMatrix(r)
		if err != nil {
			return err
		}
		results, err := plate.Transaction(r.Context(), deps, plateID, commands)
		if err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, results)
		return nil
	}))
}
