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
)

type Conn struct {
	Server *Server
	IoRw *net.TCPConn
	Logger core.Logger
	Rand *rand.Rand // the random to generate the handshake bytes.
}

func (conn *Conn) Serve() {
	defer conn.IoRw.Close()
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
}

func NewConn(svr *Server, conn *net.TCPConn) *Conn {
	return &Conn{
		Server: svr,
		IoRw: conn,
		Logger: svr.Factory.CreateLogger("conn"),
		Rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
