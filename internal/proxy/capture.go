package proxy

import (
	"encoding/hex"
	gamemitm "github.com/husanpao/game-mitm"
	"strings"
	"xyzw_study/internal/crypto/bon"
)

// Direction 定义数据包方向
type Direction int

const (
	// Send 表示从客户端发送到服务器的数据包
	Send Direction = iota
	// Receive 表示从服务器接收到客户端的数据包
	Receive
)

// GamePacket 定义游戏数据包结构
type GamePacket struct {
	Raw       []byte
	RawData   any
	Direction Direction // 使用枚举类型标识消息方向
	Session   *gamemitm.Session
}

// PacketHandler 定义处理数据包的函数类型
type PacketHandler func(packet GamePacket)

// StartCapture 开始捕获游戏数据包
func StartCapture(handler PacketHandler) {
	proxy := gamemitm.NewProxy()
	proxy.SetVerbose(false)
	proxy.OnRequest("xxz-xyzw.hortorgames.com").Do(func(body []byte, ctx *gamemitm.ProxyCtx) []byte {
		if handler == nil {
			return body
		}
		result := make([]byte, len(body))
		copy(result, body)
		hexStr := hex.EncodeToString(body)
		if strings.HasPrefix(hexStr, "7078") {

			handler(GamePacket{result, bon.DecodeX(body), Send, ctx.WSSession})
		}
		return result
	})
	proxy.OnResponse("xxz-xyzw.hortorgames.com").Do(func(body []byte, ctx *gamemitm.ProxyCtx) []byte {
		if handler == nil {
			return body
		}
		result := make([]byte, len(body))
		copy(result, body)
		hexStr := hex.EncodeToString(body)
		if strings.HasPrefix(hexStr, "7078") {

			handler(GamePacket{result, bon.DecodeX(body), Receive, ctx.WSSession})
		}
		return result
	})
	proxy.Start()
}
