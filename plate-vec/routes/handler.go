package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"plate/vec/internal/plate"
)

const defaultDatabase = "default"

type vectorHandler struct {
	deps *plate.Dependencies

	embedClient *http.Client
}

func newVectorHandler(deps *plate.Dependencies) *vectorHandler {
	return &vectorHandler{
		deps: deps,
		embedClient: &http.Client{
			Timeout: deps.Config.OpTimeout,
		},
	}
}

type ensurePlateRequest struct {
	CreateDefaultDB *bool `json:"create_default_db,omitempty"`
}

type createDatabaseRequest struct {
	Name string `json:"name"`
}

type resetDatabaseRequest struct {
	Confirm bool `json:"confirm"`
}

type embeddingConfig struct {
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	Dimensions *int   `json:"dimensions,omitempty"`

	BaseURL  *string `json:"base_url,omitempty"`
	Endpoint *string `json:"endpoint,omitempty"`
}

type profileRequest struct {
	Name       string `json:"name"`
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	Dimensions *int   `json:"dimensions,omitempty"`
	APIKeyMode string `json:"api_key_mode"`
	APIKey     string `json:"api_key,omitempty"`
}

type defaultsRequest struct {
	DefaultProfileID         string `json:"default_profile_id"`
	FallbackToHeaderProvider bool   `json:"fallback_to_header_provider"`
}

type collectionCreateRequest struct {
	Name               string         `json:"name"`
	GetOrCreate        bool           `json:"get_or_create"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	Dimension          *int           `json:"dimension,omitempty"`
	DistanceMetric     string         `json:"distance_metric,omitempty"`
	EmbeddingProfileID string         `json:"embedding_profile_id,omitempty"`
	Index              map[string]any `json:"index,omitempty"`
}

type collectionPatchRequest struct {
	NewName            *string        `json:"new_name,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	Dimension          *int           `json:"dimension,omitempty"`
	DistanceMetric     *string        `json:"distance_metric,omitempty"`
	EmbeddingProfileID *string        `json:"embedding_profile_id,omitempty"`
}

type recordEntry struct {
	ID        string         `json:"id"`
	Document  *string        `json:"document,omitempty"`
	Embedding []float64      `json:"embedding,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	URI       *string        `json:"uri,omitempty"`
}

type recordWriteRequest struct {
	Records   []recordEntry    `json:"records"`
	Embedding *embeddingConfig `json:"embedding,omitempty"`
}

type recordGetRequest struct {
	IDs           []string       `json:"ids,omitempty"`
	Where         map[string]any `json:"where,omitempty"`
	WhereDocument map[string]any `json:"where_document,omitempty"`
	Include       []string       `json:"include,omitempty"`
	Limit         *int           `json:"limit,omitempty"`
	Offset        *int           `json:"offset,omitempty"`
}

type recordDeleteRequest struct {
	IDs           []string       `json:"ids,omitempty"`
	Where         map[string]any `json:"where,omitempty"`
	WhereDocument map[string]any `json:"where_document,omitempty"`
	Limit         *int           `json:"limit,omitempty"`
}

type recordQueryRequest struct {
	QueryTexts      []string         `json:"query_texts,omitempty"`
	QueryEmbeddings [][]float64      `json:"query_embeddings,omitempty"`
	Where           map[string]any   `json:"where,omitempty"`
	WhereDocument   map[string]any   `json:"where_document,omitempty"`
	NResults        *int             `json:"n_results,omitempty"`
	Include         []string         `json:"include,omitempty"`
	Embedding       *embeddingConfig `json:"embedding,omitempty"`
}

type documentsUpsertRequest struct {
	Documents []struct {
		ID        string         `json:"id"`
		Text      string         `json:"text"`
		Metadata  map[string]any `json:"metadata,omitempty"`
		Embedding []float64      `json:"embedding,omitempty"`
	} `json:"documents"`
	Embedding *embeddingConfig `json:"embedding,omitempty"`
}

type documentsGetRequest struct {
	IDs []string `json:"ids,omitempty"`
}

type documentsSearchRequest struct {
	Query         any              `json:"query"`
	NResults      *int             `json:"n_results,omitempty"`
	Where         map[string]any   `json:"where,omitempty"`
	WhereDocument map[string]any   `json:"where_document,omitempty"`
	Embedding     *embeddingConfig `json:"embedding,omitempty"`
}

type crossSearchRequest struct {
	Collections []string         `json:"collections"`
	Query       any              `json:"query"`
	NResults    *int             `json:"n_results,omitempty"`
	Where       map[string]any   `json:"where,omitempty"`
	Embedding   *embeddingConfig `json:"embedding,omitempty"`
}

type exportSnapshot struct {
	Database    string             `json:"database"`
	Collections []exportCollection `json:"collections"`
}

type exportCollection struct {
	Name     string         `json:"name"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Records  []recordEntry  `json:"records"`
}

type importRequest struct {
	Database    string `json:"database"`
	Collections []struct {
		Name     string         `json:"name"`
		Metadata map[string]any `json:"metadata,omitempty"`
		Records  []recordEntry  `json:"records"`
	} `json:"collections"`
}

func (h *vectorHandler) handleEnsurePlate(w http.ResponseWriter, r *http.Request, plateID string) error {
	var req ensurePlateRequest
	if r.Body != nil {
		_ = plate.DecodeJSON(r, &req)
	}
	if _, err := h.ensureTenantAndDB(r.Context(), plateID, defaultDatabase); err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"plate_id": plateID, "database": defaultDatabase, "created": true})
	return nil
}

func (h *vectorHandler) handlePlateInfo(w http.ResponseWriter, r *http.Request, plateID string) error {
	if _, err := h.ensureTenantAndDB(r.Context(), plateID, defaultDatabase); err != nil {
		return err
	}
	dbs, err := h.listDatabases(r.Context(), plateID)
	if err != nil {
		return err
	}
	collectionCount := 0
	recordCount := 0
	for _, db := range dbs {
		collections, listErr := h.listCollections(r.Context(), plateID, db)
		if listErr != nil {
			continue
		}
		collectionCount += len(collections)
		for _, item := range collections {
			count, countErr := h.countRecordsByName(r.Context(), plateID, db, item)
			if countErr == nil {
				recordCount += count
			}
		}
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{
		"plate_id":           plateID,
		"database_count":     len(dbs),
		"collection_count":   collectionCount,
		"record_count":       recordCount,
		"capabilities":       []string{"collections", "records", "profiles", "defaults", "import", "export"},
		"default_database":   defaultDatabase,
		"embedding_provider": []string{"openai", "openrouter"},
	})
	return nil
}

func (h *vectorHandler) handleLimits(w http.ResponseWriter, r *http.Request, plateID string) error {
	_ = plateID
	if err := h.deps.Chroma.Heartbeat(r.Context()); err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{
		"max_batch_size":    1000,
		"max_query_results": 100,
	})
	return nil
}

func (h *vectorHandler) handleListDatabases(w http.ResponseWriter, r *http.Request, plateID string) error {
	dbs, err := h.listDatabases(r.Context(), plateID)
	if err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"databases": dbs})
	return nil
}

func (h *vectorHandler) handleCreateDatabase(w http.ResponseWriter, r *http.Request, plateID string) error {
	var req createDatabaseRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "name is required")
	}
	if _, err := h.ensureTenantAndDB(r.Context(), plateID, name); err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"name": name})
	return nil
}

func (h *vectorHandler) handleGetDatabase(w http.ResponseWriter, r *http.Request, plateID string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	response, err := h.deps.Chroma.Request(r.Context(), http.MethodGet, fmt.Sprintf("/api/v2/tenants/%s/databases/%s", urlPath(plateID), urlPath(database)), nil)
	if err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, response)
	return nil
}

func (h *vectorHandler) handleDeleteDatabase(w http.ResponseWriter, r *http.Request, plateID string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	_, err = h.deps.Chroma.Request(r.Context(), http.MethodDelete, fmt.Sprintf("/api/v2/tenants/%s/databases/%s", urlPath(plateID), urlPath(database)), nil)
	if err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"deleted": true, "name": database})
	return nil
}

func (h *vectorHandler) handleResetDatabase(w http.ResponseWriter, r *http.Request, plateID string) error {
	var req resetDatabaseRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	if !req.Confirm {
		return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "confirm must be true")
	}
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	collections, err := h.listCollections(r.Context(), plateID, database)
	if err != nil {
		return err
	}
	for _, name := range collections {
		if delErr := h.deleteCollectionByName(r.Context(), plateID, database, name); delErr != nil {
			return delErr
		}
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"reset": true, "database": database})
	return nil
}

func (h *vectorHandler) handleListCollections(w http.ResponseWriter, r *http.Request, plateID string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	collections, err := h.listCollectionsDetailed(r.Context(), plateID, database)
	if err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"collections": collections})
	return nil
}

func (h *vectorHandler) handleCountCollections(w http.ResponseWriter, r *http.Request, plateID string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	collections, err := h.listCollections(r.Context(), plateID, database)
	if err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"count": len(collections)})
	return nil
}

func (h *vectorHandler) handleCreateCollection(w http.ResponseWriter, r *http.Request, plateID string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	var req collectionCreateRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "name is required")
	}
	if req.EmbeddingProfileID != "" {
		profile, getErr := h.deps.Meta.GetProfile(plateID, req.EmbeddingProfileID)
		if getErr != nil {
			return getErr
		}
		if profile == nil {
			return plate.NewAPIError(http.StatusNotFound, "not_found", "embedding profile not found")
		}
	}
	body := map[string]any{
		"name":          req.Name,
		"get_or_create": req.GetOrCreate,
	}
	if req.Metadata != nil {
		body["metadata"] = req.Metadata
	}
	if req.Index != nil {
		body["configuration"] = req.Index
	}
	response, err := h.deps.Chroma.Request(r.Context(), http.MethodPost, fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections", urlPath(plateID), urlPath(database)), body)
	if err != nil {
		return err
	}
	settings := plate.CollectionSettings{Dimension: req.Dimension, DistanceMetric: strings.TrimSpace(req.DistanceMetric), EmbeddingProfileID: strings.TrimSpace(req.EmbeddingProfileID)}
	if saveErr := h.deps.Meta.SaveCollectionSettings(plateID, database, req.Name, settings); saveErr != nil {
		return saveErr
	}
	plate.WriteOK(w, http.StatusOK, response)
	return nil
}

func (h *vectorHandler) handleGetCollection(w http.ResponseWriter, r *http.Request, plateID string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	collection, err := plate.PathValue(r, "collection")
	if err != nil {
		return err
	}
	response, err := h.deps.Chroma.Request(r.Context(), http.MethodGet, fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections/%s", urlPath(plateID), urlPath(database), urlPath(collection)), nil)
	if err != nil {
		return err
	}
	settings, _ := h.deps.Meta.GetCollectionSettings(plateID, database, collection)
	if settings != nil {
		response["embedding_profile_id"] = settings.EmbeddingProfileID
		if settings.Dimension != nil {
			response["dimension"] = *settings.Dimension
		}
		if settings.DistanceMetric != "" {
			response["distance_metric"] = settings.DistanceMetric
		}
	}
	plate.WriteOK(w, http.StatusOK, response)
	return nil
}

func (h *vectorHandler) handlePatchCollection(w http.ResponseWriter, r *http.Request, plateID string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	collection, err := plate.PathValue(r, "collection")
	if err != nil {
		return err
	}
	var req collectionPatchRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	body := map[string]any{}
	if req.NewName != nil {
		trimmed := strings.TrimSpace(*req.NewName)
		if trimmed != "" {
			body["new_name"] = trimmed
		}
	}
	if req.Metadata != nil {
		body["new_metadata"] = req.Metadata
	}
	if len(body) > 0 {
		_, err = h.deps.Chroma.Request(r.Context(), http.MethodPut, fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections/%s", urlPath(plateID), urlPath(database), urlPath(collection)), body)
		if err != nil {
			return err
		}
	}
	settings, _ := h.deps.Meta.GetCollectionSettings(plateID, database, collection)
	if settings == nil {
		settings = &plate.CollectionSettings{}
	}
	if req.Dimension != nil {
		settings.Dimension = req.Dimension
	}
	if req.DistanceMetric != nil {
		settings.DistanceMetric = strings.TrimSpace(*req.DistanceMetric)
	}
	if req.EmbeddingProfileID != nil {
		settings.EmbeddingProfileID = strings.TrimSpace(*req.EmbeddingProfileID)
	}
	if saveErr := h.deps.Meta.SaveCollectionSettings(plateID, database, collection, *settings); saveErr != nil {
		return saveErr
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"updated": true})
	return nil
}

func (h *vectorHandler) handleDeleteCollection(w http.ResponseWriter, r *http.Request, plateID string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	collection, err := plate.PathValue(r, "collection")
	if err != nil {
		return err
	}
	if err := h.deleteCollectionByName(r.Context(), plateID, database, collection); err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"deleted": true, "collection": collection})
	return nil
}

func (h *vectorHandler) handlePeekCollection(w http.ResponseWriter, r *http.Request, plateID string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	collection, err := plate.PathValue(r, "collection")
	if err != nil {
		return err
	}
	collectionID, err := h.resolveCollectionID(r.Context(), plateID, database, collection)
	if err != nil {
		return err
	}
	result, err := h.deps.Chroma.Request(r.Context(), http.MethodPost, fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections/%s/get", urlPath(plateID), urlPath(database), urlPath(collectionID)), map[string]any{"limit": 10})
	if err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, result)
	return nil
}

func (h *vectorHandler) handleAddRecords(w http.ResponseWriter, r *http.Request, plateID string) error {
	return h.handleWriteRecords(w, r, plateID, "add")
}

func (h *vectorHandler) handleUpdateRecords(w http.ResponseWriter, r *http.Request, plateID string) error {
	return h.handleWriteRecords(w, r, plateID, "update")
}

func (h *vectorHandler) handleUpsertRecords(w http.ResponseWriter, r *http.Request, plateID string) error {
	return h.handleWriteRecords(w, r, plateID, "upsert")
}

func (h *vectorHandler) handleWriteRecords(w http.ResponseWriter, r *http.Request, plateID string, op string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	collection, err := plate.PathValue(r, "collection")
	if err != nil {
		return err
	}
	var req recordWriteRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	if len(req.Records) == 0 {
		return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "records must not be empty")
	}
	if len(req.Records) > 1000 {
		return plate.NewAPIError(http.StatusBadRequest, "batch_too_large", "records exceed max batch size")
	}
	resolved, err := h.resolveEmbeddingForWrite(r.Context(), r.Header.Get("X-Embedding-Api-Key"), plateID, database, collection, req.Embedding, req.Records)
	if err != nil {
		return err
	}
	ids := make([]string, 0, len(req.Records))
	documents := make([]string, 0, len(req.Records))
	metadatas := make([]map[string]any, 0, len(req.Records))
	uris := make([]string, 0, len(req.Records))
	embeddings := make([][]float64, 0, len(req.Records))

	for idx, record := range req.Records {
		if strings.TrimSpace(record.ID) == "" {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "record id is required")
		}
		ids = append(ids, strings.TrimSpace(record.ID))
		if record.Document != nil {
			documents = append(documents, *record.Document)
		} else {
			documents = append(documents, "")
		}
		if record.Metadata != nil {
			metadatas = append(metadatas, record.Metadata)
		} else {
			metadatas = append(metadatas, map[string]any{})
		}
		if record.URI != nil {
			uris = append(uris, *record.URI)
		} else {
			uris = append(uris, "")
		}
		if len(record.Embedding) > 0 {
			embeddings = append(embeddings, record.Embedding)
			continue
		}
		if len(resolved.Generated) > idx {
			embeddings = append(embeddings, resolved.Generated[idx])
		}
	}
	if len(embeddings) != len(ids) {
		return plate.NewAPIError(http.StatusBadRequest, "embedding_not_configured", "missing embeddings for one or more records")
	}
	if resolved.Dimension > 0 {
		for _, item := range embeddings {
			if len(item) != resolved.Dimension {
				return plate.NewAPIError(http.StatusBadRequest, "dimension_mismatch", "embedding dimension mismatch")
			}
		}
	}
	collectionID, err := h.resolveCollectionID(r.Context(), plateID, database, collection)
	if err != nil {
		return err
	}
	result, err := h.deps.Chroma.Request(r.Context(), http.MethodPost, fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections/%s/%s", urlPath(plateID), urlPath(database), urlPath(collectionID), op), map[string]any{
		"ids":        ids,
		"embeddings": embeddings,
		"documents":  documents,
		"metadatas":  metadatas,
		"uris":       uris,
	})
	if err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, result)
	return nil
}

func (h *vectorHandler) handleGetRecords(w http.ResponseWriter, r *http.Request, plateID string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	collection, err := plate.PathValue(r, "collection")
	if err != nil {
		return err
	}
	var req recordGetRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	collectionID, err := h.resolveCollectionID(r.Context(), plateID, database, collection)
	if err != nil {
		return err
	}
	body := map[string]any{}
	if len(req.IDs) > 0 {
		body["ids"] = req.IDs
	}
	if req.Where != nil {
		body["where"] = req.Where
	}
	if req.WhereDocument != nil {
		body["where_document"] = req.WhereDocument
	}
	if req.Include != nil {
		body["include"] = req.Include
	}
	if req.Limit != nil {
		body["limit"] = *req.Limit
	}
	if req.Offset != nil {
		body["offset"] = *req.Offset
	}
	result, err := h.deps.Chroma.Request(r.Context(), http.MethodPost, fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections/%s/get", urlPath(plateID), urlPath(database), urlPath(collectionID)), body)
	if err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, result)
	return nil
}

func (h *vectorHandler) handleQueryRecords(w http.ResponseWriter, r *http.Request, plateID string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	collection, err := plate.PathValue(r, "collection")
	if err != nil {
		return err
	}
	var req recordQueryRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	if req.Where != nil {
		if _, err := json.Marshal(req.Where); err != nil {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_filter", err.Error())
		}
	}
	if req.WhereDocument != nil {
		if _, err := json.Marshal(req.WhereDocument); err != nil {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_filter", err.Error())
		}
	}
	resolved, err := h.resolveEmbeddingForQuery(r.Context(), r.Header.Get("X-Embedding-Api-Key"), plateID, database, collection, req.Embedding, req.QueryTexts, req.QueryEmbeddings)
	if err != nil {
		return err
	}
	if len(resolved.QueryEmbeddings) == 0 {
		return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "query_embeddings or query_texts is required")
	}
	if resolved.Dimension > 0 {
		for _, item := range resolved.QueryEmbeddings {
			if len(item) != resolved.Dimension {
				return plate.NewAPIError(http.StatusBadRequest, "dimension_mismatch", "query embedding dimension mismatch")
			}
		}
	}
	collectionID, err := h.resolveCollectionID(r.Context(), plateID, database, collection)
	if err != nil {
		return err
	}
	body := map[string]any{
		"query_embeddings": resolved.QueryEmbeddings,
	}
	if req.NResults != nil {
		body["n_results"] = *req.NResults
	} else {
		body["n_results"] = 10
	}
	if req.Where != nil {
		body["where"] = req.Where
	}
	if req.WhereDocument != nil {
		body["where_document"] = req.WhereDocument
	}
	if req.Include != nil {
		body["include"] = req.Include
	}
	result, err := h.deps.Chroma.Request(r.Context(), http.MethodPost, fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections/%s/query", urlPath(plateID), urlPath(database), urlPath(collectionID)), body)
	if err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, result)
	return nil
}

func (h *vectorHandler) handleDeleteRecords(w http.ResponseWriter, r *http.Request, plateID string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	collection, err := plate.PathValue(r, "collection")
	if err != nil {
		return err
	}
	var req recordDeleteRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	collectionID, err := h.resolveCollectionID(r.Context(), plateID, database, collection)
	if err != nil {
		return err
	}
	body := map[string]any{}
	if len(req.IDs) > 0 {
		body["ids"] = req.IDs
	}
	if req.Where != nil {
		body["where"] = req.Where
	}
	if req.WhereDocument != nil {
		body["where_document"] = req.WhereDocument
	}
	if req.Limit != nil {
		body["limit"] = *req.Limit
	}
	result, err := h.deps.Chroma.Request(r.Context(), http.MethodPost, fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections/%s/delete", urlPath(plateID), urlPath(database), urlPath(collectionID)), body)
	if err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, result)
	return nil
}

func (h *vectorHandler) handleCountRecords(w http.ResponseWriter, r *http.Request, plateID string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	collection, err := plate.PathValue(r, "collection")
	if err != nil {
		return err
	}
	count, err := h.countRecordsByName(r.Context(), plateID, database, collection)
	if err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"count": count})
	return nil
}

func (h *vectorHandler) handleAliasListCollections(w http.ResponseWriter, r *http.Request, plateID string) error {
	r.SetPathValue("database", defaultDatabase)
	return h.handleListCollections(w, r, plateID)
}

func (h *vectorHandler) handleAliasCreateCollection(w http.ResponseWriter, r *http.Request, plateID string) error {
	r.SetPathValue("database", defaultDatabase)
	return h.handleCreateCollection(w, r, plateID)
}

func (h *vectorHandler) handleAliasGetCollection(w http.ResponseWriter, r *http.Request, plateID string) error {
	r.SetPathValue("database", defaultDatabase)
	return h.handleGetCollection(w, r, plateID)
}

func (h *vectorHandler) handleAliasPatchCollection(w http.ResponseWriter, r *http.Request, plateID string) error {
	r.SetPathValue("database", defaultDatabase)
	return h.handlePatchCollection(w, r, plateID)
}

func (h *vectorHandler) handleAliasDeleteCollection(w http.ResponseWriter, r *http.Request, plateID string) error {
	r.SetPathValue("database", defaultDatabase)
	return h.handleDeleteCollection(w, r, plateID)
}

func (h *vectorHandler) handleAliasGetRecords(w http.ResponseWriter, r *http.Request, plateID string) error {
	r.SetPathValue("database", defaultDatabase)
	return h.handleGetRecords(w, r, plateID)
}

func (h *vectorHandler) handleAliasQueryRecords(w http.ResponseWriter, r *http.Request, plateID string) error {
	r.SetPathValue("database", defaultDatabase)
	return h.handleQueryRecords(w, r, plateID)
}

func (h *vectorHandler) handleAliasUpsertRecords(w http.ResponseWriter, r *http.Request, plateID string) error {
	r.SetPathValue("database", defaultDatabase)
	return h.handleUpsertRecords(w, r, plateID)
}

func (h *vectorHandler) handleDocumentsUpsert(w http.ResponseWriter, r *http.Request, plateID string) error {
	collection, err := plate.PathValue(r, "collection")
	if err != nil {
		return err
	}
	var req documentsUpsertRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	records := make([]recordEntry, 0, len(req.Documents))
	for _, doc := range req.Documents {
		text := doc.Text
		record := recordEntry{ID: doc.ID, Document: &text, Metadata: doc.Metadata, Embedding: doc.Embedding}
		records = append(records, record)
	}
	payload := recordWriteRequest{Records: records, Embedding: req.Embedding}
	body, _ := json.Marshal(payload)
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.ContentLength = int64(len(body))
	r.SetPathValue("database", defaultDatabase)
	r.SetPathValue("collection", collection)
	return h.handleUpsertRecords(w, r, plateID)
}

func (h *vectorHandler) handleDocumentsGet(w http.ResponseWriter, r *http.Request, plateID string) error {
	collection, err := plate.PathValue(r, "collection")
	if err != nil {
		return err
	}
	var req documentsGetRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	payload := recordGetRequest{IDs: req.IDs, Include: []string{"documents", "metadatas", "ids"}}
	body, _ := json.Marshal(payload)
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.ContentLength = int64(len(body))
	r.SetPathValue("database", defaultDatabase)
	r.SetPathValue("collection", collection)
	return h.handleGetRecords(w, r, plateID)
}

func (h *vectorHandler) handleDocumentsSearch(w http.ResponseWriter, r *http.Request, plateID string) error {
	collection, err := plate.PathValue(r, "collection")
	if err != nil {
		return err
	}
	var req documentsSearchRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	query := recordQueryRequest{NResults: req.NResults, Where: req.Where, WhereDocument: req.WhereDocument, Embedding: req.Embedding, Include: []string{"documents", "metadatas", "distances", "ids"}}
	switch typed := req.Query.(type) {
	case string:
		query.QueryTexts = []string{typed}
	case []any:
		vec := make([]float64, 0, len(typed))
		for _, item := range typed {
			if number, ok := item.(float64); ok {
				vec = append(vec, number)
			}
		}
		if len(vec) > 0 {
			query.QueryEmbeddings = [][]float64{vec}
		}
	default:
		return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "query must be string or vector")
	}
	body, _ := json.Marshal(query)
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.ContentLength = int64(len(body))
	r.SetPathValue("database", defaultDatabase)
	r.SetPathValue("collection", collection)
	return h.handleQueryRecords(w, r, plateID)
}

func (h *vectorHandler) handleDocumentsDelete(w http.ResponseWriter, r *http.Request, plateID string) error {
	collection, err := plate.PathValue(r, "collection")
	if err != nil {
		return err
	}
	var req documentsGetRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	payload := recordDeleteRequest{IDs: req.IDs}
	body, _ := json.Marshal(payload)
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.ContentLength = int64(len(body))
	r.SetPathValue("database", defaultDatabase)
	r.SetPathValue("collection", collection)
	return h.handleDeleteRecords(w, r, plateID)
}

func (h *vectorHandler) handleCrossCollectionSearch(w http.ResponseWriter, r *http.Request, plateID string) error {
	var req crossSearchRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	if len(req.Collections) == 0 {
		return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "collections is required")
	}
	nResults := 10
	if req.NResults != nil && *req.NResults > 0 {
		nResults = *req.NResults
	}
	merged := make([]map[string]any, 0)
	for _, collection := range req.Collections {
		q := recordQueryRequest{NResults: &nResults, Where: req.Where, Embedding: req.Embedding, Include: []string{"documents", "metadatas", "distances", "ids"}}
		switch typed := req.Query.(type) {
		case string:
			q.QueryTexts = []string{typed}
		case []any:
			vec := make([]float64, 0, len(typed))
			for _, item := range typed {
				if number, ok := item.(float64); ok {
					vec = append(vec, number)
				}
			}
			if len(vec) > 0 {
				q.QueryEmbeddings = [][]float64{vec}
			}
		default:
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "query must be string or vector")
		}
		body, _ := json.Marshal(q)
		fakeReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, r.URL.Path, io.NopCloser(bytes.NewReader(body)))
		fakeReq.Header = r.Header.Clone()
		fakeReq.SetPathValue("database", defaultDatabase)
		fakeReq.SetPathValue("collection", collection)
		collector := newResponseCollector()
		if err := h.handleQueryRecords(collector, fakeReq, plateID); err != nil {
			continue
		}
		if collector.status >= 400 {
			continue
		}
		var envelope struct {
			OK   bool           `json:"ok"`
			Data map[string]any `json:"data"`
		}
		if unmarshalErr := json.Unmarshal(collector.body.Bytes(), &envelope); unmarshalErr != nil {
			continue
		}
		ids, _ := envelope.Data["ids"].([]any)
		docs, _ := envelope.Data["documents"].([]any)
		meta, _ := envelope.Data["metadatas"].([]any)
		dist, _ := envelope.Data["distances"].([]any)
		if len(ids) > 0 {
			firstIDs, _ := ids[0].([]any)
			firstDocs, _ := docs[0].([]any)
			firstMeta, _ := meta[0].([]any)
			firstDist, _ := dist[0].([]any)
			for i := 0; i < len(firstIDs); i++ {
				entry := map[string]any{"collection": collection, "id": firstIDs[i]}
				if i < len(firstDocs) {
					entry["document"] = firstDocs[i]
				}
				if i < len(firstMeta) {
					entry["metadata"] = firstMeta[i]
				}
				if i < len(firstDist) {
					entry["distance"] = firstDist[i]
				}
				merged = append(merged, entry)
			}
		}
	}
	sort.SliceStable(merged, func(i int, j int) bool {
		di, _ := merged[i]["distance"].(float64)
		dj, _ := merged[j]["distance"].(float64)
		if di == dj {
			a := fmt.Sprintf("%v:%v", merged[i]["collection"], merged[i]["id"])
			b := fmt.Sprintf("%v:%v", merged[j]["collection"], merged[j]["id"])
			return a < b
		}
		return di < dj
	})
	if len(merged) > nResults {
		merged = merged[:nResults]
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"results": merged})
	return nil
}

func (h *vectorHandler) handleListEmbeddingProfiles(w http.ResponseWriter, r *http.Request, plateID string) error {
	profiles, err := h.deps.Meta.ListProfiles(plateID)
	if err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"profiles": profiles})
	return nil
}

func (h *vectorHandler) handleCreateEmbeddingProfile(w http.ResponseWriter, r *http.Request, plateID string) error {
	var req profileRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	profile, secret, err := h.validateProfileRequest(req)
	if err != nil {
		return err
	}
	if saveErr := h.deps.Meta.SaveProfile(plateID, profile, secret); saveErr != nil {
		return saveErr
	}
	plate.WriteOK(w, http.StatusOK, sanitizeProfile(*profile))
	return nil
}

func (h *vectorHandler) handlePatchEmbeddingProfile(w http.ResponseWriter, r *http.Request, plateID string) error {
	profileID, err := plate.PathValue(r, "profileID")
	if err != nil {
		return err
	}
	var req profileRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	existing, err := h.deps.Meta.GetProfile(plateID, profileID)
	if err != nil {
		return err
	}
	if existing == nil {
		return plate.NewAPIError(http.StatusNotFound, "not_found", "embedding profile not found")
	}
	merged := profileRequest{
		Name:       pick(req.Name, existing.Name),
		Provider:   pick(req.Provider, string(existing.Provider)),
		Model:      pick(req.Model, existing.Model),
		Dimensions: existing.Dimensions,
		APIKeyMode: pick(req.APIKeyMode, string(existing.APIKeyMode)),
		APIKey:     req.APIKey,
	}
	if req.Dimensions != nil {
		merged.Dimensions = req.Dimensions
	}
	profile, secret, err := h.validateProfileRequest(merged)
	if err != nil {
		return err
	}
	profile.ID = profileID
	if saveErr := h.deps.Meta.SaveProfile(plateID, profile, secret); saveErr != nil {
		return saveErr
	}
	plate.WriteOK(w, http.StatusOK, sanitizeProfile(*profile))
	return nil
}

func (h *vectorHandler) handleDeleteEmbeddingProfile(w http.ResponseWriter, r *http.Request, plateID string) error {
	profileID, err := plate.PathValue(r, "profileID")
	if err != nil {
		return err
	}
	if err := h.deps.Meta.DeleteProfile(plateID, profileID); err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"deleted": true, "id": profileID})
	return nil
}

func (h *vectorHandler) handleGetEmbeddingDefaults(w http.ResponseWriter, r *http.Request, plateID string) error {
	defaults, err := h.deps.Meta.GetDefaults(plateID)
	if err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, defaults)
	return nil
}

func (h *vectorHandler) handlePutEmbeddingDefaults(w http.ResponseWriter, r *http.Request, plateID string) error {
	var req defaultsRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	if strings.TrimSpace(req.DefaultProfileID) != "" {
		profile, err := h.deps.Meta.GetProfile(plateID, strings.TrimSpace(req.DefaultProfileID))
		if err != nil {
			return err
		}
		if profile == nil {
			return plate.NewAPIError(http.StatusNotFound, "not_found", "default profile not found")
		}
	}
	defaults := plate.EmbeddingDefaults{DefaultProfileID: strings.TrimSpace(req.DefaultProfileID), FallbackToHeaderProvider: req.FallbackToHeaderProvider}
	if err := h.deps.Meta.SetDefaults(plateID, defaults); err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, defaults)
	return nil
}

func (h *vectorHandler) handleEmbeddingsRun(w http.ResponseWriter, r *http.Request, plateID string) error {
	_ = plateID
	var req struct {
		Inputs     []string `json:"inputs"`
		Provider   string   `json:"provider"`
		Model      string   `json:"model"`
		Dimensions *int     `json:"dimensions,omitempty"`
	}
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	if len(req.Inputs) == 0 {
		return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "inputs is required")
	}
	config := embeddingConfig{Provider: req.Provider, Model: req.Model, Dimensions: req.Dimensions}
	embeddings, err := h.generateEmbeddings(r.Context(), r.Header.Get("X-Embedding-Api-Key"), config, req.Inputs)
	if err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"count": len(embeddings), "dimensions": vectorDim(embeddings)})
	return nil
}

func (h *vectorHandler) handleEmbeddingProviders(w http.ResponseWriter, r *http.Request, plateID string) error {
	_ = r
	_ = plateID
	plate.WriteOK(w, http.StatusOK, map[string]any{"providers": []string{"openai", "openrouter"}})
	return nil
}

func (h *vectorHandler) handleEmbeddingModels(w http.ResponseWriter, r *http.Request, plateID string) error {
	_ = plateID
	provider := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("provider")))
	if err := plate.ValidateProvider(provider); err != nil {
		return err
	}
	if provider == "" || provider == "openai" {
		if provider == "" {
			plate.WriteOK(w, http.StatusOK, map[string]any{"models": []string{"text-embedding-3-small", "text-embedding-3-large", "text-embedding-ada-002"}})
			return nil
		}
		models, err := h.fetchOpenAIModels(r.Context(), r.Header.Get("X-Embedding-Api-Key"))
		if err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"models": models})
		return nil
	}
	models, err := h.fetchOpenRouterModels(r.Context())
	if err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"models": models})
	return nil
}

func (h *vectorHandler) handleExport(w http.ResponseWriter, r *http.Request, plateID string) error {
	database := strings.TrimSpace(r.URL.Query().Get("database"))
	if database == "" {
		database = defaultDatabase
	}
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "jsonl"
	}
	if format != "jsonl" && format != "ndjson" {
		return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "format must be jsonl or ndjson")
	}
	collections, err := h.listCollections(r.Context(), plateID, database)
	if err != nil {
		return err
	}
	snapshot := exportSnapshot{Database: database, Collections: make([]exportCollection, 0, len(collections))}
	for _, name := range collections {
		collectionID, idErr := h.resolveCollectionID(r.Context(), plateID, database, name)
		if idErr != nil {
			continue
		}
		data, getErr := h.deps.Chroma.Request(r.Context(), http.MethodPost, fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections/%s/get", urlPath(plateID), urlPath(database), urlPath(collectionID)), map[string]any{"include": []string{"documents", "metadatas", "embeddings", "uris"}})
		if getErr != nil {
			continue
		}
		records := convertGetResponseToRecords(data)
		snapshot.Collections = append(snapshot.Collections, exportCollection{Name: name, Records: records})
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/x-ndjson")
	return writeRawOK(w, encoded)
}

func (h *vectorHandler) handleImport(w http.ResponseWriter, r *http.Request, plateID string) error {
	var req importRequest
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	database := strings.TrimSpace(req.Database)
	if database == "" {
		database = defaultDatabase
	}
	if _, err := h.ensureTenantAndDB(r.Context(), plateID, database); err != nil {
		return err
	}
	for _, collection := range req.Collections {
		createPayload := collectionCreateRequest{Name: collection.Name, Metadata: collection.Metadata, GetOrCreate: true}
		body, _ := json.Marshal(createPayload)
		fakeReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, r.URL.Path, io.NopCloser(bytes.NewReader(body)))
		fakeReq.Header = r.Header.Clone()
		fakeReq.SetPathValue("database", database)
		_ = h.handleCreateCollection(newResponseCollector(), fakeReq, plateID)
		writePayload := recordWriteRequest{Records: collection.Records}
		writeBody, _ := json.Marshal(writePayload)
		writeReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, r.URL.Path, io.NopCloser(bytes.NewReader(writeBody)))
		writeReq.Header = r.Header.Clone()
		writeReq.SetPathValue("database", database)
		writeReq.SetPathValue("collection", collection.Name)
		if err := h.handleUpsertRecords(newResponseCollector(), writeReq, plateID); err != nil {
			return err
		}
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"imported": len(req.Collections)})
	return nil
}

func (h *vectorHandler) handleCloneCollection(w http.ResponseWriter, r *http.Request, plateID string) error {
	return h.copyLikeOperation(w, r, plateID, "clone")
}

func (h *vectorHandler) handleCopyCollection(w http.ResponseWriter, r *http.Request, plateID string) error {
	return h.copyLikeOperation(w, r, plateID, "copy")
}

func (h *vectorHandler) handleReindexCollection(w http.ResponseWriter, r *http.Request, plateID string) error {
	_ = r
	_ = plateID
	plate.WriteOK(w, http.StatusOK, map[string]any{"reindexed": true})
	return nil
}

func (h *vectorHandler) handleCollectionStats(w http.ResponseWriter, r *http.Request, plateID string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	collection, err := plate.PathValue(r, "collection")
	if err != nil {
		return err
	}
	count, err := h.countRecordsByName(r.Context(), plateID, database, collection)
	if err != nil {
		return err
	}
	settings, _ := h.deps.Meta.GetCollectionSettings(plateID, database, collection)
	response := map[string]any{"count": count}
	if settings != nil {
		if settings.Dimension != nil {
			response["dimension"] = *settings.Dimension
		}
		if settings.DistanceMetric != "" {
			response["distance_metric"] = settings.DistanceMetric
		}
	}
	plate.WriteOK(w, http.StatusOK, response)
	return nil
}

func (h *vectorHandler) copyLikeOperation(w http.ResponseWriter, r *http.Request, plateID string, kind string) error {
	database, err := plate.PathValue(r, "database")
	if err != nil {
		return err
	}
	collection, err := plate.PathValue(r, "collection")
	if err != nil {
		return err
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := plate.DecodeJSON(r, &req); err != nil {
		return err
	}
	newName := strings.TrimSpace(req.Name)
	if newName == "" {
		return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "name is required")
	}
	collectionID, err := h.resolveCollectionID(r.Context(), plateID, database, collection)
	if err != nil {
		return err
	}
	source, err := h.deps.Chroma.Request(r.Context(), http.MethodPost, fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections/%s/get", urlPath(plateID), urlPath(database), urlPath(collectionID)), map[string]any{"include": []string{"documents", "metadatas", "embeddings", "uris"}})
	if err != nil {
		return err
	}
	records := convertGetResponseToRecords(source)
	createPayload := collectionCreateRequest{Name: newName, GetOrCreate: false}
	body, _ := json.Marshal(createPayload)
	createReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, r.URL.Path, io.NopCloser(bytes.NewReader(body)))
	createReq.Header = r.Header.Clone()
	createReq.SetPathValue("database", database)
	if err := h.handleCreateCollection(newResponseCollector(), createReq, plateID); err != nil {
		return err
	}
	writePayload := recordWriteRequest{Records: records}
	writeBody, _ := json.Marshal(writePayload)
	writeReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, r.URL.Path, io.NopCloser(bytes.NewReader(writeBody)))
	writeReq.Header = r.Header.Clone()
	writeReq.SetPathValue("database", database)
	writeReq.SetPathValue("collection", newName)
	if err := h.handleUpsertRecords(newResponseCollector(), writeReq, plateID); err != nil {
		return err
	}
	plate.WriteOK(w, http.StatusOK, map[string]any{"operation": kind, "name": newName, "records": len(records)})
	return nil
}

type resolvedWriteEmbedding struct {
	Generated [][]float64
	Dimension int
}

type resolvedQueryEmbedding struct {
	QueryEmbeddings [][]float64
	Dimension       int
}

func (h *vectorHandler) resolveEmbeddingForWrite(ctx context.Context, headerKey string, plateID string, database string, collection string, requestEmbedding *embeddingConfig, records []recordEntry) (resolvedWriteEmbedding, error) {
	hasDirect := true
	texts := make([]string, 0, len(records))
	for _, record := range records {
		if len(record.Embedding) == 0 {
			hasDirect = false
			if record.Document != nil {
				texts = append(texts, *record.Document)
			} else {
				texts = append(texts, "")
			}
		}
	}
	if hasDirect {
		dimension := 0
		for _, record := range records {
			if len(record.Embedding) > 0 {
				dimension = len(record.Embedding)
				break
			}
		}
		return resolvedWriteEmbedding{Generated: nil, Dimension: dimension}, nil
	}
	config, err := h.resolveEmbeddingConfig(plateID, database, collection, requestEmbedding)
	if err != nil {
		return resolvedWriteEmbedding{}, err
	}
	if config == nil {
		return resolvedWriteEmbedding{}, plate.NewAPIError(http.StatusBadRequest, "embedding_not_configured", "embedding provider or direct embeddings required")
	}
	vectors, err := h.generateEmbeddings(ctx, headerKey, *config, texts)
	if err != nil {
		return resolvedWriteEmbedding{}, err
	}
	if len(vectors) != len(texts) {
		return resolvedWriteEmbedding{}, plate.NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", "embedding provider returned unexpected count")
	}
	generated := make([][]float64, 0, len(records))
	textIdx := 0
	for _, record := range records {
		if len(record.Embedding) > 0 {
			generated = append(generated, record.Embedding)
			continue
		}
		generated = append(generated, vectors[textIdx])
		textIdx++
	}
	dimension := vectorDim(generated)
	if config.Dimensions != nil {
		dimension = *config.Dimensions
	}
	return resolvedWriteEmbedding{Generated: generated, Dimension: dimension}, nil
}

func (h *vectorHandler) resolveEmbeddingForQuery(ctx context.Context, headerKey string, plateID string, database string, collection string, requestEmbedding *embeddingConfig, queryTexts []string, queryEmbeddings [][]float64) (resolvedQueryEmbedding, error) {
	if len(queryEmbeddings) > 0 {
		return resolvedQueryEmbedding{QueryEmbeddings: queryEmbeddings, Dimension: vectorDim(queryEmbeddings)}, nil
	}
	if len(queryTexts) == 0 {
		return resolvedQueryEmbedding{}, plate.NewAPIError(http.StatusBadRequest, "invalid_request", "query_texts or query_embeddings is required")
	}
	config, err := h.resolveEmbeddingConfig(plateID, database, collection, requestEmbedding)
	if err != nil {
		return resolvedQueryEmbedding{}, err
	}
	if config == nil {
		return resolvedQueryEmbedding{}, plate.NewAPIError(http.StatusBadRequest, "embedding_not_configured", "embedding provider or query embeddings required")
	}
	vectors, err := h.generateEmbeddings(ctx, headerKey, *config, queryTexts)
	if err != nil {
		return resolvedQueryEmbedding{}, err
	}
	dimension := vectorDim(vectors)
	if config.Dimensions != nil {
		dimension = *config.Dimensions
	}
	return resolvedQueryEmbedding{QueryEmbeddings: vectors, Dimension: dimension}, nil
}

func (h *vectorHandler) resolveEmbeddingConfig(plateID string, database string, collection string, requestEmbedding *embeddingConfig) (*embeddingConfig, error) {
	if requestEmbedding != nil {
		if err := validateEmbeddingConfig(*requestEmbedding); err != nil {
			return nil, err
		}
		copy := *requestEmbedding
		copy.Provider = strings.ToLower(strings.TrimSpace(copy.Provider))
		return &copy, nil
	}
	settings, err := h.deps.Meta.GetCollectionSettings(plateID, database, collection)
	if err != nil {
		return nil, err
	}
	if settings != nil && strings.TrimSpace(settings.EmbeddingProfileID) != "" {
		profile, pErr := h.deps.Meta.GetProfile(plateID, strings.TrimSpace(settings.EmbeddingProfileID))
		if pErr != nil {
			return nil, pErr
		}
		if profile != nil {
			return &embeddingConfig{Provider: string(profile.Provider), Model: profile.Model, Dimensions: profile.Dimensions}, nil
		}
	}
	defaults, err := h.deps.Meta.GetDefaults(plateID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(defaults.DefaultProfileID) != "" {
		profile, pErr := h.deps.Meta.GetProfile(plateID, strings.TrimSpace(defaults.DefaultProfileID))
		if pErr != nil {
			return nil, pErr
		}
		if profile != nil {
			return &embeddingConfig{Provider: string(profile.Provider), Model: profile.Model, Dimensions: profile.Dimensions}, nil
		}
	}
	return nil, nil
}

func validateEmbeddingConfig(config embeddingConfig) error {
	if config.BaseURL != nil || config.Endpoint != nil {
		return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "custom embedding endpoints are not supported")
	}
	if err := plate.ValidateProvider(config.Provider); err != nil {
		return err
	}
	if strings.TrimSpace(config.Provider) == "" || strings.TrimSpace(config.Model) == "" {
		return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "embedding provider and model are required")
	}
	return nil
}

func (h *vectorHandler) validateProfileRequest(req profileRequest) (*plate.EmbeddingProfile, string, error) {
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if err := plate.ValidateProvider(provider); err != nil {
		return nil, "", err
	}
	if provider == "" {
		return nil, "", plate.NewAPIError(http.StatusBadRequest, "invalid_request", "provider is required")
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		return nil, "", plate.NewAPIError(http.StatusBadRequest, "invalid_request", "model is required")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, "", plate.NewAPIError(http.StatusBadRequest, "invalid_request", "name is required")
	}
	mode := strings.ToLower(strings.TrimSpace(req.APIKeyMode))
	if mode == "" {
		mode = string(plate.APIKeyModeHeader)
	}
	if mode != string(plate.APIKeyModeHeader) && mode != string(plate.APIKeyModeStored) {
		return nil, "", plate.NewAPIError(http.StatusBadRequest, "invalid_request", "api_key_mode must be stored or header")
	}
	if mode == string(plate.APIKeyModeStored) && strings.TrimSpace(req.APIKey) == "" {
		return nil, "", plate.NewAPIError(http.StatusBadRequest, "embedding_auth_missing", "api_key required for stored mode")
	}
	profile := &plate.EmbeddingProfile{
		Name:       name,
		Provider:   plate.EmbeddingProvider(provider),
		Model:      model,
		Dimensions: req.Dimensions,
		APIKeyMode: plate.APIKeyMode(mode),
	}
	return profile, strings.TrimSpace(req.APIKey), nil
}

func (h *vectorHandler) generateEmbeddings(ctx context.Context, headerAPIKey string, config embeddingConfig, inputs []string) ([][]float64, error) {
	if err := validateEmbeddingConfig(config); err != nil {
		return nil, err
	}
	provider := strings.ToLower(strings.TrimSpace(config.Provider))
	apiKey := strings.TrimSpace(headerAPIKey)
	if apiKey == "" {
		return nil, plate.NewAPIError(http.StatusUnauthorized, "embedding_auth_missing", "X-Embedding-Api-Key header is required")
	}
	endpoint := ""
	switch provider {
	case "openai":
		endpoint = "https://api.openai.com/v1/embeddings"
	case "openrouter":
		endpoint = "https://openrouter.ai/api/v1/embeddings"
	default:
		return nil, plate.NewAPIError(http.StatusBadRequest, "embedding_provider_unsupported", "unsupported embedding provider")
	}
	body := map[string]any{"input": inputs, "model": config.Model}
	if config.Dimensions != nil {
		body["dimensions"] = *config.Dimensions
	}
	encoded, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return nil, plate.NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if provider == "openrouter" {
		req.Header.Set("HTTP-Referer", "https://smallplate.local")
		req.Header.Set("X-Title", "smallplate-vec")
	}
	resp, err := h.embedClient.Do(req)
	if err != nil {
		return nil, plate.NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", err.Error())
	}
	defer resp.Body.Close()
	content, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, plate.NewAPIError(http.StatusUnauthorized, "embedding_auth_missing", "embedding authorization failed")
		}
		return nil, plate.NewAPIError(http.StatusBadRequest, "embedding_model_invalid", strings.TrimSpace(string(content)))
	}
	var parsed struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(content, &parsed); err != nil {
		return nil, plate.NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", "invalid embedding response")
	}
	vectors := make([][]float64, 0, len(parsed.Data))
	for _, entry := range parsed.Data {
		vectors = append(vectors, entry.Embedding)
	}
	return vectors, nil
}

func (h *vectorHandler) fetchOpenAIModels(ctx context.Context, apiKey string) ([]string, error) {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		return nil, plate.NewAPIError(http.StatusUnauthorized, "embedding_auth_missing", "X-Embedding-Api-Key header is required")
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.openai.com/v1/models", nil)
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := h.embedClient.Do(req)
	if err != nil {
		return nil, plate.NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", err.Error())
	}
	defer resp.Body.Close()
	content, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, plate.NewAPIError(http.StatusBadRequest, "embedding_model_invalid", strings.TrimSpace(string(content)))
	}
	var parsed struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(content, &parsed); err != nil {
		return nil, plate.NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", "invalid model response")
	}
	models := make([]string, 0)
	for _, model := range parsed.Data {
		if strings.Contains(strings.ToLower(model.ID), "embedding") {
			models = append(models, model.ID)
		}
	}
	sort.Strings(models)
	return models, nil
}

func (h *vectorHandler) fetchOpenRouterModels(ctx context.Context) ([]string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://openrouter.ai/api/v1/embeddings/models", nil)
	resp, err := h.embedClient.Do(req)
	if err != nil {
		return nil, plate.NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", err.Error())
	}
	defer resp.Body.Close()
	content, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, plate.NewAPIError(http.StatusBadRequest, "embedding_model_invalid", strings.TrimSpace(string(content)))
	}
	var parsed struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(content, &parsed); err != nil {
		return nil, plate.NewAPIError(http.StatusServiceUnavailable, "backend_unavailable", "invalid model response")
	}
	models := make([]string, 0, len(parsed.Data))
	for _, model := range parsed.Data {
		models = append(models, model.ID)
	}
	sort.Strings(models)
	return models, nil
}

func (h *vectorHandler) ensureTenantAndDB(ctx context.Context, plateID string, database string) (bool, error) {
	_, _ = h.deps.Chroma.Request(ctx, http.MethodPost, "/api/v2/tenants", map[string]any{"name": plateID})
	_, _ = h.deps.Chroma.Request(ctx, http.MethodPost, fmt.Sprintf("/api/v2/tenants/%s/databases", urlPath(plateID)), map[string]any{"name": database})
	return true, nil
}

func (h *vectorHandler) listDatabases(ctx context.Context, plateID string) ([]string, error) {
	response, err := h.deps.Chroma.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/tenants/%s/databases", urlPath(plateID)), nil)
	if err != nil {
		return nil, err
	}
	items, _ := response["databases"].([]any)
	result := make([]string, 0, len(items))
	for _, item := range items {
		mapped, _ := item.(map[string]any)
		if name := plate.AsString(mapped["name"]); name != "" {
			result = append(result, name)
		}
	}
	if len(result) == 0 {
		result = append(result, defaultDatabase)
	}
	sort.Strings(result)
	return result, nil
}

func (h *vectorHandler) listCollectionsDetailed(ctx context.Context, plateID string, database string) ([]map[string]any, error) {
	response, err := h.deps.Chroma.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections", urlPath(plateID), urlPath(database)), nil)
	if err != nil {
		return nil, err
	}
	items, _ := response["collections"].([]any)
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		mapped, _ := item.(map[string]any)
		name := plate.AsString(mapped["name"])
		if name == "" {
			continue
		}
		entry := map[string]any{"name": name, "id": plate.AsString(mapped["id"])}
		if metadata, ok := mapped["metadata"].(map[string]any); ok {
			entry["metadata"] = metadata
		}
		settings, _ := h.deps.Meta.GetCollectionSettings(plateID, database, name)
		if settings != nil {
			if settings.Dimension != nil {
				entry["dimension"] = *settings.Dimension
			}
			if settings.DistanceMetric != "" {
				entry["distance_metric"] = settings.DistanceMetric
			}
			if settings.EmbeddingProfileID != "" {
				entry["embedding_profile_id"] = settings.EmbeddingProfileID
			}
		}
		result = append(result, entry)
	}
	sort.Slice(result, func(i int, j int) bool {
		return fmt.Sprintf("%v", result[i]["name"]) < fmt.Sprintf("%v", result[j]["name"])
	})
	return result, nil
}

func (h *vectorHandler) listCollections(ctx context.Context, plateID string, database string) ([]string, error) {
	detailed, err := h.listCollectionsDetailed(ctx, plateID, database)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(detailed))
	for _, item := range detailed {
		if name := plate.AsString(item["name"]); name != "" {
			result = append(result, name)
		}
	}
	return result, nil
}

func (h *vectorHandler) deleteCollectionByName(ctx context.Context, plateID string, database string, collection string) error {
	_, err := h.deps.Chroma.Request(ctx, http.MethodDelete, fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections/%s", urlPath(plateID), urlPath(database), urlPath(collection)), nil)
	if err != nil {
		return err
	}
	_ = h.deps.Meta.DeleteCollectionSettings(plateID, database, collection)
	return nil
}

func (h *vectorHandler) resolveCollectionID(ctx context.Context, plateID string, database string, collection string) (string, error) {
	response, err := h.deps.Chroma.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections/%s", urlPath(plateID), urlPath(database), urlPath(collection)), nil)
	if err != nil {
		return "", err
	}
	id := plate.AsString(response["id"])
	if id == "" {
		return "", plate.NewAPIError(http.StatusNotFound, "not_found", "collection id not found")
	}
	return id, nil
}

func (h *vectorHandler) countRecordsByName(ctx context.Context, plateID string, database string, collection string) (int, error) {
	collectionID, err := h.resolveCollectionID(ctx, plateID, database, collection)
	if err != nil {
		return 0, err
	}
	response, err := h.deps.Chroma.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections/%s/count", urlPath(plateID), urlPath(database), urlPath(collectionID)), nil)
	if err != nil {
		return 0, err
	}
	if count := plate.AsInt(response["count"]); count > 0 {
		return count, nil
	}
	if count := plate.AsInt(response["n"]); count > 0 {
		return count, nil
	}
	return 0, nil
}

func convertGetResponseToRecords(payload map[string]any) []recordEntry {
	idsByQuery, _ := payload["ids"].([]any)
	docsByQuery, _ := payload["documents"].([]any)
	metaByQuery, _ := payload["metadatas"].([]any)
	vecByQuery, _ := payload["embeddings"].([]any)
	uriByQuery, _ := payload["uris"].([]any)
	if len(idsByQuery) == 0 {
		return nil
	}
	ids, _ := idsByQuery[0].([]any)
	docs := safeAnySlice(docsByQuery, 0)
	metas := safeAnySlice(metaByQuery, 0)
	vecs := safeAnySlice(vecByQuery, 0)
	uris := safeAnySlice(uriByQuery, 0)
	records := make([]recordEntry, 0, len(ids))
	for idx := range ids {
		record := recordEntry{ID: fmt.Sprintf("%v", ids[idx])}
		if idx < len(docs) {
			text := fmt.Sprintf("%v", docs[idx])
			record.Document = &text
		}
		if idx < len(metas) {
			if mapped, ok := metas[idx].(map[string]any); ok {
				record.Metadata = mapped
			}
		}
		if idx < len(vecs) {
			if raw, ok := vecs[idx].([]any); ok {
				vector := make([]float64, 0, len(raw))
				for _, entry := range raw {
					if value, ok := entry.(float64); ok {
						vector = append(vector, value)
					}
				}
				record.Embedding = vector
			}
		}
		if idx < len(uris) {
			value := fmt.Sprintf("%v", uris[idx])
			record.URI = &value
		}
		records = append(records, record)
	}
	return records
}

func safeAnySlice(items []any, index int) []any {
	if index >= len(items) {
		return nil
	}
	result, _ := items[index].([]any)
	return result
}

func vectorDim(vectors [][]float64) int {
	for _, vector := range vectors {
		if len(vector) > 0 {
			return len(vector)
		}
	}
	return 0
}

func urlPath(value string) string {
	replacer := strings.NewReplacer(" ", "%20")
	return replacer.Replace(strings.TrimSpace(value))
}

func pick(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func sanitizeProfile(profile plate.EmbeddingProfile) map[string]any {
	result := map[string]any{
		"id":           profile.ID,
		"name":         profile.Name,
		"provider":     profile.Provider,
		"model":        profile.Model,
		"api_key_mode": profile.APIKeyMode,
		"created_at":   profile.CreatedAt,
		"updated_at":   profile.UpdatedAt,
	}
	if profile.Dimensions != nil {
		result["dimensions"] = *profile.Dimensions
	}
	return result
}

func writeRawOK(w http.ResponseWriter, payload []byte) error {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(payload)
	return err
}

type responseCollector struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func newResponseCollector() *responseCollector {
	return &responseCollector{header: make(http.Header), status: http.StatusOK}
}

func (r *responseCollector) Header() http.Header {
	return r.header
}

func (r *responseCollector) Write(payload []byte) (int, error) {
	return r.body.Write(payload)
}

func (r *responseCollector) WriteHeader(statusCode int) {
	r.status = statusCode
}

func (h *vectorHandler) handleHeartbeatBackends(ctx context.Context) map[string]any {
	status := map[string]any{"chroma": "ok"}
	if err := h.deps.Chroma.Heartbeat(ctx); err != nil {
		status["chroma"] = err.Error()
	}
	status["ts"] = time.Now().UTC().Format(time.RFC3339Nano)
	return status
}
