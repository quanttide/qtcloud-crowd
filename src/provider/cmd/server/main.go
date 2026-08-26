// qtcloud-crowd-provider：量潮众包管理云服务端。
//
// 环境变量：
//   - QTCLOUD_CROWD_ADDR   监听地址，默认 :8080
//   - QTCLOUD_CROWD_DATA   数据文件路径（OSS 下为对象名），默认 data/crowd.json
//   - QTCLOUD_CROWD_STORE  存储后端，默认 local；设为 oss 走阿里云 OSS
//   - QTCLOUD_OSS_ENDPOINT / QTCLOUD_OSS_BUCKET /
//     QTCLOUD_OSS_PUBLIC_BUCKET / QTCLOUD_OSS_ACCESS_KEY_ID /
//     QTCLOUD_OSS_ACCESS_KEY_SECRET  OSS 配置（PUBLIC_BUCKET=公开桶：黄页快照）
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/quanttide/qtcloud-crowd-provider/internal/partner"
	"github.com/quanttide/qtcloud-crowd-provider/internal/publish"
	"github.com/quanttide/qtcloud-crowd-provider/internal/settlement"
	"github.com/quanttide/qtcloud-crowd-provider/internal/store"
	"github.com/quanttide/qtcloud-crowd-provider/internal/task"
)

func main() {
	addr := getenv("QTCLOUD_CROWD_ADDR", ":8080")
	dataKey := getenv("QTCLOUD_CROWD_DATA", "data/crowd.json")
	storeType := getenv("QTCLOUD_CROWD_STORE", "local")

	st := newStore(storeType)

	log.Printf("qtcloud-crowd-provider listening on %s (store=%s, data=%s)", addr, storeType, dataKey)
	if err := http.ListenAndServe(addr, newMux(st, dataKey)); err != nil {
		log.Fatal(err)
	}
}

// newStore 按 QTCLOUD_CROWD_STORE 选择存储后端。
func newStore(storeType string) store.Store {
	if storeType == "oss" {
		return store.NewOSS(store.OSSConfig{
			Endpoint:        getenv("QTCLOUD_OSS_ENDPOINT", ""),
			Bucket:          getenv("QTCLOUD_OSS_BUCKET", ""),
			PublicBucket:    getenv("QTCLOUD_OSS_PUBLIC_BUCKET", ""),
			AccessKeyID:     getenv("QTCLOUD_OSS_ACCESS_KEY_ID", ""),
			AccessKeySecret: getenv("QTCLOUD_OSS_ACCESS_KEY_SECRET", ""),
		})
	}
	return store.NewLocal()
}

// newMux 组装路由：人（执行方）/财（结算）/事（任务）三组 REST API。
// 任务注册在 /api/tasks 与 /api/tasks/（写回 API 走子路径 /claim、/deliver）。
func newMux(st store.Store, dataKey string) http.Handler {
	mux := http.NewServeMux()
	tasks := task.NewHandler(task.NewRepository(st, dataKey), publish.NewPublisher(st))
	mux.Handle("/api/tasks", tasks)
	mux.Handle("/api/tasks/", tasks)
	mux.Handle("/api/partners", partner.NewHandler(partner.NewRepository(st, dataKey)))
	mux.Handle("/api/settlements", settlement.NewHandler(settlement.NewRepository(st, dataKey)))
	return mux
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
