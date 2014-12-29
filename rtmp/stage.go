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

package rtmp

import "github.com/winlinvip/go-srs/core"

type Stage interface {
    ConsumeMessage(msg *RtmpMessage) error
}

type commonStage struct {
    logger core.Logger
}

type connectStage struct {
    commonStage
    conn *Conn
}

func NewConnectStage(conn *Conn) Stage {
    return &connectStage{
        commonStage:commonStage{
            logger:conn.Logger,
        },
        conn: conn,
    }
}

func (cs *connectStage) ConsumeMessage(msg *RtmpMessage) (err error) {
    // always expect the connect app message.
    if !msg.Header.IsAmf0Command() && !msg.Header.IsAmf3Command() {
        return
    }

    var pkt RtmpPacket
    if pkt,err = cs.conn.Protocol.DecodeMessage(msg); err != nil || pkt == nil {
        return err
    }

    return
}
