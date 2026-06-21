package data

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sort"
)

/*
 * Wire format (all integers are signed and big-endian):
 *
 *   object  := 0x12 <uint16 count> ( key value )*
 *   array   := 0x11 <uint16 count> ( value )*
 *   key     := <uint16 byteLen> <utf8 bytes>
 *   value   := <typeID byte> <payload>
 *
 * Array element counts are uint16 except BYTE_ARRAY (uint32)
 * Floats & doubles are IEEE-754 big-endian.
 */

var (
	_ encoding.BinaryMarshaler   = (*GFSObject)(nil)
	_ encoding.BinaryUnmarshaler = (*GFSObject)(nil)
	_ encoding.BinaryMarshaler   = (*GFSArray)(nil)
	_ encoding.BinaryUnmarshaler = (*GFSArray)(nil)
)

func (o *GFSObject) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte(GFS_OBJECT.TypeID)
	if err := encodeObjectBody(buf, o); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (o *GFSObject) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	id, err := r.ReadByte()
	if err != nil {
		return err
	}
	if id != GFS_OBJECT.TypeID {
		return fmt.Errorf("gofox2x: expected GFS_OBJECT marker %d, got %d", GFS_OBJECT.TypeID, id)
	}
	decoded, err := decodeObjectBody(r)
	if err != nil {
		return err
	}
	o.DataHolder = decoded.DataHolder
	return nil
}

func (a *GFSArray) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte(GFS_ARRAY.TypeID)
	if err := encodeArrayBody(buf, a); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (a *GFSArray) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	id, err := r.ReadByte()
	if err != nil {
		return err
	}
	if id != GFS_ARRAY.TypeID {
		return fmt.Errorf("gofox2x: expected GFS_ARRAY marker %d, got %d", GFS_ARRAY.TypeID, id)
	}
	decoded, err := decodeArrayBody(r)
	if err != nil {
		return err
	}
	a.DataHolder = decoded.DataHolder
	return nil
}

/*
* Encoding:
*
 */

// writes the count and key/value entries, without type marker
func encodeObjectBody(buf *bytes.Buffer, o *GFSObject) error {
	keys := make([]string, 0, len(o.DataHolder))
	for k := range o.DataHolder {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	putU16(buf, uint16(len(keys)))
	for _, k := range keys {
		putUTF(buf, k)
		if err := encodeValue(buf, o.DataHolder[k]); err != nil {
			return err
		}
	}
	return nil
}

// no leading type marker
func encodeArrayBody(buf *bytes.Buffer, a *GFSArray) error {
	putU16(buf, uint16(len(a.DataHolder)))
	for _, w := range a.DataHolder {
		if err := encodeValue(buf, w); err != nil {
			return err
		}
	}
	return nil
}

// encodeValue writes its type byte followed by payload.
func encodeValue(buf *bytes.Buffer, w *GFSDataWrapper) error {
	id := w.TypeID.TypeID
	buf.WriteByte(id)

	switch id {
	case NULL.TypeID:
		// no payload
	case BOOL.TypeID:
		buf.WriteByte(boolToByte(w.Data.(bool)))
	case BYTE.TypeID:
		buf.WriteByte(w.Data.(byte))
	case SHORT.TypeID:
		putU16(buf, uint16(w.Data.(int16)))
	case INT.TypeID:
		putU32(buf, uint32(int32(w.Data.(int))))
	case LONG.TypeID:
		putU64(buf, uint64(w.Data.(int64)))
	case FLOAT.TypeID:
		putU32(buf, math.Float32bits(w.Data.(float32)))
	case DOUBLE.TypeID:
		putU64(buf, math.Float64bits(w.Data.(float64)))
	case UTF_STRING.TypeID:
		putUTF(buf, w.Data.(string))
	case TEXT.TypeID:
		s := w.Data.(string)
		putU32(buf, uint32(len(s)))
		buf.WriteString(s)
	case BOOL_ARRAY.TypeID:
		a := w.Data.([]bool)
		putU16(buf, uint16(len(a)))
		for _, v := range a {
			buf.WriteByte(boolToByte(v))
		}
	case BYTE_ARRAY.TypeID:
		a := w.Data.([]byte)
		putU32(buf, uint32(len(a)))
		buf.Write(a)
	case SHORT_ARRAY.TypeID:
		a := w.Data.([]int16)
		putU16(buf, uint16(len(a)))
		for _, v := range a {
			putU16(buf, uint16(v))
		}
	case INT_ARRAY.TypeID:
		a := w.Data.([]int)
		putU16(buf, uint16(len(a)))
		for _, v := range a {
			putU32(buf, uint32(int32(v)))
		}
	case LONG_ARRAY.TypeID:
		a := w.Data.([]int64)
		putU16(buf, uint16(len(a)))
		for _, v := range a {
			putU64(buf, uint64(v))
		}
	case FLOAT_ARRAY.TypeID:
		a := w.Data.([]float32)
		putU16(buf, uint16(len(a)))
		for _, v := range a {
			putU32(buf, math.Float32bits(v))
		}
	case DOUBLE_ARRAY.TypeID:
		a := w.Data.([]float64)
		putU16(buf, uint16(len(a)))
		for _, v := range a {
			putU64(buf, math.Float64bits(v))
		}
	case UTF_STRING_ARRAY.TypeID:
		a := w.Data.([]string)
		putU16(buf, uint16(len(a)))
		for _, v := range a {
			putUTF(buf, v)
		}
	case GFS_ARRAY.TypeID:
		return encodeArrayBody(buf, w.Data.(*GFSArray))
	case GFS_OBJECT.TypeID:
		return encodeObjectBody(buf, w.Data.(*GFSObject))
	default:
		return fmt.Errorf("gofox2x: cannot serialize unsupported type %q (id %d)", w.TypeID.Name, id)
	}
	return nil
}

/*
* Decoding:
*
 */

func decodeObjectBody(r *bytes.Reader) (*GFSObject, error) {
	count, err := readU16(r)
	if err != nil {
		return nil, err
	}
	o := MakeGFSObject()
	for i := 0; i < int(count); i++ {
		key, err := readUTF(r)
		if err != nil {
			return nil, err
		}
		value, err := decodeValue(r)
		if err != nil {
			return nil, err
		}
		o.DataHolder[key] = value
	}
	return o, nil
}

func decodeArrayBody(r *bytes.Reader) (*GFSArray, error) {
	count, err := readU16(r)
	if err != nil {
		return nil, err
	}
	a := MakeGFSArray()
	for i := 0; i < int(count); i++ {
		value, err := decodeValue(r)
		if err != nil {
			return nil, err
		}
		a.DataHolder = append(a.DataHolder, value)
	}
	return a, nil
}

func decodeValue(r *bytes.Reader) (*GFSDataWrapper, error) {
	id, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	wrap := func(data interface{}, t GFSDataType) (*GFSDataWrapper, error) {
		return &GFSDataWrapper{Data: data, TypeID: t}, nil
	}

	switch id {
	case NULL.TypeID:
		return wrap(nil, NULL)
	case BOOL.TypeID:
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		return wrap(b != 0, BOOL)
	case BYTE.TypeID:
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		return wrap(b, BYTE)
	case SHORT.TypeID:
		v, err := readU16(r)
		if err != nil {
			return nil, err
		}
		return wrap(int16(v), SHORT)
	case INT.TypeID:
		v, err := readU32(r)
		if err != nil {
			return nil, err
		}
		return wrap(int(int32(v)), INT)
	case LONG.TypeID:
		v, err := readU64(r)
		if err != nil {
			return nil, err
		}
		return wrap(int64(v), LONG)
	case FLOAT.TypeID:
		v, err := readU32(r)
		if err != nil {
			return nil, err
		}
		return wrap(math.Float32frombits(v), FLOAT)
	case DOUBLE.TypeID:
		v, err := readU64(r)
		if err != nil {
			return nil, err
		}
		return wrap(math.Float64frombits(v), DOUBLE)
	case UTF_STRING.TypeID:
		s, err := readUTF(r)
		if err != nil {
			return nil, err
		}
		return wrap(s, UTF_STRING)
	case TEXT.TypeID:
		n, err := readU32(r)
		if err != nil {
			return nil, err
		}
		b, err := readBytes(r, int(n))
		if err != nil {
			return nil, err
		}
		return wrap(string(b), TEXT)
	case BOOL_ARRAY.TypeID:
		n, err := readU16(r)
		if err != nil {
			return nil, err
		}
		if err := ensure(r, int(n)); err != nil {
			return nil, err
		}
		arr := make([]bool, n)
		for i := range arr {
			b, err := r.ReadByte()
			if err != nil {
				return nil, err
			}
			arr[i] = b != 0
		}
		return wrap(arr, BOOL_ARRAY)
	case BYTE_ARRAY.TypeID:
		n, err := readU32(r)
		if err != nil {
			return nil, err
		}
		b, err := readBytes(r, int(n))
		if err != nil {
			return nil, err
		}
		return wrap(b, BYTE_ARRAY)
	case SHORT_ARRAY.TypeID:
		n, err := readU16(r)
		if err != nil {
			return nil, err
		}
		if err := ensure(r, int(n)*2); err != nil {
			return nil, err
		}
		arr := make([]int16, n)
		for i := range arr {
			v, err := readU16(r)
			if err != nil {
				return nil, err
			}
			arr[i] = int16(v)
		}
		return wrap(arr, SHORT_ARRAY)
	case INT_ARRAY.TypeID:
		n, err := readU16(r)
		if err != nil {
			return nil, err
		}
		if err := ensure(r, int(n)*4); err != nil {
			return nil, err
		}
		arr := make([]int, n)
		for i := range arr {
			v, err := readU32(r)
			if err != nil {
				return nil, err
			}
			arr[i] = int(int32(v))
		}
		return wrap(arr, INT_ARRAY)
	case LONG_ARRAY.TypeID:
		n, err := readU16(r)
		if err != nil {
			return nil, err
		}
		if err := ensure(r, int(n)*8); err != nil {
			return nil, err
		}
		arr := make([]int64, n)
		for i := range arr {
			v, err := readU64(r)
			if err != nil {
				return nil, err
			}
			arr[i] = int64(v)
		}
		return wrap(arr, LONG_ARRAY)
	case FLOAT_ARRAY.TypeID:
		n, err := readU16(r)
		if err != nil {
			return nil, err
		}
		if err := ensure(r, int(n)*4); err != nil {
			return nil, err
		}
		arr := make([]float32, n)
		for i := range arr {
			v, err := readU32(r)
			if err != nil {
				return nil, err
			}
			arr[i] = math.Float32frombits(v)
		}
		return wrap(arr, FLOAT_ARRAY)
	case DOUBLE_ARRAY.TypeID:
		n, err := readU16(r)
		if err != nil {
			return nil, err
		}
		if err := ensure(r, int(n)*8); err != nil {
			return nil, err
		}
		arr := make([]float64, n)
		for i := range arr {
			v, err := readU64(r)
			if err != nil {
				return nil, err
			}
			arr[i] = math.Float64frombits(v)
		}
		return wrap(arr, DOUBLE_ARRAY)
	case UTF_STRING_ARRAY.TypeID:
		n, err := readU16(r)
		if err != nil {
			return nil, err
		}
		arr := make([]string, n)
		for i := range arr {
			s, err := readUTF(r)
			if err != nil {
				return nil, err
			}
			arr[i] = s
		}
		return wrap(arr, UTF_STRING_ARRAY)
	case GFS_ARRAY.TypeID:
		a, err := decodeArrayBody(r)
		if err != nil {
			return nil, err
		}
		return wrap(a, GFS_ARRAY)
	case GFS_OBJECT.TypeID:
		o, err := decodeObjectBody(r)
		if err != nil {
			return nil, err
		}
		return wrap(o, GFS_OBJECT)
	default:
		return nil, fmt.Errorf("gofox2x: unknown or unsupported type id %d", id)
	}
}

/*
* Helpers:
*
 */

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func putU16(buf *bytes.Buffer, v uint16) {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], v)
	buf.Write(b[:])
}

func putU32(buf *bytes.Buffer, v uint32) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	buf.Write(b[:])
}

func putU64(buf *bytes.Buffer, v uint64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], v)
	buf.Write(b[:])
}

func putUTF(buf *bytes.Buffer, s string) {
	putU16(buf, uint16(len(s)))
	buf.WriteString(s)
}

func readU16(r *bytes.Reader) (uint16, error) {
	var b [2]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b[:]), nil
}

func readU32(r *bytes.Reader) (uint32, error) {
	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b[:]), nil
}

func readU64(r *bytes.Reader) (uint64, error) {
	var b [8]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b[:]), nil
}

func readUTF(r *bytes.Reader) (string, error) {
	n, err := readU16(r)
	if err != nil {
		return "", err
	}
	b, err := readBytes(r, int(n))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// checks n against bytes remaining to avoid corrupt length fields causing unwanted allocations
func readBytes(r *bytes.Reader, n int) ([]byte, error) {
	if err := ensure(r, n); err != nil {
		return nil, err
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}
	return b, nil
}

func ensure(r *bytes.Reader, n int) error {
	if n < 0 || n > r.Len() {
		return fmt.Errorf("gofox2x: declared length %d exceeds %d remaining bytes", n, r.Len())
	}
	return nil
}
