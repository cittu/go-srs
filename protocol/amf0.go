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
var Amf0StringMarkerCheck = errors.New("amf0 check string marker failed.")
var Amf0StringLengthRead = errors.New("amf0 read string length failed")
var Amf0StringDataRead = errors.New("amf0 read string data failed")
var Amf0NumberMarkerRead = errors.New("amf0 read number marker failed.")
var Amf0NumberMarkerCheck = errors.New("amf0 check number marker failed.")
var Amf0NumberValueRead = errors.New("amf0 read number value failed.")
var Amf0AnyMarkerRead = errors.New("amf0 read marker failed.")
var Amf0AnyMarkerCheck = errors.New("amf0 invalid marker.")
var Amf0BooleanMarkerRead = errors.New("amf0 read bool marker failed.")
var Amf0BooleanMarkerCheck = errors.New("amf0 check bool marker failed.")
var Amf0BooleanValueRead = errors.New("amf0 read bool value failed.")
var Amf0NullMarkerRead = errors.New("amf0 read null marker failed.")
var Amf0NullMarkerCheck = errors.New("amf0 check null marker failed.")
var Amf0UndefinedMarkerRead = errors.New("amf0 read undefined marker failed.")
var Amf0UndefinedMarkerCheck = errors.New("amf0 check undefined marker failed.")
var Amf0ObjectMarkerRead = errors.New("amf0 read object marker failed.")
var Amf0ObjectMarkerCheck = errors.New("amf0 check object marker failed.")
var Amf0ObjectEofRequired = errors.New("amf0 required object eof.")

type Amf0String string

func ParseAmf0String(buffer *bytes.Buffer) (v Amf0String, err error) {
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
    if utf8,err = parseAmf0Utf8(buffer); err != nil {
        return
    }

    v = Amf0String(utf8)

    return
}

type Amf0Number float64

func ParseAmf0Number(buffer *bytes.Buffer) (v Amf0Number, err error) {
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

type Amf0Boolean bool

func ParseAmf0Boolean(buffer *bytes.Buffer) (v Amf0Boolean, err error) {
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

type Amf0Null byte

func ParseAmf0Null(buffer *bytes.Buffer) (err error) {
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

type Amf0Undefined byte

func ParseAmf0Undefined(buffer *bytes.Buffer) (err error) {
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

type amf0Property struct {
    name string
    value Amf0Any
}

type Amf0Object struct {
    properties map[string]Amf0Any
    sorted_properties []*amf0Property
}

func NewAmf0Object() *Amf0Object {
    return &Amf0Object {
        properties: make(map[string]Amf0Any),
        sorted_properties: make([]*amf0Property, 0),
    }
}

func (obj *Amf0Object) GetString(name string) (v Amf0String, ok bool) {
    var any Amf0Any
    if any,ok = obj.properties[name]; !ok {
        return
    }
    v,ok = any.(Amf0String)
    return
}

func (obj *Amf0Object) GetNumber(name string) (v Amf0Number, ok bool) {
    var any Amf0Any
    if any,ok = obj.properties[name]; !ok {
        return
    }
    v,ok = any.(Amf0Number)
    return
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
        if name,err = parseAmf0Utf8(buffer); err != nil {
            return
        }

        var value Amf0Any
        if value,err = ParseAmf0Any(buffer); err != nil {
            return
        }

        obj.properties[name] = value
        prop := &amf0Property{
            name: name,
            value: value,
        }
        obj.sorted_properties = append(obj.sorted_properties, prop)
    }
    return
}

type Amf0Any interface {}

func ParseAmf0Any(buffer *bytes.Buffer) (v Amf0Any, err error) {
    if buffer.Len()  == 0 {
        err = Amf0AnyMarkerRead
        return
    }
    // peek the marker
    marker := buffer.Bytes()[0]

    switch marker {
    case RTMP_AMF0_String:
        v,err = ParseAmf0String(buffer)
    case RTMP_AMF0_Number:
        v,err = ParseAmf0Number(buffer)
    case RTMP_AMF0_Boolean:
        v,err = ParseAmf0Boolean(buffer)
    case RTMP_AMF0_Null:
        v = Amf0Null(0)
        err = ParseAmf0Null(buffer)
    case RTMP_AMF0_Undefined:
        v = Amf0Undefined(0)
        err = ParseAmf0Undefined(buffer)
    case RTMP_AMF0_Object:
        v = NewAmf0Object()
        err = v.(*Amf0Object).Decode(buffer)
    default:
        err = Amf0AnyMarkerCheck
    }

    return
}

func parseAmf0Utf8(buffer *bytes.Buffer) (v string, err error) {
    // len
    var len int16
    if err = binary.Read(buffer, binary.BigEndian, &len); err != nil {
        err = Amf0StringLengthRead
        return
    }

    // empty string
    if len <= 0 {
        return
    }

    // data
    data := make([]byte, len)
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
