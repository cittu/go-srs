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

import (
	"net"
	"github.com/winlinvip/go-srs/core"
	"math/rand"
	"time"
	"runtime"
	"io"
)

type Conn struct {
	Server *Server
	IoRw *net.TCPConn
	Logger core.Logger
	Rand *rand.Rand // the random to generate the handshake bytes.
	InChannel chan *RtmpMessage // the incoming messages channel
	OutChannel chan *RtmpMessage // the outgoing messages channel
	Protocol *Protocol // the protocol stack.
}

func (conn *Conn) Serve() {
	defer func(){
		// any error for each connection must be recover
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			conn.Logger.Error("rtmp: panic serving %v\n%v\n%s", conn.IoRw.RemoteAddr(), err, buf)
		}
		conn.IoRw.Close()
	}()
	conn.Logger.Trace("serve client ip=%v", conn.IoRw.RemoteAddr().String())

	if err := conn.IoRw.SetNoDelay(false); err != nil {
		conn.Logger.Error("tcp SetNoDelay failed, err is %v", err)
		return
	}
	conn.Logger.Info("tcp SetNoDelay ok")

	hs := SimpleHandshake{}
	if err := hs.WithClient(conn); err != nil {
		conn.Logger.Error("handshake failed, err is %v", err)
		return
	}
	conn.Logger.Trace("simple handshake with client ok")

	// pump message goroutine
	go conn.pumpMessage()

	// rtmp msg loop
	if err := conn.messageCycle(); err != nil {
		conn.Logger.Error("message cycle failed, err is %v", err)
		return
	}
	conn.Logger.Trace("serve conn ok")
}

func (conn *Conn) messageCycle() error {
	for {
		select {
			// when incoming message, process it.
		case msg,ok := <- conn.InChannel:
			if !ok {
				return nil
			}
			conn.Logger.Info("received msg %v", msg)
			// TODO: FIXME: to process the msg.
			continue
			// when got msg to send, send it immeidately.
		case msg,ok := <- conn.OutChannel:
			if !ok {
				return nil
			}
			conn.Logger.Info("send msg %v", msg)
			// TODO: FIXME: to sendout the msg.
			continue
		}
	}
}

func (conn *Conn) pumpMessage() {
	defer func(){
		// any error for each connection msg pump must be recover
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			conn.Logger.Error("rtmp: panic pump message %v\n%v\n%s", conn.IoRw.RemoteAddr(), err, buf)
		}

		conn.Logger.Info("close the incoming channel")
		close(conn.InChannel)

		conn.Logger.Info("stop pump rtmp messages")
	}()
	conn.Logger.Info("start pump rtmp messages")

	for {
		var msg *RtmpMessage
		var err error
		if msg,err = conn.Protocol.PumpMessage(); err != nil {
			if err != io.EOF {
				conn.Logger.Error("pump message failed, err is %v", err)
			}
			return
		}

		if msg == nil {
			continue
		}

		conn.Logger.Info("incoming msg")
		conn.InChannel <- msg
	}
}

func NewConn(svr *Server, conn *net.TCPConn) *Conn {
	v := &Conn{
		Server: svr,
		IoRw: conn,
		Logger: svr.Factory.CreateLogger("conn"),
		Rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// TODO: FIXME: channel with buffer
	v.InChannel = make(chan *RtmpMessage)
	v.OutChannel = make(chan *RtmpMessage)

	// initialize the protocol stack.
	v.Protocol = NewProtocol(conn, v.Logger)

	return v
}
