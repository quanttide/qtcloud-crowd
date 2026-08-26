package task

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/quanttide/qtcloud-crowd-provider/internal/store"
)

// Handler 暴露任务 REST API：
//
//	GET    /api/tasks               任务列表（?status=published 供前台拉取上架）
//	PUT    /api/tasks               保存单个任务（id 相同则覆盖；status 走状态机）
//	POST   /api/tasks/{id}/claim    写回 API：认领（published→accepted，body: partner_id）
//	POST   /api/tasks/{id}/deliver  写回 API：交付（accepted→reviewing）
type Handler struct {
	repo *Repository
}

// NewHandler 创建任务 HTTP handler。
func NewHandler(repo *Repository) http.Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/tasks":
		h.list(w, r)
	case r.Method == http.MethodPut && r.URL.Path == "/api/tasks":
		h.upsert(w, r)
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/tasks/"):
		h.writeBack(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// list 返回任务列表；支持 ?status= 过滤（如 ?status=published——前台 qtcrowd-provider
// 上架时拉取可上架任务）。status 为空返回全部；非法 status 返回 400。
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	status := Status(r.URL.Query().Get("status"))
	if status != "" && !isValidStatus(status) {
		http.Error(w, "invalid status: "+string(status), http.StatusBadRequest)
		return
	}

	tasks, err := h.repo.List(r.Context())
	if err != nil {
		log.Printf("task list: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if status != "" {
		filtered := make([]Task, 0, len(tasks))
		for _, t := range tasks {
			if t.Status == status {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}
	writeJSON(w, http.StatusOK, tasks)
}

// isValidStatus 判断是否为已知任务状态（pending/reviewing/published/accepted/done）。
func isValidStatus(s Status) bool {
	switch s {
	case StatusPending, StatusReviewing, StatusPublished, StatusAccepted, StatusDone:
		return true
	default:
		return false
	}
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
	// 验收准则兜底：说不清验收不能发布（进入 reviewing 或 published 都不行）。
	if (t.Status == StatusReviewing || t.Status == StatusPublished) && !t.CanPublish() {
		http.Error(w, "acceptance criteria required", http.StatusBadRequest)
		return
	}

	if err := h.repo.Upsert(r.Context(), t); err != nil {
		log.Printf("task upsert: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// writeBack 处理前台写回：POST /api/tasks/{id}/claim 与 /deliver。
// 前台（qtcrowd）依赖后台做写操作，本 API 是唯一入口。
func (h *Handler) writeBack(w http.ResponseWriter, r *http.Request) {
	id, action, ok := parseWriteBackPath(r.URL.Path)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	switch action {
	case "claim":
		h.claim(w, r, id)
	case "deliver":
		h.deliver(w, r, id)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

// claim 认领：published→accepted，body 带 partner_id。
func (h *Handler) claim(w http.ResponseWriter, r *http.Request, id string) {
	t, ok := h.getForWriteBack(w, r, id)
	if !ok {
		return
	}
	if t.Status != StatusPublished {
		http.Error(w, "task not published: only published tasks can be claimed", http.StatusBadRequest)
		return
	}
	var body struct {
		PartnerID string `json:"partner_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if body.PartnerID == "" {
		http.Error(w, "partner_id required", http.StatusBadRequest)
		return
	}

	t.Status = StatusAccepted
	t.PartnerID = body.PartnerID
	if err := h.repo.Upsert(r.Context(), t); err != nil {
		log.Printf("task claim upsert: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// deliver 交付：accepted→reviewing（待验收）。
func (h *Handler) deliver(w http.ResponseWriter, r *http.Request, id string) {
	t, ok := h.getForWriteBack(w, r, id)
	if !ok {
		return
	}
	if t.Status != StatusAccepted {
		http.Error(w, "task not accepted: only accepted tasks can deliver", http.StatusBadRequest)
		return
	}

	t.Status = StatusReviewing
	if err := h.repo.Upsert(r.Context(), t); err != nil {
		log.Printf("task deliver upsert: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// getForWriteBack 读取写回目标任务：未知任务 404，读失败 500。
func (h *Handler) getForWriteBack(w http.ResponseWriter, r *http.Request, id string) (Task, bool) {
	t, err := h.repo.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return Task{}, false
		}
		log.Printf("task get: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return Task{}, false
	}
	return t, true
}

// parseWriteBackPath 解析 POST /api/tasks/{id}/{action}，返回 id 与动作。
func parseWriteBackPath(path string) (id, action string, ok bool) {
	rest := strings.TrimPrefix(path, "/api/tasks/")
	i := strings.LastIndex(rest, "/")
	if i <= 0 || i == len(rest)-1 {
		return "", "", false
	}
	return rest[:i], rest[i+1:], true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}
