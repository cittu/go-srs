/*
The MIT License (MIT)

Copyright (c) 2013-2014 winlin

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package protocol

import (
    "bytes"
    "errors"
    "encoding/binary"
)

// AMF0 marker
const (
    RTMP_AMF0_Number = iota
    RTMP_AMF0_Boolean
    RTMP_AMF0_String
    RTMP_AMF0_Object
    RTMP_AMF0_MovieClip // reserved, not supported
    RTMP_AMF0_Null
    RTMP_AMF0_Undefined
    RTMP_AMF0_Reference
    RTMP_AMF0_EcmaArray
    RTMP_AMF0_ObjectEnd
    RTMP_AMF0_StrictArray
    RTMP_AMF0_Date
    RTMP_AMF0_LongString
    RTMP_AMF0_UnSupported
    RTMP_AMF0_RecordSet // reserved, not supported
    RTMP_AMF0_XmlDocument
    RTMP_AMF0_TypedObject
    // AVM+ object is the AMF3 object.
    RTMP_AMF0_AVMplusObject
    // origin array whos data takes the same form as LengthValueBytes
    RTMP_AMF0_OriginStrictArray = 0x20
    // User defined
    RTMP_AMF0_Invalid = 0x3F
)

var Amf0StringMarkerRead = errors.New("amf0 read string marker failed.")
var Amf0StringMarkerWrite = errors.New("amf0 write string marker failed.")
var Amf0StringMarkerCheck = errors.New("amf0 check string marker failed.")
var Amf0StringLengthRead = errors.New("amf0 read string length failed")
var Amf0StringLengthWrite = errors.New("amf0 write string length failed")
var Amf0StringDataRead = errors.New("amf0 read string data failed")
var Amf0StringDataWrite= errors.New("amf0 write string data failed")
var Amf0NumberMarkerRead = errors.New("amf0 read number marker failed.")
var Amf0NumberMarkerWrite = errors.New("amf0 write number marker failed.")
var Amf0NumberMarkerCheck = errors.New("amf0 check number marker failed.")
var Amf0NumberValueRead = errors.New("amf0 read number value failed.")
var Amf0NumberValueWrite = errors.New("amf0 write number value failed.")
var Amf0AnyMarkerRead = errors.New("amf0 read marker failed.")
var Amf0AnyMarkerCheck = errors.New("amf0 invalid marker.")
var Amf0BooleanMarkerRead = errors.New("amf0 read bool marker failed.")
var Amf0BooleanMarkerWrite = errors.New("amf0 write bool marker failed.")
var Amf0BooleanMarkerCheck = errors.New("amf0 check bool marker failed.")
var Amf0BooleanValueRead = errors.New("amf0 read bool value failed.")
var Amf0NullMarkerRead = errors.New("amf0 read null marker failed.")
var Amf0NullMarkerWrite = errors.New("amf0 write null marker failed.")
var Amf0NullMarkerCheck = errors.New("amf0 check null marker failed.")
var Amf0UndefinedMarkerRead = errors.New("amf0 read undefined marker failed.")
var Amf0UndefinedMarkerWrite= errors.New("amf0 write undefined marker failed.")
var Amf0UndefinedMarkerCheck = errors.New("amf0 check undefined marker failed.")
var Amf0ObjectMarkerRead = errors.New("amf0 read object marker failed.")
var Amf0ObjectMarkerWrite = errors.New("amf0 write object marker failed.")
var Amf0ObjectMarkerCheck = errors.New("amf0 check object marker failed.")
var Amf0ObjectEofRequired = errors.New("amf0 required object eof.")
var Amf0EcmaArrayMarkerRead = errors.New("amf0 read ecma array marker failed.")
var Amf0EcmaArrayMarkerWrite = errors.New("amf0 write ecma array marker failed.")
var Amf0EcmaArrayMarkerCheck = errors.New("amf0 check ecma array marker failed.")
var Amf0EcmaArrayCountRead = errors.New("amf0 read ecma array value failed.")
var Amf0EcmaArrayCountWrite = errors.New("amf0 write ecma array value failed.")
var Amf0EcmaArrayEofRequired = errors.New("amf0 required ecma array eof.")

type Amf0String string

func DecodeAmf0String(buffer *bytes.Buffer) (v Amf0String, err error) {
    // marker
    var marker byte
    if marker,err = buffer.ReadByte(); err != nil {
        err = Amf0StringMarkerRead
        return
    }

    if marker != RTMP_AMF0_String {
        err = Amf0StringMarkerCheck
        return
    }

    var utf8 string
    if utf8,err = decodeAmf0Utf8(buffer); err != nil {
        return
    }

    v = Amf0String(utf8)

    return
}

func EncodeAmf0String(buffer *bytes.Buffer, v Amf0String) (err error) {
    if err = buffer.WriteByte(RTMP_AMF0_String); err != nil {
        err = Amf0StringMarkerWrite
        return
    }
    if err = encodeAmf0Utf8(buffer, string(v)); err != nil {
        return
    }
    return
}

type Amf0Number float64

func DecodeAmf0Number(buffer *bytes.Buffer) (v Amf0Number, err error) {
    // marker
    var marker byte
    if marker,err = buffer.ReadByte(); err != nil {
        err = Amf0NumberMarkerRead
        return
    }

    if marker != RTMP_AMF0_Number {
        err = Amf0NumberMarkerCheck
        return
    }

    var data float64
    if err = binary.Read(buffer, binary.BigEndian, &data); err != nil {
        err = Amf0NumberValueRead
        return
    }

    v = Amf0Number(data)

    return
}

func EncodeAmf0Number(buffer *bytes.Buffer, v Amf0Number) (err error) {
    if err = buffer.WriteByte(RTMP_AMF0_Number); err != nil {
        err = Amf0NumberMarkerWrite
        return
    }
    data := float64(v)
    if err = binary.Write(buffer, binary.BigEndian, data); err != nil {
        err = Amf0NumberValueWrite
        return
    }
    return
}

type Amf0Boolean bool

func DecodeAmf0Boolean(buffer *bytes.Buffer) (v Amf0Boolean, err error) {
    // marker
    var marker byte
    if marker,err = buffer.ReadByte(); err != nil {
        err = Amf0BooleanMarkerRead
        return
    }

    if marker != RTMP_AMF0_Boolean {
        err = Amf0BooleanMarkerCheck
        return
    }

    var data byte
    if data,err = buffer.ReadByte(); err != nil {
        err = Amf0BooleanValueRead
        return
    }

    v = Amf0Boolean(data != 0)

    return
}

func EncodeAmf0Boolean(buffer *bytes.Buffer, v Amf0Boolean) (err error) {
    if err = buffer.WriteByte(RTMP_AMF0_Boolean); err != nil {
        err = Amf0BooleanMarkerWrite
        return
    }

    if bool(v) {
        return buffer.WriteByte(byte(1))
    } else {
        return buffer.WriteByte(byte(0))
    }
}

type Amf0Null byte

func DecodeAmf0Null(buffer *bytes.Buffer) (err error) {
    // marker
    var marker byte
    if marker,err = buffer.ReadByte(); err != nil {
        err = Amf0NullMarkerRead
        return
    }

    if marker != RTMP_AMF0_Null {
        err = Amf0NullMarkerCheck
        return
    }

    return
}

func EncodeAmf0Null(buffer *bytes.Buffer) (err error) {
    if err = buffer.WriteByte(RTMP_AMF0_Null); err != nil {
        err = Amf0NullMarkerWrite
        return
    }
    return
}

type Amf0Undefined byte

func DecodeAmf0Undefined(buffer *bytes.Buffer) (err error) {
    // marker
    var marker byte
    if marker,err = buffer.ReadByte(); err != nil {
        err = Amf0UndefinedMarkerRead
        return
    }

    if marker != RTMP_AMF0_Undefined {
        err = Amf0UndefinedMarkerCheck
        return
    }

    return
}

func EncodeAmf0Undefined(buffer *bytes.Buffer) (err error) {
    if err = buffer.WriteByte(RTMP_AMF0_Undefined); err != nil {
        err = Amf0UndefinedMarkerWrite
        return
    }
    return
}

type amf0Property struct {
    name string
    value Amf0Any
}

type amf0SortedDict struct {
    properties map[string]Amf0Any
    sorted_properties []*amf0Property
}

func (sd *amf0SortedDict) Set(name string, value Amf0Any) {
    // remove the elem when exists
    if _,ok := sd.properties[name]; ok {
        delete(sd.properties, name)
        for i,v := range sd.sorted_properties {
            if v.name == name {
                sd.sorted_properties = append(sd.sorted_properties[:i], sd.sorted_properties[i+1:]...)
                break
            }
        }
    }

    // append to object.
    sd.properties[name] = value
    prop := &amf0Property{
        name: name,
        value: value,
    }
    sd.sorted_properties = append(sd.sorted_properties, prop)
}

func (sd *amf0SortedDict) GetString(name string) (v Amf0String, ok bool) {
    var any Amf0Any
    if any,ok = sd.properties[name]; !ok {
        return
    }
    v,ok = any.(Amf0String)
    return
}

func (sd *amf0SortedDict) GetNumber(name string) (v Amf0Number, ok bool) {
    var any Amf0Any
    if any,ok = sd.properties[name]; !ok {
        return
    }
    v,ok = any.(Amf0Number)
    return
}

func (sd *amf0SortedDict) Encode(buffer *bytes.Buffer) (err error) {
    for _,v := range sd.sorted_properties {
        if err = encodeAmf0Utf8(buffer, v.name); err != nil {
            return
        }
        if err = EncodeAmf0Any(buffer, v.value); err != nil {
            return
        }
    }
    return
}

/**
* 2.10 ECMA Array Type
* ecma-array-type = associative-count *(object-property)
* associative-count = U32
* object-property = (UTF-8 value-type) | (UTF-8-empty object-end-marker)
*/
type Amf0EcmaArray struct {
    properties amf0SortedDict
    count int32
}

func NewAmf0EcmaArray() *Amf0EcmaArray {
    v := &Amf0EcmaArray {}
    v.properties.properties = make(map[string]Amf0Any)
    v.properties.sorted_properties = make([]*amf0Property, 0)
    return v
}

func (arr *Amf0EcmaArray) Set(name string, value Amf0Any) {
    arr.properties.Set(name, value)
}

func (arr *Amf0EcmaArray) GetString(name string) (v Amf0String, ok bool) {
    return arr.properties.GetString(name)
}

func (arr *Amf0EcmaArray) GetNumber(name string) (v Amf0Number, ok bool) {
    return arr.properties.GetNumber(name)
}

func (arr *Amf0EcmaArray) Decode(buffer *bytes.Buffer) (err error) {
    // marker
    var marker byte
    if marker,err = buffer.ReadByte(); err != nil {
        err = Amf0EcmaArrayMarkerRead
        return
    }

    if marker != RTMP_AMF0_EcmaArray {
        err = Amf0EcmaArrayMarkerCheck
        return
    }

    // ecma array count
    if err = binary.Read(buffer, binary.BigEndian, &arr.count); err != nil {
        err = Amf0EcmaArrayCountRead
        return
    }

    // ecma array properties
    for buffer.Len() > 0 {
        // atleast an object EOF
        if buffer.Len() < 3 {
            return Amf0EcmaArrayEofRequired
        }
        // peek the marker
        marker := buffer.Bytes()[2]

        // read object EOF.
        if marker == RTMP_AMF0_ObjectEnd {
            _,err = buffer.Read(make([]byte, 3))
            return
        }

        // read an object property
        var name string
        if name,err = decodeAmf0Utf8(buffer); err != nil {
            return
        }

        var value Amf0Any
        if value,err = DecodeAmf0Any(buffer); err != nil {
            return
        }

        arr.Set(name, value)
    }
    return
}

func (arr *Amf0EcmaArray) Encode(buffer *bytes.Buffer) (err error) {
    // marker
    if err = buffer.WriteByte(RTMP_AMF0_EcmaArray); err != nil {
        err = Amf0EcmaArrayMarkerWrite
        return
    }

    // ecma array count
    if err = binary.Write(buffer, binary.BigEndian, arr.count); err != nil {
        err = Amf0EcmaArrayCountWrite
        return
    }

    // arr properties
    if err = arr.properties.Encode(buffer); err != nil {
        return
    }

    // object EOF
    if _,err = buffer.Write([]byte{0x00, 0x00, RTMP_AMF0_ObjectEnd}); err != nil {
        return
    }
    return
}

/**
* 2.5 Object Type
* anonymous-object-type = object-marker *(object-property)
* object-property = (UTF-8 value-type) | (UTF-8-empty object-end-marker)
*/
type Amf0Object struct {
    properties amf0SortedDict
}

func NewAmf0Object() *Amf0Object {
    v := &Amf0Object {}
    v.properties.properties = make(map[string]Amf0Any)
    v.properties.sorted_properties = make([]*amf0Property, 0)
    return v
}

func (obj *Amf0Object) Set(name string, value Amf0Any) {
    obj.properties.Set(name, value)
}

func (obj *Amf0Object) GetString(name string) (v Amf0String, ok bool) {
    return obj.properties.GetString(name)
}

func (obj *Amf0Object) GetNumber(name string) (v Amf0Number, ok bool) {
    return obj.properties.GetNumber(name)
}

func (obj *Amf0Object) Decode(buffer *bytes.Buffer) (err error) {
    // marker
    var marker byte
    if marker,err = buffer.ReadByte(); err != nil {
        err = Amf0ObjectMarkerRead
        return
    }

    if marker != RTMP_AMF0_Object {
        err = Amf0ObjectMarkerCheck
        return
    }

    // object properties
    for buffer.Len() > 0 {
        // atleast an object EOF
        if buffer.Len() < 3 {
            return Amf0ObjectEofRequired
        }
        // peek the marker
        marker := buffer.Bytes()[2]

        // read object EOF.
        if marker == RTMP_AMF0_ObjectEnd {
            _,err = buffer.Read(make([]byte, 3))
            return
        }

        // read an object property
        var name string
        if name,err = decodeAmf0Utf8(buffer); err != nil {
            return
        }

        var value Amf0Any
        if value,err = DecodeAmf0Any(buffer); err != nil {
            return
        }

        obj.Set(name, value)
    }
    return
}

func (obj *Amf0Object) Encode(buffer *bytes.Buffer) (err error) {
    // marker
    if err = buffer.WriteByte(RTMP_AMF0_Object); err != nil {
        err = Amf0ObjectMarkerWrite
        return
    }

    // object properties
    if err = obj.properties.Encode(buffer); err != nil {
        return
    }

    // object EOF
    if _,err = buffer.Write([]byte{0x00, 0x00, RTMP_AMF0_ObjectEnd}); err != nil {
        return
    }
    return
}

type Amf0Any interface {}

func DecodeAmf0Any(buffer *bytes.Buffer) (v Amf0Any, err error) {
    if buffer.Len()  == 0 {
        err = Amf0AnyMarkerRead
        return
    }
    // peek the marker
    marker := buffer.Bytes()[0]

    switch marker {
    case RTMP_AMF0_String:
        v,err = DecodeAmf0String(buffer)
    case RTMP_AMF0_Number:
        v,err = DecodeAmf0Number(buffer)
    case RTMP_AMF0_Boolean:
        v,err = DecodeAmf0Boolean(buffer)
    case RTMP_AMF0_Null:
        v = Amf0Null(0)
        err = DecodeAmf0Null(buffer)
    case RTMP_AMF0_Undefined:
        v = Amf0Undefined(0)
        err = DecodeAmf0Undefined(buffer)
    case RTMP_AMF0_Object:
        v = NewAmf0Object()
        err = v.(*Amf0Object).Decode(buffer)
    case RTMP_AMF0_EcmaArray:
        v = NewAmf0EcmaArray()
        err = v.(*Amf0EcmaArray).Decode(buffer)
    default:
        err = Amf0AnyMarkerCheck
    }

    return
}

func EncodeAmf0Any(buffer *bytes.Buffer, v Amf0Any) (err error) {
    switch v := v.(type) {
    case Amf0String:
        err = EncodeAmf0String(buffer, v)
    case Amf0Number:
        err = EncodeAmf0Number(buffer, v)
    case Amf0Boolean:
        err = EncodeAmf0Boolean(buffer, v)
    case Amf0Null:
        err = EncodeAmf0Null(buffer)
    case Amf0Undefined:
        err = EncodeAmf0Undefined(buffer)
    case Amf0Object:
        err = v.Encode(buffer)
    case *Amf0Object:
        err = v.Encode(buffer)
    case Amf0EcmaArray:
        err = v.Encode(buffer)
    case *Amf0EcmaArray:
        err = v.Encode(buffer)
    default:
        err = Amf0AnyMarkerCheck
    }

    return
}

func decodeAmf0Utf8(buffer *bytes.Buffer) (v string, err error) {
    // len
    var length int16
    if err = binary.Read(buffer, binary.BigEndian, &length); err != nil {
        err = Amf0StringLengthRead
        return
    }

    // empty string
    if length <= 0 {
        return
    }

    // data
    data := make([]byte, length)
    if _,err = buffer.Read(data); err != nil {
        err = Amf0StringDataRead
        return
    }

    // support utf8-1 only
    // 1.3.1 Strings and UTF-8
    // UTF8-1 = %x00-7F
    // TODO: support other utf-8 strings

    v = string(data)

    return
}

func encodeAmf0Utf8(buffer *bytes.Buffer, v string) (err error) {
    // len
    length := int16(len(v))
    if err = binary.Write(buffer, binary.BigEndian, length); err != nil {
        err = Amf0StringLengthWrite
        return
    }

    // empty string
    if length <= 0 {
        return
    }

    // data
    if _,err = buffer.Write([]byte(v)); err != nil {
        err = Amf0StringDataWrite
        return
    }

    return
}
