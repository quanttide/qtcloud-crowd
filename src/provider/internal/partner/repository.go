package partner

import (
	"context"
	"encoding/json"

	"github.com/quanttide/qtcloud-crowd-provider/internal/store"
)

// Repository 是执行方仓储：从数据集文档中读写执行方。
type Repository struct {
	st  store.Store
	key string
}

// NewRepository 创建执行方仓储。
func NewRepository(st store.Store, key string) *Repository {
	return &Repository{st: st, key: key}
}

// List 返回全部执行方（不存在时为空切片）。
func (r *Repository) List(ctx context.Context) ([]Partner, error) {
	ds, err := store.Load(ctx, r.st, r.key)
	if err != nil {
		return nil, err
	}
	partners := make([]Partner, 0, len(ds.Partners))
	for _, raw := range ds.Partners {
		var p Partner
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, err
		}
		partners = append(partners, p)
	}
	return partners, nil
}

// Upsert 保存执行方：同 id 覆盖，否则追加。
func (r *Repository) Upsert(ctx context.Context, p Partner) error {
	ds, err := store.Load(ctx, r.st, r.key)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	for i, existing := range ds.Partners {
		var cur Partner
		if err := json.Unmarshal(existing, &cur); err != nil {
			return err
		}
		if cur.ID == p.ID {
			ds.Partners[i] = raw
			return ds.Save(ctx, r.st, r.key)
		}
	}
	ds.Partners = append(ds.Partners, raw)
	return ds.Save(ctx, r.st, r.key)
}
