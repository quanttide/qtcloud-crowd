package task

import (
	"context"
	"encoding/json"

	"github.com/quanttide/qtcloud-crowd-provider/internal/store"
)

// Repository 是任务仓储：从数据集文档中读写任务。
type Repository struct {
	st  store.Store
	key string
}

// NewRepository 创建任务仓储。
func NewRepository(st store.Store, key string) *Repository {
	return &Repository{st: st, key: key}
}

// List 返回全部任务（不存在时为空切片）。
func (r *Repository) List(ctx context.Context) ([]Task, error) {
	ds, err := store.Load(ctx, r.st, r.key)
	if err != nil {
		return nil, err
	}
	tasks := make([]Task, 0, len(ds.Tasks))
	for _, raw := range ds.Tasks {
		var t Task
		if err := json.Unmarshal(raw, &t); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// Upsert 保存任务：同 id 覆盖，否则追加。
func (r *Repository) Upsert(ctx context.Context, t Task) error {
	ds, err := store.Load(ctx, r.st, r.key)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(t)
	if err != nil {
		return err
	}
	for i, existing := range ds.Tasks {
		var cur Task
		if err := json.Unmarshal(existing, &cur); err != nil {
			return err
		}
		if cur.ID == t.ID {
			ds.Tasks[i] = raw
			return ds.Save(ctx, r.st, r.key)
		}
	}
	ds.Tasks = append(ds.Tasks, raw)
	return ds.Save(ctx, r.st, r.key)
}
