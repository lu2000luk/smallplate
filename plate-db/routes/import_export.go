package routes

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"os"
	"strings"

	"plate/db/internal/plate"
)

func handleExport(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		format := strings.TrimSpace(r.URL.Query().Get("format"))
		if format == "" {
			format = "sqlite"
		}
		if _, err := deps.DBs.Ensure(r.Context(), dbID); err != nil {
			return err
		}

		switch strings.ToLower(format) {
		case "sqlite":
			path, err := deps.DBs.FilePath(dbID)
			if err != nil {
				return err
			}
			payload, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", "attachment; filename=\""+dbID+".db\"")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(payload)
			return nil
		case "sql":
			db, err := deps.DBs.Open(r.Context(), dbID)
			if err != nil {
				return err
			}
			dump, err := dumpSQL(r.Context(), db)
			if err != nil {
				return err
			}
			w.Header().Set("Content-Type", "application/sql")
			w.Header().Set("Content-Disposition", "attachment; filename=\""+dbID+".sql\"")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(dump))
			return nil
		default:
			return unknownFormatError(format)
		}
	}
}

func handleImportDB(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		if len(payload) == 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "request body must not be empty")
		}
		deps.Transactions.DeleteDB(dbID)
		if err := deps.DBs.ReplaceWithBytes(r.Context(), dbID, payload); err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"id": dbID, "imported": true, "format": "sqlite"})
		return nil
	}
}

func handleImportSQL(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		sqlText := strings.TrimSpace(string(payload))
		if sqlText == "" {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "request body must not be empty")
		}
		db, err := deps.DBs.Open(r.Context(), dbID)
		if err != nil {
			return err
		}
		if _, err := db.ExecContext(r.Context(), sqlText); err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"id": dbID, "imported": true, "format": "sql"})
		return nil
	}
}

func dumpSQL(ctx context.Context, db *sql.DB) (string, error) {
	rows, err := db.QueryContext(ctx, "SELECT sql FROM sqlite_master WHERE sql IS NOT NULL AND name NOT LIKE 'sqlite_%' ORDER BY type, name")
	if err != nil {
		return "", err
	}
	defer rows.Close()
	parts := make([]string, 0)
	for rows.Next() {
		var sqlText string
		if err := rows.Scan(&sqlText); err != nil {
			return "", err
		}
		parts = append(parts, sqlText+";")
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	return joinLines(parts), nil
}
