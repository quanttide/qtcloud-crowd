package settlement

import (
	"encoding/json"
	"log"
	"net/http"
)

// Handler 暴露结算 REST API：
//
//	GET  /api/settlements  结算台账
//	POST /api/settlements  验收通过记一笔
type Handler struct {
	repo *Repository
}

// NewHandler 创建结算 HTTP handler。
func NewHandler(repo *Repository) http.Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.List(r.Context())
	if err != nil {
		log.Printf("settlement list: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var s Settlement
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if s.ID == "" || s.TaskID == "" || s.PartnerID == "" {
		http.Error(w, "id, task_id and partner_id required", http.StatusBadRequest)
		return
	}
	if s.Amount <= 0 {
		http.Error(w, "amount must be positive", http.StatusBadRequest)
		return
	}
	if err := h.repo.Create(r.Context(), s); err != nil {
		log.Printf("settlement create: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}
