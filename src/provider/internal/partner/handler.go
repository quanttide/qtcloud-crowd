package partner

import (
	"encoding/json"
	"log"
	"net/http"
)

// Handler 暴露执行方 REST API：
//
//	GET /api/partners  执行方名单
//	PUT /api/partners  保存单个执行方（含认证动作：certified=true）
type Handler struct {
	repo *Repository
}

// NewHandler 创建执行方 HTTP handler。
func NewHandler(repo *Repository) http.Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPut:
		h.upsert(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	partners, err := h.repo.List(r.Context())
	if err != nil {
		log.Printf("partner list: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, partners)
}

func (h *Handler) upsert(w http.ResponseWriter, r *http.Request) {
	var p Partner
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if p.ID == "" || p.Name == "" {
		http.Error(w, "partner id and name required", http.StatusBadRequest)
		return
	}
	if !validType(p.Type) {
		http.Error(w, "partner type must be channel/agent/training", http.StatusBadRequest)
		return
	}
	if err := h.repo.Upsert(r.Context(), p); err != nil {
		log.Printf("partner upsert: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func validType(t Type) bool {
	switch t {
	case TypeChannel, TypeAgent, TypeTraining:
		return true
	default:
		return false
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}
