package routes

import (
	"errors"
	"net/http"
	"os"

	"plate/db/internal/plate"
)

func handleCreateDB(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		created, err := deps.DBs.Ensure(r.Context(), dbID)
		if err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"id": dbID, "created": created})
		return nil
	}
}

func handleDBInfo(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		if _, err := deps.DBs.Ensure(r.Context(), dbID); err != nil {
			return err
		}
		path, err := deps.DBs.FilePath(dbID)
		if err != nil {
			return err
		}
		stat, err := os.Stat(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return plate.NewAPIError(http.StatusNotFound, "not_found", "database not found")
			}
			return err
		}
		createdAt, err := deps.DBs.CreatedAt(r.Context(), dbID)
		if err != nil {
			return err
		}
		tableCount, err := deps.DBs.TableCount(r.Context(), dbID)
		if err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{
			"id":         dbID,
			"createdAt":  createdAt,
			"size_bytes": stat.Size(),
			"tableCount": tableCount,
		})
		return nil
	}
}
