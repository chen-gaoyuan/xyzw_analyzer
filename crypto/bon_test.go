package crypto

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestDecode(t *testing.T) {
	b, _ := hex.DecodeString("706c09185464cb3c1414941c161113627166677d7b7a111d253a222c3a2539636c111c64787560727b667911127c7b66607b6614141414")
	r, e := DecryptXAndDecode(b)
	if e != nil {
		panic(e)
	}
	fmt.Println(r)
	r2 := r.(map[string]interface{})
	body, ok := r2["body"].([]byte)
	if ok {
		body_r := DecodeFromBytes(body)
		r2["body"] = body_r
	}

}
func TestEncode(t *testing.T) {
	r := map[string]interface{}{
		"ClientVersion": "1.37.2-wx",
		"InviteUid":     0,
		"Platform":      "hortor",
		"PlatformExt":   "mix",
		"Scene":         "",
	}
	fmt.Println(r)
	data := EncodeToBytes(r)

	fmt.Println(hex.EncodeToString(data))
	r2 := map[string]interface{}{
		"ack":  1,
		"body": data,
		"cmd":  "role_getroleinfo",
		"seq":  1,
		"time": int64(1741969398878),
	}
	data2, e := EncodeAndEncryptX(r2)
	if e != nil {
		panic(e)
	}
	fmt.Println(hex.EncodeToString(data2))
}
