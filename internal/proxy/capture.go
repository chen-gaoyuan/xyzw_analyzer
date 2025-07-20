package proxy

import (
	"log"
	"os"
	"sync/atomic"
	"xyzw_study/internal/crypto/bon"

	gamemitm "github.com/husanpao/game-mitm"
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
var clientMSg map[int32]int32

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
	// 创建或打开日志文件
	file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 设置日志输出到文件
	log.SetOutput(file)

	proxy := gamemitm.NewProxy()
	proxy.SetVerbose(false)
	seq = 0
	clientMSg = make(map[int32]int32)
	proxy.OnRequest("xxz-xyzw.hortorgames.com").Do(func(body []byte, ctx *gamemitm.ProxyCtx) []byte {
		if handler == nil {
			return body
		}
		// 拷贝原始请求体
		original := make([]byte, len(body))
		copy(original, body)
		if len(body) >= 2 && body[0] == 0x70 && body[1] == 0x78 {
			msg := bon.DecodeXAsMap(body)
			if msg["seq"] != nil {
				gameSeq := msg["seq"].(int32)
				if gameSeq == 1 {
					seq = 0
				}
			}

			var processed []byte
			if msg["cmd"] == nil {
				return original
			}
			if msg["cmd"].(string) == "_sys/ack" {
				processed = bon.EncodeReplaceAck(original, seq+1)
			} else {
				processed = bon.EncodeReplaceSeq(original, NextSeq())
				clientMSg[CurrentSeq()] = msg["seq"].(int32)
			}
			// 给 DecodeX 使用一份拷贝，避免修改原 processed
			decodedInput := make([]byte, len(processed))
			copy(decodedInput, processed)
			updateStr := bon.DecodeX(decodedInput)
			log.Printf("Send => %s", updateStr)

			handler(GamePacket{processed, updateStr, Send, ctx.WSSession})
			return processed
		}
		return original
	})
	proxy.OnResponse("xxz-xyzw.hortorgames.com").Do(func(body []byte, ctx *gamemitm.ProxyCtx) []byte {
		if handler == nil {
			return body
		}
		// 拷贝原始请求体
		original := make([]byte, len(body))
		copy(original, body)
		if len(body) >= 2 && body[0] == 0x70 && body[1] == 0x78 {

			msg := bon.DecodeXAsMap(body)
			var processed []byte
			if msg["cmd"] == nil {
				return original
			}
			if msg["cmd"].(string) == "_sys/ack" {
				processed = bon.EncodeReplaceAck(original, CurrentSeq())
			} else {
				if msg["resp"] != nil {
					if rseq, ok := clientMSg[msg["resp"].(int32)]; ok {
						processed = bon.EncodeReplaceResp(original, rseq)
					} else {
						processed = original
					}
				} else {
					processed = original
				}

			}
			decodedInput := make([]byte, len(processed))
			copy(decodedInput, processed)
			// 给 DecodeX 使用一份拷贝，避免修改原 processed
			updateStr := bon.DecodeX(decodedInput)
			log.Printf("Recv <= %s", updateStr)
			handler(GamePacket{processed, updateStr, Receive, ctx.WSSession})
			return processed
		}
		return original
	})
	proxy.OnRequest("xxz-xyzw-new.hortorgames.com").Do(func(body []byte, ctx *gamemitm.ProxyCtx) []byte {
		if handler == nil {
			return body
		}
		// 拷贝原始请求体
		original := make([]byte, len(body))
		copy(original, body)
		if len(body) >= 2 && body[0] == 0x70 && body[1] == 0x78 {
			msg := bon.DecodeXAsMap(body)
			if msg["seq"] != nil {
				gameSeq := msg["seq"].(int32)
				if gameSeq == 1 {
					seq = 0
				}
			}

			var processed []byte
			if msg["cmd"] == nil {
				return original
			}
			if msg["cmd"].(string) == "_sys/ack" {
				processed = bon.EncodeReplaceAck(original, seq+1)
			} else {
				processed = bon.EncodeReplaceSeq(original, NextSeq())
				clientMSg[CurrentSeq()] = msg["seq"].(int32)
			}
			// 给 DecodeX 使用一份拷贝，避免修改原 processed
			decodedInput := make([]byte, len(processed))
			copy(decodedInput, processed)
			updateStr := bon.DecodeX(decodedInput)
			log.Printf("NewSend => %s", updateStr)

			handler(GamePacket{processed, updateStr, Send, ctx.WSSession})
			return processed
		}
		return original
	})
	proxy.OnResponse("xxz-xyzw-new.hortorgames.com").Do(func(body []byte, ctx *gamemitm.ProxyCtx) []byte {
		if handler == nil {
			return body
		}
		// 拷贝原始请求体
		original := make([]byte, len(body))
		copy(original, body)
		if len(body) >= 2 && body[0] == 0x70 && body[1] == 0x78 {

			msg := bon.DecodeXAsMap(body)
			var processed []byte
			if msg["cmd"] == nil {
				return original
			}
			if msg["cmd"].(string) == "_sys/ack" {
				processed = bon.EncodeReplaceAck(original, CurrentSeq())
			} else {
				if msg["resp"] != nil {
					if rseq, ok := clientMSg[msg["resp"].(int32)]; ok {
						processed = bon.EncodeReplaceResp(original, rseq)
					} else {
						processed = original
					}
				} else {
					processed = original
				}

			}
			decodedInput := make([]byte, len(processed))
			copy(decodedInput, processed)
			// 给 DecodeX 使用一份拷贝，避免修改原 processed
			updateStr := bon.DecodeX(decodedInput)
			log.Printf("NewRecv <= %s", updateStr)
			handler(GamePacket{processed, updateStr, Receive, ctx.WSSession})
			return processed
		}
		return original
	})
	proxy.Start()
}
