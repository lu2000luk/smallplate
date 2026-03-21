package routes

import (
	"net/http"
	"strings"

	"plate/link/internal/plate"
)

func Register(mux *http.ServeMux, deps *plate.Dependencies) {
	create := plate.Authenticated(deps, handleCreateLink(deps, false))
	createDynamic := plate.Authenticated(deps, handleCreateLink(deps, true))
	listLinks := plate.Authenticated(deps, handleListLinks(deps))
	getLink := plate.Authenticated(deps, handleGetLink(deps))
	updateLink := plate.Authenticated(deps, handleUpdateLink(deps))
	updateMetadata := plate.Authenticated(deps, handleUpdateMetadata(deps))
	deleteLink := plate.Authenticated(deps, handleDeleteLink(deps))
	resolveCORS := handleResolveCORS()
	resolveJSON := handleResolveJSON(deps)
	redirect := handleRedirect(deps)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		segments := splitPath(r.URL.Path)
		if len(segments) == 0 {
			plate.WriteError(w, plate.NewAPIError(http.StatusNotFound, "not_found", "endpoint not found"))
			return
		}

		if segments[0] == "url" {
			if r.Method != http.MethodGet {
				plate.WriteError(w, plate.NewAPIError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed"))
				return
			}
			if len(segments) < 2 {
				plate.WriteError(w, plate.NewAPIError(http.StatusBadRequest, "missing_path_value", "missing path value \"id\""))
				return
			}
			r.SetPathValue("id", segments[1])
			if len(segments) > 2 {
				r.SetPathValue("tail", strings.Join(segments[2:], "/"))
			}
			redirect(w, r)
			return
		}

		plateID := segments[0]
		r.SetPathValue("plateID", plateID)

		if len(segments) >= 3 && segments[1] == "resolve" {
			if r.Method != http.MethodGet && r.Method != http.MethodOptions {
				plate.WriteError(w, plate.NewAPIError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed"))
				return
			}
			r.SetPathValue("id", segments[2])
			if len(segments) > 3 {
				r.SetPathValue("tail", strings.Join(segments[3:], "/"))
			}
			if r.Method == http.MethodOptions {
				resolveCORS(w, r)
				return
			}
			resolveJSON(w, r)
			return
		}

		if len(segments) == 2 && segments[1] == "create" && r.Method == http.MethodPost {
			create(w, r)
			return
		}

		if len(segments) == 3 && segments[1] == "create" && segments[2] == "dynamic" && r.Method == http.MethodPost {
			createDynamic(w, r)
			return
		}

		if len(segments) == 2 && segments[1] == "links" && r.Method == http.MethodGet {
			listLinks(w, r)
			return
		}

		if len(segments) >= 3 && segments[1] == "links" {
			r.SetPathValue("id", segments[2])
			switch {
			case len(segments) == 3 && r.Method == http.MethodGet:
				getLink(w, r)
				return
			case len(segments) == 4 && segments[3] == "update" && r.Method == http.MethodPost:
				updateLink(w, r)
				return
			case len(segments) == 4 && segments[3] == "metadata" && r.Method == http.MethodPost:
				updateMetadata(w, r)
				return
			case len(segments) == 3 && r.Method == http.MethodDelete:
				deleteLink(w, r)
				return
			}
		}

		plate.WriteError(w, plate.NewAPIError(http.StatusNotFound, "not_found", "endpoint not found"))
	})
}

func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, "/")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
