package routes

import (
	"net/http"

	"plate/link/internal/plate"
)

func handleResolveCORS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r
		plate.WriteCORSHeaders(w)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleResolveJSON(deps *plate.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if plate.HandleCORS(w, r) {
			return
		}

		plateID, err := plate.PathValue(r, "plateID")
		if err != nil {
			plate.WriteError(w, err)
			return
		}
		id, err := plate.PathValue(r, "id")
		if err != nil {
			plate.WriteError(w, err)
			return
		}

		resolved, err := deps.Links.Resolve(r.Context(), plateID, id, parseTail(r), r.URL.Query())
		if err != nil {
			plate.WriteError(w, err)
			return
		}

		plate.WriteOK(w, http.StatusOK, map[string]any{
			"id":          resolved.Record.ID,
			"plate_id":    resolved.Record.PlateID,
			"destination": resolved.Destination,
			"status":      resolved.Status,
			"uses":        resolved.Record.Uses,
		})
	}
}

func handleRedirect(deps *plate.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := plate.PathValue(r, "id")
		if err != nil {
			plate.WriteError(w, err)
			return
		}

		resolved, err := deps.Links.ResolveByPublicID(r.Context(), id, parseTail(r), r.URL.Query())
		if err != nil {
			plate.WriteError(w, err)
			return
		}

		http.Redirect(w, r, resolved.Destination, http.StatusTemporaryRedirect)
	}
}
