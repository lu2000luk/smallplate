package routes

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"plate/db/internal/plate"
)

func handleQuery(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		var request statementRequest
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		sqlText, params, err := validateStatement(request)
		if err != nil {
			return err
		}
		db, err := deps.DBs.Open(r.Context(), dbID)
		if err != nil {
			return err
		}
		started := time.Now()
		rows, err := db.QueryContext(r.Context(), sqlText, params...)
		if err != nil {
			return err
		}
		columns, matrix, err := rowsToMatrix(rows)
		if err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{
			"columns":  columns,
			"rows":     matrix,
			"rowCount": len(matrix),
			"time_ms":  float64(time.Since(started).Microseconds()) / 1000.0,
		})
		return nil
	}
}

func handleExecute(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		var request statementRequest
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		sqlText, params, err := validateStatement(request)
		if err != nil {
			return err
		}
		db, err := deps.DBs.Open(r.Context(), dbID)
		if err != nil {
			return err
		}
		started := time.Now()
		result, err := db.ExecContext(r.Context(), sqlText, params...)
		if err != nil {
			return err
		}
		rowsAffected, _ := result.RowsAffected()
		lastInsertID, _ := result.LastInsertId()
		plate.WriteOK(w, http.StatusOK, map[string]any{
			"rowsAffected":    rowsAffected,
			"lastInsertRowid": lastInsertID,
			"time_ms":         float64(time.Since(started).Microseconds()) / 1000.0,
		})
		return nil
	}
}

func handleBatch(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		var request batchRequest
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		if len(request.Statements) == 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "statements must not be empty")
		}
		db, err := deps.DBs.Open(r.Context(), dbID)
		if err != nil {
			return err
		}

		useTransaction := queryBoolDefaultTrue(&request)
		started := time.Now()

		if !useTransaction {
			results, err := runBatchDirect(r.Context(), db, request.Statements)
			if err != nil {
				return err
			}
			plate.WriteOK(w, http.StatusOK, map[string]any{
				"results": results,
				"time_ms": float64(time.Since(started).Microseconds()) / 1000.0,
			})
			return nil
		}

		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			return err
		}
		results, err := runBatchTx(r.Context(), tx, request.Statements)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{
			"results": results,
			"time_ms": float64(time.Since(started).Microseconds()) / 1000.0,
		})
		return nil
	}
}

func runBatchDirect(ctx context.Context, db *sql.DB, statements []statementRequest) ([]any, error) {
	results := make([]any, 0, len(statements))
	for _, statement := range statements {
		sqlText, params, err := validateStatement(statement)
		if err != nil {
			return nil, err
		}
		if statementType(sqlText) == "query" {
			rows, err := db.QueryContext(ctx, sqlText, params...)
			if err != nil {
				return nil, err
			}
			columns, matrix, err := rowsToMatrix(rows)
			if err != nil {
				return nil, err
			}
			results = append(results, map[string]any{"type": "query", "columns": columns, "rows": matrix, "rowCount": len(matrix)})
			continue
		}
		res, err := db.ExecContext(ctx, sqlText, params...)
		if err != nil {
			return nil, err
		}
		rowsAffected, _ := res.RowsAffected()
		lastInsertID, _ := res.LastInsertId()
		results = append(results, map[string]any{"type": "execute", "rowsAffected": rowsAffected, "lastInsertRowid": lastInsertID})
	}
	return results, nil
}

func runBatchTx(ctx context.Context, tx *sql.Tx, statements []statementRequest) ([]any, error) {
	results := make([]any, 0, len(statements))
	for _, statement := range statements {
		sqlText, params, err := validateStatement(statement)
		if err != nil {
			return nil, err
		}
		if statementType(sqlText) == "query" {
			rows, err := tx.QueryContext(ctx, sqlText, params...)
			if err != nil {
				return nil, err
			}
			columns, matrix, err := rowsToMatrix(rows)
			if err != nil {
				return nil, err
			}
			results = append(results, map[string]any{"type": "query", "columns": columns, "rows": matrix, "rowCount": len(matrix)})
			continue
		}
		res, err := tx.ExecContext(ctx, sqlText, params...)
		if err != nil {
			return nil, err
		}
		rowsAffected, _ := res.RowsAffected()
		lastInsertID, _ := res.LastInsertId()
		results = append(results, map[string]any{"type": "execute", "rowsAffected": rowsAffected, "lastInsertRowid": lastInsertID})
	}
	return results, nil
}
