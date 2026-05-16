package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/butbeautifulv/veil/graph/serve/internal/domain"
	"github.com/butbeautifulv/veil/graph/serve/internal/usecase"
)

var prodMode atomic.Bool

// SetProdMode toggles generic API error messages (no internal details).
func SetProdMode(prod bool) {
	prodMode.Store(prod)
}

func Register(mux *http.ServeMux, uc *usecase.ReadUsecase) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "service": "veil-api"})
	})
	mux.HandleFunc("GET /v1/categories", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"categories": uc.ListCategoryMeta()})
	})
	mux.HandleFunc("GET /v1/categories/{category}/kinds", func(w http.ResponseWriter, r *http.Request) {
		cat := r.PathValue("category")
		kinds, err := uc.ListKindsInCategory(r.Context(), cat)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"category": cat, "kinds": kinds})
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
		writeJSON(w, http.StatusOK, map[string]any{"category": cat, "kind": kind, "nodes": nodes})
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
		writeJSON(w, http.StatusOK, map[string]any{"category": cat, "query": q, "kind": kind, "nodes": nodes})
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
			writeJSON(w, http.StatusOK, map[string]any{"host": host, "found": false})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"host": host, "found": ctx.Target != nil, "context": ctx})
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
		writeJSON(w, http.StatusOK, map[string]any{"node": n})
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
		writeJSON(w, http.StatusOK, map[string]any{"graph": g})
	})
	mux.HandleFunc("GET /v1/kinds", func(w http.ResponseWriter, r *http.Request) {
		kinds, err := uc.ListKinds(r.Context())
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"kinds": kinds})
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, err error) {
	msg := err.Error()
	if prodMode.Load() {
		msg = "bad request"
		if status == http.StatusNotFound {
			msg = "not found"
		}
		if status >= 500 {
			msg = "internal error"
		}
	}
	writeJSON(w, status, map[string]any{"error": msg})
}
