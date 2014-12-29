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
    "io"
    "errors"
    "encoding/binary"
    "bytes"
    "time"
    "net"
)

var RtmpPlainRequired = errors.New("only support rtmp plain text")

type SimpleHandshake struct {
}

func (hs *SimpleHandshake) WithClient(conn *Conn) error {
    var iorw *net.TCPConn = conn.IoRw

    c0c1 := make([]byte, 1537)
    conn.Logger.Info("read c0c1 from conn, size=%v", len(c0c1))
    if _,err := io.ReadFull(iorw, c0c1); err != nil {
        conn.Logger.Error("read c0c1 failed, err is %v", err)
        return err
    }
    conn.Logger.Info("read c0c1 ok")

    if c0c1[0] != 0x03 {
        conn.Logger.Error("rtmp plain required 0x03, actual is %#x", c0c1[0])
        return RtmpPlainRequired
    }
    conn.Logger.Info("check rtmp plain protocol ok")

    // use bytes buffer to write content.
    s0s1s2 := bytes.NewBuffer(make([]byte, 0, 3073))

    // plain text required.
    binary.Write(s0s1s2, binary.BigEndian, byte(0x03))
    // s1 time
    binary.Write(s0s1s2, binary.BigEndian, int32(time.Now().Unix()))
    // s1 time2 copy from c1
    if _,err := s0s1s2.Write(c0c1[1:5]); err != nil {
        conn.Logger.Error("copy c0c1 time to s0s1s2 failed, err is %v", err)
        return err
    }
    // s1 1528 random bytes
    s0s1s2Random := make([]byte, 1528)
    RandomGenerate(conn.Rand, s0s1s2Random)
    if _,err := s0s1s2.Write(s0s1s2Random); err != nil {
        conn.Logger.Error("fill s1 random bytes failed, err is %v", err)
        return err
    }
    // if c1 specified, copy c1 to s2.
    // @see: https://github.com/winlinvip/simple-rtmp-server/issues/46
    if _,err := s0s1s2.Write(c0c1[1:1537]); err != nil {
        conn.Logger.Error("copy c1 to s1 failed, err is %v", err)
        return err
    }
    conn.Logger.Info("generate s0s1s2 ok, buf=%d", s0s1s2.Len())

    if written,err := iorw.Write(s0s1s2.Bytes()); err != nil {
        conn.Logger.Error("send s0s1s2 failed, written=%d, err is %v", written, err)
        return err
    }
    conn.Logger.Info("send s0s1s2 ok")

    c2 := make([]byte, 1536)
    conn.Logger.Info("read c2 from conn, size=%v", len(c2))
    if _,err := io.ReadFull(iorw, c2); err != nil {
        conn.Logger.Error("read c2 failed, err is %v", err)
        return err
    }
    conn.Logger.Info("read c2 ok")

    return nil
}
