package crypto

import (
	"errors"
)

var (
	// 全局单例解码器和编码器
	globalDecoder = NewBonDecoder()
	globalEncoder = NewBonEncoder()
)

type XYMsg struct {
	Ack  int
	Body any
	Cmd  string
	Seq  int
	Time int64
}

// Decode 解码二进制数据为Go对象
func Decode(data []byte) (interface{}, error) {
	if data == nil || len(data) == 0 {
		return nil, errors.New("empty data")
	}

	globalDecoder.Reset(data)
	return globalDecoder.Decode()
}

// Encode 将Go对象编码为二进制数据
func Encode(value interface{}, copy bool) ([]byte, error) {
	globalEncoder.Reset()
	if err := globalEncoder.Encode(value); err != nil {
		return nil, err
	}
	return globalEncoder.GetBytes(copy), nil
}

// EncodeAndEncryptLX 编码并使用LX算法加密
func EncodeAndEncryptLX(value interface{}) ([]byte, error) {
	data, err := Encode(value, true)
	if err != nil {
		return nil, err
	}
	return EncryptLX(data), nil
}

// DecryptLXAndDecode 使用LX算法解密并解码
func DecryptLXAndDecode(data []byte) (interface{}, error) {
	if data == nil || len(data) == 0 {
		return nil, errors.New("empty data")
	}

	decrypted := DecryptLX(data)
	return Decode(decrypted)
}

// EncodeAndEncryptX 编码并使用X算法加密
func EncodeAndEncryptX(value interface{}) ([]byte, error) {
	data, err := Encode(value, true)
	if err != nil {
		return nil, err
	}
	return EncryptX(data), nil
}

// DecryptXAndDecode 使用X算法解密并解码
func DecryptXAndDecode(data []byte) (interface{}, error) {
	if data == nil || len(data) == 0 {
		return nil, errors.New("empty data")
	}

	decrypted := DecryptX(data)
	return Decode(decrypted)
}

// 提供一些辅助函数，方便使用
func EncodeToBytes(value interface{}) []byte {
	data, err := Encode(value, true)
	if err != nil {
		return nil
	}
	return data
}

func DecodeFromBytes(data []byte) interface{} {
	result, err := Decode(data)
	if err != nil {
		return nil
	}
	return result
}
