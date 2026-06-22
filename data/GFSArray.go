package data

import (
	"fmt"
	"strings"
)

type GFSArray struct {
	DataHolder []*GFSDataWrapper
}

func (s *GFSArray) String() string {
	return s.Dump(0)
}

func (s *GFSArray) Dump(indents int) string {
	dump := fmt.Sprintf("(%s):", strings.ToLower(GFS_ARRAY.Name))
	for _, value := range s.DataHolder {
		dump += "\n" + value.Dump("", indents+1)
	}
	return dump
}

func (s *GFSArray) Size() int {
	return len(s.DataHolder)
}

func (s *GFSArray) Get(index int) *GFSDataWrapper {
	if index < 0 || index >= len(s.DataHolder) {
		return nil
	}
	return s.DataHolder[index]
}

func (s *GFSArray) Add(value interface{}) {
	s.AddObject(GFSDataWrapper{
		Data:   value,
		TypeID: GFSDataTypeFromVar(value),
	})
}

func (s *GFSArray) AddObject(wrapper GFSDataWrapper) {
	s.DataHolder = append(s.DataHolder, &wrapper)
}

func (s *GFSArray) add(value interface{}, typeID GFSDataType) {
	s.DataHolder = append(s.DataHolder, &GFSDataWrapper{Data: value, TypeID: typeID})
}

func (s *GFSArray) AddNull()                      { s.add(nil, NULL) }
func (s *GFSArray) AddBool(value bool)            { s.add(value, BOOL) }
func (s *GFSArray) AddByte(value byte)            { s.add(value, BYTE) }
func (s *GFSArray) AddShort(value int16)          { s.add(value, SHORT) }
func (s *GFSArray) AddInt(value int)              { s.add(value, INT) }
func (s *GFSArray) AddLong(value int64)           { s.add(value, LONG) }
func (s *GFSArray) AddFloat(value float32)        { s.add(value, FLOAT) }
func (s *GFSArray) AddDouble(value float64)       { s.add(value, DOUBLE) }
func (s *GFSArray) AddUtfString(value string)     { s.add(value, UTF_STRING) }
func (s *GFSArray) AddSFSArray(value *GFSArray)   { s.add(value, GFS_ARRAY) }
func (s *GFSArray) AddSFSObject(value *GFSObject) { s.add(value, GFS_OBJECT) }

func (s *GFSArray) RemoveObject(wrapper *GFSDataWrapper) {
	newDataHolder := make([]*GFSDataWrapper, 0, len(s.DataHolder))
	for _, value := range s.DataHolder {
		if value != wrapper {
			newDataHolder = append(newDataHolder, value)
		}
	}
	s.DataHolder = newDataHolder
}

func MakeGFSArray() *GFSArray {
	return &GFSArray{}
}
