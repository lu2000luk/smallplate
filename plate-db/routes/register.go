package routes

import (
	"net/http"

	"plate/db/internal/plate"
)

func Register(mux *http.ServeMux, deps *plate.Dependencies) {
	mux.HandleFunc("POST /{id}", plate.Authenticated(deps, handleCreateDB(deps)))
	mux.HandleFunc("GET /{id}", plate.Authenticated(deps, handleDBInfo(deps)))
	mux.HandleFunc("GET /{id}/info", plate.Authenticated(deps, handleDBInfo(deps)))

	mux.HandleFunc("POST /{id}/query", plate.Authenticated(deps, handleQuery(deps)))
	mux.HandleFunc("POST /{id}/execute", plate.Authenticated(deps, handleExecute(deps)))
	mux.HandleFunc("POST /{id}/batch", plate.Authenticated(deps, handleBatch(deps)))

	mux.HandleFunc("POST /{id}/transactions", plate.Authenticated(deps, handleBeginTransaction(deps)))
	mux.HandleFunc("POST /{id}/transactions/{txnId}/query", plate.Authenticated(deps, handleTxnQuery(deps)))
	mux.HandleFunc("POST /{id}/transactions/{txnId}/execute", plate.Authenticated(deps, handleTxnExecute(deps)))
	mux.HandleFunc("POST /{id}/transactions/{txnId}/commit", plate.Authenticated(deps, handleTxnCommit(deps)))
	mux.HandleFunc("POST /{id}/transactions/{txnId}/rollback", plate.Authenticated(deps, handleTxnRollback(deps)))

	mux.HandleFunc("GET /{id}/tables", plate.Authenticated(deps, handleTables(deps)))
	mux.HandleFunc("GET /{id}/tables/{table}", plate.Authenticated(deps, handleTableDetails(deps)))
	mux.HandleFunc("GET /{id}/tables/{table}/indexes", plate.Authenticated(deps, handleTableIndexes(deps)))
	mux.HandleFunc("GET /{id}/schema", plate.Authenticated(deps, handleSchema(deps)))

	mux.HandleFunc("GET /{id}/export", plate.Authenticated(deps, handleExport(deps)))
	mux.HandleFunc("POST /{id}/import", plate.Authenticated(deps, handleImportDB(deps)))
	mux.HandleFunc("POST /{id}/import/sql", plate.Authenticated(deps, handleImportSQL(deps)))
}
