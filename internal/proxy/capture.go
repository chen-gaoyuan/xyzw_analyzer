package proxy

import (
	"encoding/hex"
	gamemitm "github.com/husanpao/game-mitm"
	"strings"
	"sync/atomic"
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

var seq int32

func NextSeq() int32 {
	return atomic.AddInt32(&seq, 1)
}

func CurrentSeq() int32 {
	return atomic.LoadInt32(&seq)
}

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
	seq = 0
	proxy.OnRequest("xxz-xyzw.hortorgames.com").Do(func(body []byte, ctx *gamemitm.ProxyCtx) []byte {
		if handler == nil {
			return body
		}
		// 拷贝原始请求体
		original := make([]byte, len(body))
		copy(original, body)

		// 判断前两个字节是否是 0x70 0x78（即 "px"）
		if len(body) >= 2 && body[0] == 0x70 && body[1] == 0x78 {
			originalStr := bon.DecodeX(body)
			if strings.Contains(originalStr, "_sys/ack") {
				return original
			}

			processed := bon.EncodeReplaceSeq(original, NextSeq())
			// 给 DecodeX 使用一份拷贝，避免修改原 processed
			decodedInput := make([]byte, len(processed))
			copy(decodedInput, processed)
			handler(GamePacket{processed, bon.DecodeX(decodedInput), Send, ctx.WSSession})

			return processed
		}
		return original
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
