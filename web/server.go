package web

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	gamemitm "github.com/husanpao/game-mitm"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"xyzw_study/crypto"
	"xyzw_study/data"
	"xyzw_study/debug"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有跨域请求
		},
	}
	clients       = make(map[*websocket.Conn]bool)
	clientsMu     sync.Mutex
	notesFilePath = "./web/notes.json"           // 备注数据文件路径
	debugQueue    = make(chan DebugMessage, 100) // 调试消息队列
	game          *gamemitm.Session
	sendSeq       float64
)

// Notes 定义备注数据结构
type Notes struct {
	CommandNotes map[string]string            `json:"commandNotes"`
	KeyNotes     map[string]map[string]string `json:"keyNotes"`
}

// 保存备注数据到文件
func saveNotes(notes Notes) error {
	data, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(notesFilePath, data, 0644)
}

// 从文件加载备注数据
func loadNotes() (Notes, error) {
	var notes Notes
	notes.CommandNotes = make(map[string]string)
	notes.KeyNotes = make(map[string]map[string]string)

	data, err := os.ReadFile(notesFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，返回空备注
			return notes, nil
		}
		return notes, err
	}

	err = json.Unmarshal(data, &notes)
	return notes, err
}

// 处理备注保存请求
func handleSaveNotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	var notes Notes
	err := json.NewDecoder(r.Body).Decode(&notes)
	if err != nil {
		http.Error(w, "解析请求数据失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = saveNotes(notes)
	if err != nil {
		http.Error(w, "保存备注失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// 处理备注加载请求
func handleLoadNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := loadNotes()
	if err != nil {
		http.Error(w, "加载备注失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

// 处理调试消息发送
func handleDebugMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	var message DebugMessage
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, "解析请求数据失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 将消息添加到队列
	select {
	case debugQueue <- message:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	default:
		http.Error(w, "调试队列已满", http.StatusServiceUnavailable)
	}
}

// 消费调试消息队列
func consumeDebugQueue() {
	for {
		select {
		case msg := <-debugQueue:
			// 构造消息
			log.Println("收到调试消息:", msg)
			if game != nil {
				debugMsg := data.XYMsg{
					Ack:  0,
					Body: crypto.EncodeToBytes(msg.Data),
					Cmd:  msg.Cmd,
					Seq:  int32(sendSeq + 10000),
					Time: time.Now().UnixMilli(),
				}
				bs, err := crypto.EncodeAndEncryptX(debugMsg)
				result := make([]byte, len(bs))
				copy(result, bs)
				if err != nil {
					log.Println(err)
				} else {
					handleGamePacket(debug.GamePacket{bs, crypto.DecodeX(bs), debug.Send, nil})
					game.SendBinaryToServer(result)
				}

			}
			// 等待2秒
			time.Sleep(2 * time.Second)
		}
	}
}

// WSMessage 定义发送到前端的消息结构
type WSMessage struct {
	Call string `json:"call"` // "client" 或 "server"
	Msg  any    `json:"msg"`  // JSON 字符串
}
type DebugMessage struct {
	Cmd  string `json:"cmd"`
	Data any    `json:"data"`
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

	if game == nil {
		game = packet.Session
	}
	// 确定消息方向
	var call string
	if packet.Direction == debug.Send {
		call = "client"
		var m map[string]interface{}
		json.Unmarshal([]byte(packet.RawData.(string)), &m)
		if m["seq"] != nil {
			tempSeq := m["seq"].(float64) + 1
			if sendSeq < tempSeq {

				sendSeq = tempSeq
			}
		}
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

	// 设置备注API路由
	http.HandleFunc("/api/notes/save", handleSaveNotes)
	http.HandleFunc("/api/notes/load", handleLoadNotes)

	// 设置调试消息API路由
	http.HandleFunc("/api/debug/send", handleDebugMessage)

	// 直接提供静态文件
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", http.StripPrefix("/", fs))

	// 启动调试消息队列消费者
	go consumeDebugQueue()

	// 开始捕获游戏数据包
	go debug.StartCapture(handleGamePacket)

	// 启动 HTTP 服务器
	port := 12582
	fmt.Printf("Web 服务器已启动，访问 http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
