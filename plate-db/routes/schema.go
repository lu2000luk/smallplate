package routes

import (
	"database/sql"
	"net/http"

	"plate/db/internal/plate"
)

func handleTables(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		db, err := deps.DBs.Open(r.Context(), dbID)
		if err != nil {
			return err
		}
		rows, err := db.QueryContext(r.Context(), "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name != '__plate_meta' ORDER BY name")
		if err != nil {
			return err
		}
		defer rows.Close()
		tables := make([]string, 0)
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return err
			}
			tables = append(tables, name)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"tables": tables})
		return nil
	}
}

func handleTableDetails(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		table, err := plate.PathValue(r, "table")
		if err != nil {
			return err
		}
		table, err = requireTableName(table)
		if err != nil {
			return err
		}

		db, err := deps.DBs.Open(r.Context(), dbID)
		if err != nil {
			return err
		}

		columnsRows, err := db.QueryContext(r.Context(), "PRAGMA table_info("+quoteIdentifier(table)+")")
		if err != nil {
			return err
		}
		defer columnsRows.Close()

		columns := make([]map[string]any, 0)
		for columnsRows.Next() {
			var cid int
			var name, typ string
			var notNull, pk int
			var defaultValue sql.NullString
			if err := columnsRows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
				return err
			}
			var defaultAny any
			if defaultValue.Valid {
				defaultAny = defaultValue.String
			}
			columns = append(columns, map[string]any{
				"name":       name,
				"type":       typ,
				"primaryKey": pk > 0,
				"nullable":   notNull == 0,
				"default":    defaultAny,
			})
		}
		if err := columnsRows.Err(); err != nil {
			return err
		}
		if len(columns) == 0 {
			return plate.NewAPIError(http.StatusNotFound, "not_found", "table not found")
		}

		fkRows, err := db.QueryContext(r.Context(), "PRAGMA foreign_key_list("+quoteIdentifier(table)+")")
		if err != nil {
			return err
		}
		defer fkRows.Close()
		foreignKeys := make([]map[string]any, 0)
		for fkRows.Next() {
			var id, seq int
			var refTable, from, to, onUpdate, onDelete, match string
			if err := fkRows.Scan(&id, &seq, &refTable, &from, &to, &onUpdate, &onDelete, &match); err != nil {
				return err
			}
			foreignKeys = append(foreignKeys, map[string]any{
				"id":       id,
				"seq":      seq,
				"table":    refTable,
				"from":     from,
				"to":       to,
				"onUpdate": onUpdate,
				"onDelete": onDelete,
				"match":    match,
			})
		}

		var rowCount int64
		if err := db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM "+quoteIdentifier(table)).Scan(&rowCount); err != nil {
			return err
		}

		var createSQL sql.NullString
		if err := db.QueryRowContext(r.Context(), "SELECT sql FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&createSQL); err != nil {
			return err
		}

		plate.WriteOK(w, http.StatusOK, map[string]any{
			"name":        table,
			"columns":     columns,
			"foreignKeys": foreignKeys,
			"rowCount":    rowCount,
			"sql":         createSQL.String,
		})
		return nil
	}
}

func handleTableIndexes(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		table, err := plate.PathValue(r, "table")
		if err != nil {
			return err
		}
		table, err = requireTableName(table)
		if err != nil {
			return err
		}

		db, err := deps.DBs.Open(r.Context(), dbID)
		if err != nil {
			return err
		}

		rows, err := db.QueryContext(r.Context(), "PRAGMA index_list("+quoteIdentifier(table)+")")
		if err != nil {
			return err
		}
		defer rows.Close()

		indexes := make([]map[string]any, 0)
		for rows.Next() {
			var seq int
			var name string
			var unique int
			var origin string
			var partial int
			if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
				return err
			}
			indexes = append(indexes, map[string]any{
				"name":    name,
				"unique":  unique == 1,
				"origin":  origin,
				"partial": partial == 1,
			})
		}
		if err := rows.Err(); err != nil {
			return err
		}

		plate.WriteOK(w, http.StatusOK, map[string]any{"table": table, "indexes": indexes})
		return nil
	}
}

func handleSchema(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, dbID string) error {
		db, err := deps.DBs.Open(r.Context(), dbID)
		if err != nil {
			return err
		}

		rows, err := db.QueryContext(r.Context(), "SELECT sql FROM sqlite_master WHERE sql IS NOT NULL AND name NOT LIKE 'sqlite_%' ORDER BY type, name")
		if err != nil {
			return err
		}
		defer rows.Close()

		statements := make([]string, 0)
		for rows.Next() {
			var sqlText string
			if err := rows.Scan(&sqlText); err != nil {
				return err
			}
			statements = append(statements, sqlText+";")
		}
		if err := rows.Err(); err != nil {
			return err
		}

		plate.WriteOK(w, http.StatusOK, map[string]any{"sql": joinLines(statements)})
		return nil
	}
}

func quoteIdentifier(name string) string {
	return `"` + stringsReplaceAll(name, `"`, `""`) + `"`
}

func stringsReplaceAll(value string, old string, new string) string {
	for {
		next := value
		idx := -1
		for i := 0; i+len(old) <= len(next); i++ {
			if next[i:i+len(old)] == old {
				idx = i
				break
			}
		}
		if idx < 0 {
			return value
		}
		value = value[:idx] + new + value[idx+len(old):]
	}
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	joined := lines[0]
	for i := 1; i < len(lines); i++ {
		joined += "\n" + lines[i]
	}
	return joined
}
