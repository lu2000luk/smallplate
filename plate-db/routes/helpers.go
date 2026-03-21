package routes

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"plate/db/internal/plate"
)

type statementRequest struct {
	SQL    string `json:"sql"`
	Params []any  `json:"params"`
}

type batchRequest struct {
	Transaction *bool              `json:"transaction,omitempty"`
	Statements  []statementRequest `json:"statements"`
}

func validateStatement(stmt statementRequest) (string, []any, error) {
	sqlText := strings.TrimSpace(stmt.SQL)
	if sqlText == "" {
		return "", nil, plate.NewAPIError(http.StatusBadRequest, "invalid_request", "sql is required")
	}
	return sqlText, stmt.Params, nil
}

func statementType(sqlText string) string {
	trimmed := strings.TrimSpace(sqlText)
	if trimmed == "" {
		return "execute"
	}
	upper := strings.ToUpper(trimmed)
	for _, prefix := range []string{"SELECT", "PRAGMA", "WITH", "EXPLAIN"} {
		if strings.HasPrefix(upper, prefix) {
			return "query"
		}
	}
	return "execute"
}

func rowsToMatrix(rows *sql.Rows) ([]string, [][]any, error) {
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	values := make([][]any, 0)
	for rows.Next() {
		raw := make([]any, len(columns))
		scan := make([]any, len(columns))
		for i := range raw {
			scan[i] = &raw[i]
		}
		if err := rows.Scan(scan...); err != nil {
			return nil, nil, err
		}
		entry := make([]any, len(columns))
		for i, item := range raw {
			entry[i] = normalizeSQLiteValue(item)
		}
		values = append(values, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return columns, values, nil
}

func normalizeSQLiteValue(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case []byte:
		if len(typed) == 0 {
			return ""
		}
		if isMostlyText(typed) {
			return string(typed)
		}
		return base64.StdEncoding.EncodeToString(typed)
	case time.Time:
		return typed.UTC().Format(time.RFC3339Nano)
	default:
		return typed
	}
}

func isMostlyText(value []byte) bool {
	for _, b := range value {
		if b == '\n' || b == '\r' || b == '\t' {
			continue
		}
		if b < 32 || b > 126 {
			return false
		}
	}
	return true
}

func readTxnID(r *http.Request) (string, error) {
	txnID, err := plate.PathValue(r, "txnId")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(txnID) == "" {
		return "", plate.NewAPIError(http.StatusBadRequest, "invalid_request", "txnId is required")
	}
	return txnID, nil
}

func queryBoolDefaultTrue(r *batchRequest) bool {
	if r.Transaction == nil {
		return true
	}
	return *r.Transaction
}

func requireTableName(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", plate.NewAPIError(http.StatusBadRequest, "invalid_request", "table is required")
	}
	return trimmed, nil
}

func unknownFormatError(format string) error {
	return plate.NewAPIError(http.StatusBadRequest, "invalid_request", fmt.Sprintf("unsupported export format %q", format))
}
