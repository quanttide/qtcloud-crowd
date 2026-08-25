package store

import (
	"context"
	"encoding/json"
	"errors"
)

// Dataset 是整个数据集文档（data/crowd.json 的结构）：
// 任务 / 执行方 / 结算三个数组共处一个对象，各资源包只读写自己的部分。
type Dataset struct {
	Tasks       []json.RawMessage `json:"tasks"`
	Partners    []json.RawMessage `json:"partners"`
	Settlements []json.RawMessage `json:"settlements"`
}

// Load 从存储读取数据集；数据不存在时返回空数据集（首次启动）。
func Load(ctx context.Context, st Store, key string) (*Dataset, error) {
	data, err := st.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return &Dataset{}, nil
		}
		return nil, err
	}
	var ds Dataset
	if err := json.Unmarshal(data, &ds); err != nil {
		return nil, err
	}
	return &ds, nil
}

// Save 把数据集整体写回存储。
func (d *Dataset) Save(ctx context.Context, st Store, key string) error {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return st.Put(ctx, key, data)
}
