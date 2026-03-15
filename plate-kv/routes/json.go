// Endpoints contained in this file:
// GET /{plateID}/json/{key}
// POST /{plateID}/json/{key}
// DELETE /{plateID}/json/{key}
package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"plain/kv/internal/plate"
)

func registerJSON(mux *http.ServeMux, deps *plate.Dependencies) {
	mux.HandleFunc("GET /{plateID}/json/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		value, err := deps.Redis.Get(r.Context(), plate.PrefixKey(plateID, key)).Result()
		if err != nil {
			return err
		}
		var decoded any
		if err := json.Unmarshal([]byte(value), &decoded); err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, decoded)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/json/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		var request struct {
			Value any   `json:"value"`
			TTLMS int64 `json:"ttl_ms"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		payload, err := json.Marshal(request.Value)
		if err != nil {
			return err
		}
		ttl := time.Duration(request.TTLMS) * time.Millisecond
		if err := deps.Redis.Set(r.Context(), plate.PrefixKey(plateID, key), payload, ttl).Err(); err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"stored": true})
		return nil
	}))
	mux.HandleFunc("DELETE /{plateID}/json/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "DEL", key)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
}
