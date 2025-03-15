package web

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"xyzw_study/debug"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有跨域请求
		},
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

// WSMessage 定义发送到前端的消息结构
type WSMessage struct {
	Call string `json:"call"` // "client" 或 "server"
	Msg  any    `json:"msg"`  // JSON 字符串
}

// 处理 WebSocket 连接
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket 升级错误:", err)
		return
	}
	defer conn.Close()

	// 添加新客户端
	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	// 客户端断开连接时移除
	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
	}()

	// 保持连接，直到客户端断开
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// 向所有连接的客户端广播消息
func broadcastMessage(message WSMessage) {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Println("JSON 编码错误:", err)
		return
	}

	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, jsonMessage)
		if err != nil {
			log.Println("发送消息错误:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// 处理游戏数据包
func handleGamePacket(packet debug.GamePacket) {

	// 确定消息方向
	var call string
	if packet.Direction == debug.Send {
		call = "client"
	} else {
		call = "server"
	}

	// 广播消息到所有连接的 WebSocket 客户端
	broadcastMessage(WSMessage{
		Call: call,
		Msg:  packet.RawData,
	})
}

// StartWebServer 启动 WebSocket 服务器并开始捕获游戏数据包
func StartWebServer() {
	// 设置 WebSocket 路由
	http.HandleFunc("/ws", handleWebSocket)

	// 直接提供静态文件
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", http.StripPrefix("/", fs))

	// 开始捕获游戏数据包
	go debug.StartCapture(handleGamePacket)

	// 启动 HTTP 服务器
	port := 12582
	fmt.Printf("Web 服务器已启动，访问 http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// 在 main 函数中调用 StartWebServer
func main() {
	StartWebServer()
}
