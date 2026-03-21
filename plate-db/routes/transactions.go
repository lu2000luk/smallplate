package routes

import (
	"net/http"
	"time"

	"plate/db/internal/plate"
)

func handleBeginTransaction(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		if _, err := deps.DBs.Ensure(r.Context(), dbID); err != nil {
			return err
		}
		txn, err := deps.Transactions.Start(r.Context(), dbID)
		if err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"txnId": txn.ID, "expiresAt": txn.ExpiresAt.UTC().Format(time.RFC3339Nano)})
		return nil
	}
}

func handleTxnQuery(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		txnID, err := readTxnID(r)
		if err != nil {
			return err
		}
		var request statementRequest
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		sqlText, params, err := validateStatement(request)
		if err != nil {
			return err
		}
		txn, err := deps.Transactions.Get(dbID, txnID)
		if err != nil {
			return err
		}
		started := time.Now()
		rows, err := txn.Tx.QueryContext(r.Context(), sqlText, params...)
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

func handleTxnExecute(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		txnID, err := readTxnID(r)
		if err != nil {
			return err
		}
		var request statementRequest
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		sqlText, params, err := validateStatement(request)
		if err != nil {
			return err
		}
		txn, err := deps.Transactions.Get(dbID, txnID)
		if err != nil {
			return err
		}
		started := time.Now()
		result, err := txn.Tx.ExecContext(r.Context(), sqlText, params...)
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

func handleTxnCommit(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		txnID, err := readTxnID(r)
		if err != nil {
			return err
		}
		if err := deps.Transactions.Commit(dbID, txnID); err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"txnId": txnID, "committed": true})
		return nil
	}
}

func handleTxnRollback(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		txnID, err := readTxnID(r)
		if err != nil {
			return err
		}
		if err := deps.Transactions.Rollback(dbID, txnID); err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"txnId": txnID, "rolledBack": true})
		return nil
	}
}
