package crypto

import (
	"reflect"
	"time"
)

// BonEncoder 实现二进制对象表示法的编码器
type BonEncoder struct {
	dw     *DataWriter
	strMap map[string]int
}

// NewBonEncoder 创建一个新的BonEncoder实例
func NewBonEncoder() *BonEncoder {
	return &BonEncoder{
		dw:     NewDataWriter(),
		strMap: make(map[string]int),
	}
}

// Reset 重置编码器状态
func (e *BonEncoder) Reset() {
	e.dw.Reset()
	e.strMap = make(map[string]int)
}

// EncodeInt 编码整数
func (e *BonEncoder) EncodeInt(val int) error {
	e.dw.WriteInt8(1)
	e.dw.WriteInt32(val)
	return nil
}

// EncodeLong 编码长整数
func (e *BonEncoder) EncodeLong(val interface{}) error {
	e.dw.WriteInt8(2)

	switch v := val.(type) {
	case int64:
		e.dw.WriteInt64(v)
	case Int64:
		e.dw.WriteInt32(v.Low)
		e.dw.WriteInt32(v.High)
	default:
		// 尝试转换为int64
		e.dw.WriteInt64(reflect.ValueOf(val).Int())
	}

	return nil
}

// EncodeFloat 编码32位浮点数
func (e *BonEncoder) EncodeFloat(val float32) error {
	e.dw.WriteInt8(3)
	e.dw.WriteFloat32(val)
	return nil
}

// EncodeDouble 编码64位浮点数
func (e *BonEncoder) EncodeDouble(val float64) error {
	e.dw.WriteInt8(4)
	e.dw.WriteFloat64(val)
	return nil
}

// EncodeNumber 根据数值类型自动选择编码方式
func (e *BonEncoder) EncodeNumber(val float64) error {
	intVal := int(val)

	// 检查是否为整数
	if float64(intVal) == val {
		// 检查是否为32位整数范围
		if intVal == int(int32(intVal)) {
			return e.EncodeInt(intVal)
		}
		// 否则编码为长整数
		return e.EncodeLong(int64(val))
	}

	// 检查是否为浮点数
	if float64(float32(val)) == val {
		return e.EncodeFloat(float32(val))
	}

	// 否则编码为双精度浮点数
	return e.EncodeDouble(val)
}

// EncodeString 编码字符串
func (e *BonEncoder) EncodeString(val string) error {
	// 检查字符串是否已经编码过
	if index, exists := e.strMap[val]; exists {
		e.dw.WriteInt8(99)
		e.dw.Write7BitInt(index)
	} else {
		e.dw.WriteInt8(5)
		e.dw.WriteUTF(val)
		e.strMap[val] = len(e.strMap)
	}
	return nil
}

// EncodeBoolean 编码布尔值
func (e *BonEncoder) EncodeBoolean(val bool) error {
	e.dw.WriteInt8(6)
	if val {
		e.dw.WriteInt8(1)
	} else {
		e.dw.WriteInt8(0)
	}
	return nil
}

// EncodeNull 编码空值
func (e *BonEncoder) EncodeNull() error {
	e.dw.WriteInt8(0)
	return nil
}

// EncodeDateTime 编码日期时间
func (e *BonEncoder) EncodeDateTime(val time.Time) error {
	e.dw.WriteInt8(10)
	e.dw.WriteInt64(val.UnixNano() / int64(time.Millisecond))
	return nil
}

// EncodeBinary 编码二进制数据
func (e *BonEncoder) EncodeBinary(val []byte) error {
	e.dw.WriteInt8(7)
	e.dw.Write7BitInt(len(val))
	e.dw.WriteUint8Array(val, 0, len(val))
	return nil
}

// EncodeArray 编码数组
func (e *BonEncoder) EncodeArray(val []interface{}) error {
	e.dw.WriteInt8(9)
	e.dw.Write7BitInt(len(val))

	for _, item := range val {
		if err := e.Encode(item); err != nil {
			return err
		}
	}

	return nil
}

// EncodeMap 编码映射
func (e *BonEncoder) EncodeMap(val map[string]interface{}) error {
	e.dw.WriteInt8(8)
	e.dw.Write7BitInt(len(val))

	for k, v := range val {
		if err := e.Encode(k); err != nil {
			return err
		}
		if err := e.Encode(v); err != nil {
			return err
		}
	}

	return nil
}

// EncodeObject 编码对象
func (e *BonEncoder) EncodeObject(val interface{}) error {
	// 使用反射获取对象的字段
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		// 如果不是结构体，尝试转换为map
		return e.EncodeMap(structToMap(val))
	}

	// 计算有效字段数量
	validFields := 0
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		// 跳过以$开头的字段和函数类型字段
		if field.Name[0] == '$' || field.Type.Kind() == reflect.Func {
			continue
		}
		validFields++
	}

	e.dw.WriteInt8(8)
	e.dw.Write7BitInt(validFields)

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		// 跳过以$开头的字段和函数类型字段
		if field.Name[0] == '$' || field.Type.Kind() == reflect.Func {
			continue
		}

		// 编码字段名
		if err := e.Encode(field.Name); err != nil {
			return err
		}

		// 编码字段值
		if err := e.Encode(v.Field(i).Interface()); err != nil {
			return err
		}
	}

	return nil
}

// Encode 根据值的类型自动选择编码方式
func (e *BonEncoder) Encode(val interface{}) error {
	if val == nil {
		return e.EncodeNull()
	}

	switch v := val.(type) {
	case int:
		return e.EncodeInt(v)
	case int8:
		return e.EncodeInt(int(v))
	case int16:
		return e.EncodeInt(int(v))
	case int32:
		return e.EncodeInt(int(v))
	case int64:
		return e.EncodeLong(v)
	case uint:
		return e.EncodeLong(int64(v))
	case uint8:
		return e.EncodeInt(int(v))
	case uint16:
		return e.EncodeInt(int(v))
	case uint32:
		return e.EncodeLong(int64(v))
	case uint64:
		return e.EncodeLong(int64(v))
	case float32:
		return e.EncodeFloat(v)
	case float64:
		return e.EncodeDouble(v)
	case bool:
		return e.EncodeBoolean(v)
	case string:
		return e.EncodeString(v)
	case Int64:
		return e.EncodeLong(v)
	case []byte:
		return e.EncodeBinary(v)
	case []interface{}:
		return e.EncodeArray(v)
	case map[string]interface{}:
		return e.EncodeMap(v)
	case time.Time:
		return e.EncodeDateTime(v)
	default:
		// 处理其他类型
		rv := reflect.ValueOf(val)

		// 处理数组和切片
		if rv.Kind() == reflect.Array || rv.Kind() == reflect.Slice {
			if rv.Type().Elem().Kind() == reflect.Uint8 {
				// []byte类型
				bytes := make([]byte, rv.Len())
				for i := 0; i < rv.Len(); i++ {
					bytes[i] = byte(rv.Index(i).Uint())
				}
				return e.EncodeBinary(bytes)
			}

			// 其他数组类型
			array := make([]interface{}, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				array[i] = rv.Index(i).Interface()
			}
			return e.EncodeArray(array)
		}

		// 处理Map
		if rv.Kind() == reflect.Map {
			if rv.Type().Key().Kind() == reflect.String {
				mapVal := make(map[string]interface{}, rv.Len())
				iter := rv.MapRange()
				for iter.Next() {
					mapVal[iter.Key().String()] = iter.Value().Interface()
				}
				return e.EncodeMap(mapVal)
			}
		}

		// 处理结构体
		if rv.Kind() == reflect.Struct {
			return e.EncodeObject(val)
		}

		// 默认处理为null
		return e.EncodeNull()
	}
}

// GetBytes 获取编码后的字节数据
func (e *BonEncoder) GetBytes(copy bool) []byte {
	return e.dw.GetBytes(copy)
}

// structToMap 将结构体转换为map
func structToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return result
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		// 跳过以$开头的字段和函数类型字段
		if field.Name[0] == '$' || field.Type.Kind() == reflect.Func {
			continue
		}

		fieldValue := val.Field(i).Interface()
		result[field.Name] = fieldValue
	}

	return result
}
