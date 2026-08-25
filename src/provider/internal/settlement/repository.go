package settlement

import (
	"context"
	"encoding/json"

	"github.com/quanttide/qtcloud-crowd-provider/internal/store"
)

// Repository 是结算仓储：从数据集文档中读写结算记录。
type Repository struct {
	st  store.Store
	key string
}

// NewRepository 创建结算仓储。
func NewRepository(st store.Store, key string) *Repository {
	return &Repository{st: st, key: key}
}

// List 返回全部结算记录（不存在时为空切片）。
func (r *Repository) List(ctx context.Context) ([]Settlement, error) {
	ds, err := store.Load(ctx, r.st, r.key)
	if err != nil {
		return nil, err
	}
	items := make([]Settlement, 0, len(ds.Settlements))
	for _, raw := range ds.Settlements {
		var s Settlement
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, nil
}

// Create 新增一笔结算记录（同 id 覆盖防重复记账）。
func (r *Repository) Create(ctx context.Context, s Settlement) error {
	ds, err := store.Load(ctx, r.st, r.key)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	for i, existing := range ds.Settlements {
		var cur Settlement
		if err := json.Unmarshal(existing, &cur); err != nil {
			return err
		}
		if cur.ID == s.ID {
			ds.Settlements[i] = raw
			return ds.Save(ctx, r.st, r.key)
		}
	}
	ds.Settlements = append(ds.Settlements, raw)
	return ds.Save(ctx, r.st, r.key)
}
