package routes

import (
	"net/http"

	"plate/vec/internal/plate"
)

func Register(mux *http.ServeMux, deps *plate.Dependencies) {
	h := newVectorHandler(deps)

	mux.HandleFunc("POST /{plateID}", plate.Authenticated(deps, h.handleEnsurePlate))
	mux.HandleFunc("GET /{plateID}", plate.Authenticated(deps, h.handlePlateInfo))
	mux.HandleFunc("GET /{plateID}/info", plate.Authenticated(deps, h.handlePlateInfo))
	mux.HandleFunc("GET /{plateID}/limits", plate.Authenticated(deps, h.handleLimits))

	mux.HandleFunc("GET /{plateID}/databases", plate.Authenticated(deps, h.handleListDatabases))
	mux.HandleFunc("POST /{plateID}/databases", plate.Authenticated(deps, h.handleCreateDatabase))
	mux.HandleFunc("GET /{plateID}/databases/{database}", plate.Authenticated(deps, h.handleGetDatabase))
	mux.HandleFunc("DELETE /{plateID}/databases/{database}", plate.Authenticated(deps, h.handleDeleteDatabase))
	mux.HandleFunc("POST /{plateID}/databases/{database}/reset", plate.Authenticated(deps, h.handleResetDatabase))

	mux.HandleFunc("GET /{plateID}/databases/{database}/collections", plate.Authenticated(deps, h.handleListCollections))
	mux.HandleFunc("GET /{plateID}/databases/{database}/collections/count", plate.Authenticated(deps, h.handleCountCollections))
	mux.HandleFunc("POST /{plateID}/databases/{database}/collections", plate.Authenticated(deps, h.handleCreateCollection))
	mux.HandleFunc("GET /{plateID}/databases/{database}/collections/{collection}", plate.Authenticated(deps, h.handleGetCollection))
	mux.HandleFunc("PATCH /{plateID}/databases/{database}/collections/{collection}", plate.Authenticated(deps, h.handlePatchCollection))
	mux.HandleFunc("DELETE /{plateID}/databases/{database}/collections/{collection}", plate.Authenticated(deps, h.handleDeleteCollection))
	mux.HandleFunc("POST /{plateID}/databases/{database}/collections/{collection}/peek", plate.Authenticated(deps, h.handlePeekCollection))

	mux.HandleFunc("POST /{plateID}/databases/{database}/collections/{collection}/records/add", plate.Authenticated(deps, h.handleAddRecords))
	mux.HandleFunc("POST /{plateID}/databases/{database}/collections/{collection}/records/update", plate.Authenticated(deps, h.handleUpdateRecords))
	mux.HandleFunc("POST /{plateID}/databases/{database}/collections/{collection}/records/upsert", plate.Authenticated(deps, h.handleUpsertRecords))
	mux.HandleFunc("POST /{plateID}/databases/{database}/collections/{collection}/records/get", plate.Authenticated(deps, h.handleGetRecords))
	mux.HandleFunc("POST /{plateID}/databases/{database}/collections/{collection}/records/query", plate.Authenticated(deps, h.handleQueryRecords))
	mux.HandleFunc("POST /{plateID}/databases/{database}/collections/{collection}/records/delete", plate.Authenticated(deps, h.handleDeleteRecords))
	mux.HandleFunc("GET /{plateID}/databases/{database}/collections/{collection}/records/count", plate.Authenticated(deps, h.handleCountRecords))

	mux.HandleFunc("GET /{plateID}/collections", plate.Authenticated(deps, h.handleAliasListCollections))
	mux.HandleFunc("POST /{plateID}/collections", plate.Authenticated(deps, h.handleAliasCreateCollection))
	mux.HandleFunc("GET /{plateID}/collections/{collection}", plate.Authenticated(deps, h.handleAliasGetCollection))
	mux.HandleFunc("PATCH /{plateID}/collections/{collection}", plate.Authenticated(deps, h.handleAliasPatchCollection))
	mux.HandleFunc("DELETE /{plateID}/collections/{collection}", plate.Authenticated(deps, h.handleAliasDeleteCollection))
	mux.HandleFunc("POST /{plateID}/collections/{collection}/records/get", plate.Authenticated(deps, h.handleAliasGetRecords))
	mux.HandleFunc("POST /{plateID}/collections/{collection}/records/query", plate.Authenticated(deps, h.handleAliasQueryRecords))
	mux.HandleFunc("POST /{plateID}/collections/{collection}/records/upsert", plate.Authenticated(deps, h.handleAliasUpsertRecords))

	mux.HandleFunc("POST /{plateID}/collections/{collection}/documents/upsert", plate.Authenticated(deps, h.handleDocumentsUpsert))
	mux.HandleFunc("POST /{plateID}/collections/{collection}/documents/get", plate.Authenticated(deps, h.handleDocumentsGet))
	mux.HandleFunc("POST /{plateID}/collections/{collection}/documents/search", plate.Authenticated(deps, h.handleDocumentsSearch))
	mux.HandleFunc("POST /{plateID}/collections/{collection}/documents/delete", plate.Authenticated(deps, h.handleDocumentsDelete))

	mux.HandleFunc("POST /{plateID}/search", plate.Authenticated(deps, h.handleCrossCollectionSearch))

	mux.HandleFunc("GET /{plateID}/embedding-profiles", plate.Authenticated(deps, h.handleListEmbeddingProfiles))
	mux.HandleFunc("POST /{plateID}/embedding-profiles", plate.Authenticated(deps, h.handleCreateEmbeddingProfile))
	mux.HandleFunc("PATCH /{plateID}/embedding-profiles/{profileID}", plate.Authenticated(deps, h.handlePatchEmbeddingProfile))
	mux.HandleFunc("DELETE /{plateID}/embedding-profiles/{profileID}", plate.Authenticated(deps, h.handleDeleteEmbeddingProfile))

	mux.HandleFunc("GET /{plateID}/embedding-defaults", plate.Authenticated(deps, h.handleGetEmbeddingDefaults))
	mux.HandleFunc("PUT /{plateID}/embedding-defaults", plate.Authenticated(deps, h.handlePutEmbeddingDefaults))

	mux.HandleFunc("POST /{plateID}/embeddings/run", plate.Authenticated(deps, h.handleEmbeddingsRun))
	mux.HandleFunc("GET /{plateID}/embedding-providers", plate.Authenticated(deps, h.handleEmbeddingProviders))
	mux.HandleFunc("GET /{plateID}/embedding-models", plate.Authenticated(deps, h.handleEmbeddingModels))

	mux.HandleFunc("GET /{plateID}/export", plate.Authenticated(deps, h.handleExport))
	mux.HandleFunc("POST /{plateID}/import", plate.Authenticated(deps, h.handleImport))
	mux.HandleFunc("POST /{plateID}/databases/{database}/collections/{collection}/clone", plate.Authenticated(deps, h.handleCloneCollection))
	mux.HandleFunc("POST /{plateID}/databases/{database}/collections/{collection}/copy", plate.Authenticated(deps, h.handleCopyCollection))
	mux.HandleFunc("POST /{plateID}/databases/{database}/collections/{collection}/reindex", plate.Authenticated(deps, h.handleReindexCollection))
	mux.HandleFunc("GET /{plateID}/databases/{database}/collections/{collection}/stats", plate.Authenticated(deps, h.handleCollectionStats))
}
