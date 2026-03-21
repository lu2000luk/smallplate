package routes

import (
	"net/http"

	"plate/link/internal/plate"
)

type createLinkRequest struct {
	Destination string         `json:"destination"`
	Template    string         `json:"template"`
	ExpiresAt   int64          `json:"expires_at"`
	MaxUses     int64          `json:"max_uses"`
	Metadata    map[string]any `json:"metadata"`
	IDPrefix    string         `json:"id_prefix"`
}

type updateLinkRequest struct {
	Destination *string         `json:"destination"`
	Template    *string         `json:"template"`
	ExpiresAt   *int64          `json:"expires_at"`
	MaxUses     *int64          `json:"max_uses"`
	Enabled     *bool           `json:"enabled"`
	Metadata    *map[string]any `json:"metadata"`
}

type updateMetadataRequest struct {
	Metadata map[string]any `json:"metadata"`
}

func handleCreateLink(deps *plate.Dependencies, forceDynamic bool) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var req createLinkRequest
		if err := plate.DecodeJSON(r, &req); err != nil {
			return err
		}
		if req.MaxUses < 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_max_uses", "max_uses must be >= 0")
		}
		if req.ExpiresAt < 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_expiry", "expires_at must be >= 0")
		}

		if forceDynamic {
			if _, err := requiredString(req.Template, "template"); err != nil {
				return err
			}
		}

		record, err := deps.Links.Create(r.Context(), plate.CreateLinkInput{
			PlateID:     plateID,
			Destination: req.Destination,
			Template:    req.Template,
			ExpiresAt:   req.ExpiresAt,
			MaxUses:     req.MaxUses,
			Metadata:    req.Metadata,
			IDPrefix:    req.IDPrefix,
		})
		if err != nil {
			return err
		}

		plate.WriteOK(w, http.StatusCreated, map[string]any{
			"link":     record,
			"shortUrl": "/url/" + record.ID,
		})
		return nil
	}
}

func handleListLinks(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, plateID string) error {
		links, err := deps.Links.List(r.Context(), plateID)
		if err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"links": links})
		return nil
	}
}

func handleGetLink(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, plateID string) error {
		id, err := plate.PathValue(r, "id")
		if err != nil {
			return err
		}
		record, err := deps.Links.GetByPlateAndID(r.Context(), plateID, id)
		if err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"link": record})
		return nil
	}
}

func handleUpdateLink(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, plateID string) error {
		id, err := plate.PathValue(r, "id")
		if err != nil {
			return err
		}

		var req updateLinkRequest
		if err := plate.DecodeJSON(r, &req); err != nil {
			return err
		}
		if err := optionalInt64(req.ExpiresAt, "expires_at"); err != nil {
			return err
		}
		if err := optionalInt64(req.MaxUses, "max_uses"); err != nil {
			return err
		}

		record, err := deps.Links.Update(r.Context(), plateID, id, plate.UpdateLinkInput{
			Destination: req.Destination,
			Template:    req.Template,
			ExpiresAt:   req.ExpiresAt,
			MaxUses:     req.MaxUses,
			Enabled:     req.Enabled,
			Metadata:    req.Metadata,
		})
		if err != nil {
			return err
		}

		plate.WriteOK(w, http.StatusOK, map[string]any{"link": record})
		return nil
	}
}

func handleUpdateMetadata(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, plateID string) error {
		id, err := plate.PathValue(r, "id")
		if err != nil {
			return err
		}

		var req updateMetadataRequest
		if err := plate.DecodeJSON(r, &req); err != nil {
			return err
		}

		record, err := deps.Links.Update(r.Context(), plateID, id, plate.UpdateLinkInput{Metadata: &req.Metadata})
		if err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"link": record})
		return nil
	}
}

func handleDeleteLink(deps *plate.Dependencies) func(http.ResponseWriter, *http.Request, string) error {
	return func(w http.ResponseWriter, r *http.Request, plateID string) error {
		id, err := plate.PathValue(r, "id")
		if err != nil {
			return err
		}
		if err := deps.Links.Delete(r.Context(), plateID, id); err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"deleted": true, "id": id})
		return nil
	}
}
