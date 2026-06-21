package data

import (
	"fmt"
	"strings"
)

type GFSObject struct {
	DataHolder map[string]*GFSDataWrapper
}

func MakeGFSObject() *GFSObject {
	return &GFSObject{
		DataHolder: make(map[string]*GFSDataWrapper),
	}
}

func (o *GFSObject) String() string {
	return o.Dump(0)
}

func (o *GFSObject) Dump(indents int) string {
	dump := fmt.Sprintf("(%s):", strings.ToLower(GFS_OBJECT.Name))
	for key, value := range o.DataHolder {
		dump += "\n" + value.Dump(key, indents+1)
	}
	return dump
}

func (o *GFSObject) Size() int {
	return len(o.DataHolder)
}

func (o *GFSObject) ContainsKey(key string) bool {
	_, exists := o.DataHolder[key]
	return exists
}

func (o *GFSObject) RemoveKey(key string) {
	delete(o.DataHolder, key)
}

func (o *GFSObject) Get(key string) *GFSDataWrapper {
	return o.DataHolder[key]
}

// infers the wire type from the Go type of value.
func (o *GFSObject) Put(key string, value interface{}) *GFSObject {
	o.DataHolder[key] = NewGFSDataWrapper(value)
	return o
}

// used when inference would be wrong (like forcing SHORT over INT).
func (o *GFSObject) PutObject(key string, wrapper *GFSDataWrapper) *GFSObject {
	o.DataHolder[key] = wrapper
	return o
}

func (o *GFSObject) put(key string, data interface{}, typeID GFSDataType) *GFSObject {
	o.DataHolder[key] = &GFSDataWrapper{Data: data, TypeID: typeID}
	return o
}

/*
 * Type insertion:
 *
 */
func (o *GFSObject) PutNull(key string) *GFSObject {
	return o.put(key, nil, NULL)
}
func (o *GFSObject) PutBool(key string, value bool) *GFSObject {
	return o.put(key, value, BOOL)
}
func (o *GFSObject) PutByte(key string, value byte) *GFSObject {
	return o.put(key, value, BYTE)
}
func (o *GFSObject) PutShort(key string, value int16) *GFSObject {
	return o.put(key, value, SHORT)
}
func (o *GFSObject) PutInt(key string, value int) *GFSObject {
	return o.put(key, value, INT)
}
func (o *GFSObject) PutLong(key string, value int64) *GFSObject {
	return o.put(key, value, LONG)
}
func (o *GFSObject) PutFloat(key string, value float32) *GFSObject {
	return o.put(key, value, FLOAT)
}
func (o *GFSObject) PutDouble(key string, value float64) *GFSObject {
	return o.put(key, value, DOUBLE)
}
func (o *GFSObject) PutUtfString(key string, value string) *GFSObject {
	return o.put(key, value, UTF_STRING)
}
func (o *GFSObject) PutBoolArray(key string, value []bool) *GFSObject {
	return o.put(key, value, BOOL_ARRAY)
}
func (o *GFSObject) PutByteArray(key string, value []byte) *GFSObject {
	return o.put(key, value, BYTE_ARRAY)
}
func (o *GFSObject) PutShortArray(key string, value []int16) *GFSObject {
	return o.put(key, value, SHORT_ARRAY)
}
func (o *GFSObject) PutIntArray(key string, value []int) *GFSObject {
	return o.put(key, value, INT_ARRAY)
}
func (o *GFSObject) PutLongArray(key string, value []int64) *GFSObject {
	return o.put(key, value, LONG_ARRAY)
}
func (o *GFSObject) PutFloatArray(key string, value []float32) *GFSObject {
	return o.put(key, value, FLOAT_ARRAY)
}
func (o *GFSObject) PutDoubleArray(key string, value []float64) *GFSObject {
	return o.put(key, value, DOUBLE_ARRAY)
}
func (o *GFSObject) PutUtfStringArray(key string, value []string) *GFSObject {
	return o.put(key, value, UTF_STRING_ARRAY)
}
func (o *GFSObject) PutGFSArray(key string, value *GFSArray) *GFSObject {
	return o.put(key, value, GFS_ARRAY)
}
func (o *GFSObject) PutGFSObject(key string, value *GFSObject) *GFSObject {
	return o.put(key, value, GFS_OBJECT)
}

/*
 * Type Retrieval:
 *   each returns (value, ok)
 */
func getTyped[T any](o *GFSObject, key string) (T, bool) {
	var zero T
	wrapper := o.DataHolder[key]
	if wrapper == nil {
		return zero, false
	}
	value, ok := wrapper.Data.(T)
	if !ok {
		return zero, false
	}
	return value, true
}

func (o *GFSObject) GetBool(key string) (bool, bool)             { return getTyped[bool](o, key) }
func (o *GFSObject) GetByte(key string) (byte, bool)             { return getTyped[byte](o, key) }
func (o *GFSObject) GetShort(key string) (int16, bool)           { return getTyped[int16](o, key) }
func (o *GFSObject) GetInt(key string) (int, bool)               { return getTyped[int](o, key) }
func (o *GFSObject) GetLong(key string) (int64, bool)            { return getTyped[int64](o, key) }
func (o *GFSObject) GetFloat(key string) (float32, bool)         { return getTyped[float32](o, key) }
func (o *GFSObject) GetDouble(key string) (float64, bool)        { return getTyped[float64](o, key) }
func (o *GFSObject) GetUtfString(key string) (string, bool)      { return getTyped[string](o, key) }
func (o *GFSObject) GetBoolArray(key string) ([]bool, bool)      { return getTyped[[]bool](o, key) }
func (o *GFSObject) GetByteArray(key string) ([]byte, bool)      { return getTyped[[]byte](o, key) }
func (o *GFSObject) GetShortArray(key string) ([]int16, bool)    { return getTyped[[]int16](o, key) }
func (o *GFSObject) GetIntArray(key string) ([]int, bool)        { return getTyped[[]int](o, key) }
func (o *GFSObject) GetLongArray(key string) ([]int64, bool)     { return getTyped[[]int64](o, key) }
func (o *GFSObject) GetFloatArray(key string) ([]float32, bool)  { return getTyped[[]float32](o, key) }
func (o *GFSObject) GetDoubleArray(key string) ([]float64, bool) { return getTyped[[]float64](o, key) }
func (o *GFSObject) GetUtfStringArray(key string) ([]string, bool) {
	return getTyped[[]string](o, key)
}
func (o *GFSObject) GetGFSArray(key string) (*GFSArray, bool)   { return getTyped[*GFSArray](o, key) }
func (o *GFSObject) GetGFSObject(key string) (*GFSObject, bool) { return getTyped[*GFSObject](o, key) }
