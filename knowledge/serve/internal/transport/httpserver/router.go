package httpserver

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/butbeautifulv/veil/knowledge/serve/internal/domain"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/usecase"
	"github.com/butbeautifulv/veil/pkg/api"
)

// SetProdMode toggles generic API error messages (no internal details).
func SetProdMode(prod bool) {
	api.SetProdMode(prod)
}

func Register(mux *http.ServeMux, uc *usecase.ReadUsecase) {
	api.RegisterHealth(mux, "veil-api", nil)
	mux.HandleFunc("GET /v1/categories", func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, map[string]any{"categories": uc.ListCategoryMeta()})
	})
	mux.HandleFunc("GET /v1/categories/{category}/kinds", func(w http.ResponseWriter, r *http.Request) {
		cat := r.PathValue("category")
		kinds, err := uc.ListKindsInCategory(r.Context(), cat)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"category": cat, "kinds": kinds})
	})
	mux.HandleFunc("GET /v1/categories/{category}/nodes", func(w http.ResponseWriter, r *http.Request) {
		cat := r.PathValue("category")
		kind := r.URL.Query().Get("kind")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		nodes, err := uc.NodesByCategory(r.Context(), cat, kind, limit, offset)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"category": cat, "kind": kind, "nodes": nodes})
	})
	mux.HandleFunc("GET /v1/categories/{category}/search", func(w http.ResponseWriter, r *http.Request) {
		cat := r.PathValue("category")
		q := r.URL.Query().Get("q")
		kind := r.URL.Query().Get("kind")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		nodes, err := uc.SearchInCategory(r.Context(), cat, q, kind, limit)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"category": cat, "query": q, "kind": kind, "nodes": nodes})
	})
	mux.HandleFunc("GET /v1/categories/engage/context", func(w http.ResponseWriter, r *http.Request) {
		host := r.URL.Query().Get("q")
		if host == "" {
			host = r.URL.Query().Get("host")
		}
		ctx, err := uc.EngageTargetContext(r.Context(), host)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		if ctx == nil {
			api.WriteJSON(w, http.StatusOK, map[string]any{"host": host, "found": false})
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"host": host, "found": ctx.Target != nil, "context": ctx})
	})
	mux.HandleFunc("GET /v1/nodes/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		n, err := uc.GetNodeForAPI(r.Context(), id)
		if err != nil {
			if errors.Is(err, domain.ErrNodeNotFound) {
				writeErr(w, http.StatusNotFound, err)
				return
			}
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"node": n})
	})
	mux.HandleFunc("GET /v1/nodes/{id}/neighbors", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		depth, _ := strconv.Atoi(r.URL.Query().Get("depth"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		g, err := uc.Neighbors(r.Context(), id, depth, limit)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"graph": g})
	})
	mux.HandleFunc("GET /v1/kinds", func(w http.ResponseWriter, r *http.Request) {
		kinds, err := uc.ListKinds(r.Context())
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"kinds": kinds})
	})
}

func writeErr(w http.ResponseWriter, status int, err error) {
	api.WriteError(w, status, err)
}
