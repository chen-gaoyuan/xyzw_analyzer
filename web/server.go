package web

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"xyzw_study/internal/proxy"
	"xyzw_study/web/api"
)

//go:embed static
var staticFiles embed.FS

// StartWebServer 启动 WebSocket 服务器并开始捕获游戏数据包
func StartWebServer() {
	// 初始化存储
	if err := api.InitStorage(); err != nil {
		log.Fatal("初始化存储失败:", err)
	}
	// 设置 WebSocket 路由
	http.HandleFunc("/ws", api.HandleWebSocket)

	// 设置备注API路由
	http.HandleFunc("/api/notes/save", api.HandleSaveNotes)
	http.HandleFunc("/api/notes/load", api.HandleLoadNotes)

	// 设置调试消息API路由
	http.HandleFunc("/api/debug/send", api.HandleDebugMessage)

	// 添加脚本相关API路由
	http.HandleFunc("/api/scripts/save", api.HandleSaveScript)
	http.HandleFunc("/api/scripts/load", api.HandleLoadScripts)
	http.HandleFunc("/api/scripts/delete", api.HandleDeleteScript)

	// 使用嵌入的静态文件
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal("无法加载嵌入的静态文件:", err)
	}

	// 提供静态文件服务
	http.Handle("/", http.FileServer(http.FS(staticFS)))

	// 启动调试消息队列消费者
	go api.ConsumeDebugQueue()

	// 开始捕获游戏数据包
	go proxy.StartCapture(api.HandleGamePacket)

	// 启动 HTTP 服务器
	port := 12582
	log.Printf("咸鱼之王调试服务器已启动，访问 http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
