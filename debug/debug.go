package debug

import (
	"encoding/hex"
	gamemitm "github.com/husanpao/game-mitm"
	"strings"
	"xyzw_study/crypto"
)

// Direction 定义消息方向类型
type Direction int

const (
	Send Direction = iota
	Receive
)

type GamePacket struct {
	Raw       []byte
	RawData   any
	Direction Direction // 使用枚举类型标识消息方向
	*gamemitm.Session
}

func StartCapture(f func(message GamePacket)) {
	proxy := gamemitm.NewProxy()
	proxy.SetVerbose(false)
	proxy.OnRequest("xxz-xyzw.hortorgames.com").Do(func(body []byte, ctx *gamemitm.ProxyCtx) []byte {
		if f == nil {
			return body
		}
		result := make([]byte, len(body))
		copy(result, body)
		hexStr := hex.EncodeToString(body)
		if strings.HasPrefix(hexStr, "7078") {

			f(GamePacket{result, crypto.DecodeX(body), Send, ctx.WSSession})
		}
		return result
	})
	proxy.OnResponse("xxz-xyzw.hortorgames.com").Do(func(body []byte, ctx *gamemitm.ProxyCtx) []byte {
		if f == nil {
			return body
		}
		result := make([]byte, len(body))
		copy(result, body)
		hexStr := hex.EncodeToString(body)
		if strings.HasPrefix(hexStr, "7078") {

			f(GamePacket{result, crypto.DecodeX(body), Receive, ctx.WSSession})
		}
		return result
	})
	proxy.Start()
}
