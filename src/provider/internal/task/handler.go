package task

import (
	"encoding/json"
	"log"
	"net/http"
)

// Handler 暴露任务 REST API：
//
//	GET /api/tasks  任务列表
//	PUT /api/tasks  保存单个任务（id 相同则覆盖）
type Handler struct {
	repo *Repository
}

// NewHandler 创建任务 HTTP handler。
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
	tasks, err := h.repo.List(r.Context())
	if err != nil {
		log.Printf("task list: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) upsert(w http.ResponseWriter, r *http.Request) {
	var t Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if t.ID == "" {
		http.Error(w, "task id required", http.StatusBadRequest)
		return
	}
	// 验收准则兜底：说不清验收不能发布（进入 reviewing）。
	if t.Status == StatusReviewing && !t.CanPublish() {
		http.Error(w, "acceptance criteria required before publish", http.StatusBadRequest)
		return
	}
	if err := h.repo.Upsert(r.Context(), t); err != nil {
		log.Printf("task upsert: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}
