package crypto

import (
	"bytes"
	"fmt"
	"github.com/pierrec/lz4"
	"io"
	"log"
	"math/rand"
	"time"
)

// 初始化随机数生成器
func init() {
	rand.Seed(time.Now().UnixNano())
}

// CompressLZ4 使用LZ4算法压缩数据
func CompressLZ4(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}

	var compressedData bytes.Buffer
	encoder := lz4.NewWriter(&compressedData)
	_, err := encoder.Write(data)
	if err != nil {
		log.Fatalf("压缩数据失败: %v", err)
	}
	encoder.Close()
	return compressedData.Bytes()
}

// DecompressLZ4 使用LZ4算法解压缩数据
func DecompressLZ4(data []byte, originalSize int) []byte {
	if len(data) == 0 {
		return []byte{}
	}
	r := lz4.NewReader(bytes.NewReader(data))
	// 将解压后的数据复制到 decompressed
	r2, err := io.ReadAll(r)
	if err != nil {
		fmt.Println("Error decompressing data:", err)
		return nil
	}
	return r2
}

// EncryptLX 实现"lx"加密算法
// 对应JavaScript中的E函数
func EncryptLX(data []byte) []byte {
	// 首先压缩数据
	compressed := CompressLZ4(data)

	// 生成随机密钥 (2-250范围内)
	key := byte(2 + rand.Intn(248))

	// 对前100个字节进行异或加密
	n := min(len(compressed), 100)
	for i := 0; i < n; i++ {
		compressed[i] ^= key
	}

	// 设置特定的头部字节
	if len(compressed) >= 4 {
		compressed[0] = 112 // 'p'
		compressed[1] = 108 // 'l'

		// 将密钥信息嵌入到头部字节中
		compressed[2] = 170&compressed[2] |
			((key >> 7 & 1) << 6) |
			((key >> 6 & 1) << 4) |
			((key >> 5 & 1) << 2) |
			((key >> 4 & 1) << 0)

		compressed[3] = 170&compressed[3] |
			((key >> 3 & 1) << 6) |
			((key >> 2 & 1) << 4) |
			((key >> 1 & 1) << 2) |
			((key >> 0 & 1) << 0)
	}

	return compressed
}

// DecryptLX 实现"lx"解密算法
// 对应JavaScript中的y函数
func DecryptLX(data []byte) []byte {
	if len(data) < 4 {
		return data
	}

	// 从头部字节中提取密钥
	key := byte(
		((data[2] >> 6 & 1) << 7) |
			((data[2] >> 4 & 1) << 6) |
			((data[2] >> 2 & 1) << 5) |
			((data[2] >> 0 & 1) << 4) |
			((data[3] >> 6 & 1) << 3) |
			((data[3] >> 4 & 1) << 2) |
			((data[3] >> 2 & 1) << 1) |
			((data[3] >> 0 & 1) << 0))

	// 对前100个字节进行异或解密
	n := min(len(data), 100)
	for i := 2; i < n; i++ {
		data[i] ^= key
	}

	// 设置LZ4头部字节
	data[0] = 4
	data[1] = 34
	data[2] = 77
	data[3] = 24

	// 解压缩数据
	return DecompressLZ4(data, 0)
}

// EncryptX 实现"x"加密算法
// 对应JavaScript中的R函数
func EncryptX(data []byte) []byte {
	// 生成随机ID (32位整数)
	randomID := rand.Uint32()

	// 创建新的缓冲区，前4字节存放随机ID
	result := make([]byte, len(data)+4)
	result[0] = byte(randomID & 0xFF)
	result[1] = byte((randomID >> 8) & 0xFF)
	result[2] = byte((randomID >> 16) & 0xFF)
	result[3] = byte((randomID >> 24) & 0xFF)

	// 复制原始数据
	copy(result[4:], data)

	// 生成随机密钥 (2-250范围内)
	key := byte(2 + rand.Intn(248))

	// 对所有字节进行异或加密
	for i := len(result) - 1; i >= 0; i-- {
		result[i] ^= key
	}

	// 设置特定的头部字节
	result[0] = 112 // 'p'
	result[1] = 120 // 'x'

	// 将密钥信息嵌入到头部字节中
	result[2] = 170&result[2] |
		((key >> 7 & 1) << 6) |
		((key >> 6 & 1) << 4) |
		((key >> 5 & 1) << 2) |
		((key >> 4 & 1) << 0)

	result[3] = 170&result[3] |
		((key >> 3 & 1) << 6) |
		((key >> 2 & 1) << 4) |
		((key >> 1 & 1) << 2) |
		((key >> 0 & 1) << 0)

	return result
}

// DecryptX 实现"x"解密算法
// 对应JavaScript中的M函数
func DecryptX(data []byte) []byte {
	if len(data) < 4 {
		return data
	}

	// 从头部字节中提取密钥
	key := byte(
		((data[2] >> 6 & 1) << 7) |
			((data[2] >> 4 & 1) << 6) |
			((data[2] >> 2 & 1) << 5) |
			((data[2] >> 0 & 1) << 4) |
			((data[3] >> 6 & 1) << 3) |
			((data[3] >> 4 & 1) << 2) |
			((data[3] >> 2 & 1) << 1) |
			((data[3] >> 0 & 1) << 0))

	// 对所有字节进行异或解密
	for i := len(data) - 1; i >= 4; i-- {
		data[i] ^= key
	}

	// 返回去除前4个字节的数据
	return data[4:]
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
