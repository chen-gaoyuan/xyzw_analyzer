package crypto

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestDecode(t *testing.T) {
	b, _ := hex.DecodeString("70784b3d9f929294f6f4fc96969797979293f5f8f3ee90ce9f92929cc7fbf6e3f1f8e5fad2efe39294fafeef9292c4f4f2f9f29297929ad4fbfef2f9e3c1f2e5e4fef8f9929ea6b9a4a0b9a5bae0ef929edef9e1fee3f2c2fef39697979797929fc7fbf6e3f1f8e5fa9291fff8e5e3f8e59294f4faf39287e5f8fbf2c8f0f2e3e5f8fbf2fef9f1f89294e4f2e696969797979293e3fefaf295c90be00202969797")
	r, e := DecryptXAndDecode(b)
	if e != nil {
		panic(e)
	}
	fmt.Println(r)
	r2 := r.(map[string]interface{})
	body, ok := r2["body"].([]byte)
	if ok {
		body_r, _ := Decode(body)
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
	data, e := Encode(r, true)
	if e != nil {
		panic(e)
	}
	fmt.Println(hex.EncodeToString(data))
	r2 := map[string]interface{}{
		"ack":  1,
		"body": data,
		"cmd":  "role_getroleinfo",
		"seq":  1,
		"time": int64(1741969398878),
	}
	data2, e := EncodeAndEncryptX(r2)
	fmt.Println(hex.EncodeToString(data2))
}
