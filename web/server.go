package web

import (
	"fmt"
	"log"
	"net/http"
	"xyzw_study/internal/proxy"
	"xyzw_study/web/api"
)

// StartWebServer 启动 WebSocket 服务器并开始捕获游戏数据包
func StartWebServer() {
	// 设置 WebSocket 路由
	http.HandleFunc("/ws", api.HandleWebSocket)

	// 设置备注API路由
	http.HandleFunc("/api/notes/save", api.HandleSaveNotes)
	http.HandleFunc("/api/notes/load", api.HandleLoadNotes)

	// 设置调试消息API路由
	http.HandleFunc("/api/debug/send", api.HandleDebugMessage)
	// 添加以下路由到现有的路由注册部分
	http.HandleFunc("/api/scripts/save", api.HandleSaveScript)
	http.HandleFunc("/api/scripts/load", api.HandleLoadScripts)
	http.HandleFunc("/api/scripts/delete", api.HandleDeleteScript)
	// 直接提供静态文件
	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/", http.StripPrefix("/", fs))

	// 启动调试消息队列消费者
	go api.ConsumeDebugQueue()

	// 开始捕获游戏数据包
	go proxy.StartCapture(api.HandleGamePacket)

	// 启动 HTTP 服务器
	port := 12582
	fmt.Printf("Web 服务器已启动，访问 http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
