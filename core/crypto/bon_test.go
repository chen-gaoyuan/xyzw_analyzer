package crypto

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestDecode(t *testing.T) {
	b, _ := hex.DecodeString("7078d642e0ededeb898b83e9e9e8e8e8edec8a878c91efb1e0edede5ab84818d869cbe8d9a9b818786ede1d9c6dbdfc6dac59f90ede1a1869e819c8dbd818ce9e8e8e8e8ede0b884899c8e879a85edee80879a9c879aede3b884899c8e879a85ad909cedeb858190ededbb8b8d868dede8edeb8b858cedf89a87848db78f8d9c9a87848d81868e87edeb9b8d99e9e9e8e8e8edec9c81858deab6749f7d7de9e8e8")
	r, e := DecryptXAndDecode(b)
	if e != nil {
		panic(e)
	}
	fmt.Println(r)
	r2 := r.(map[string]interface{})
	body, ok := r2["body"].([]byte)
	if ok {
		body_r := DecodeFromBytes(body)
		fmt.Println(body_r)
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
