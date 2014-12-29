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
)

type Server struct {
	Addr string
	Factory Factory
	Logger core.Logger
}

func (svr *Server) ListenAndServe() error {
	addr := svr.Addr
	if len(addr) == 0 {
		addr = ":1935"
	}
	svr.Logger.Trace("server addr is %v", addr)

	ln, err := net.Listen("tcp", addr)
	if err  != nil {
		return err
	}
	svr.Logger.Trace("listen at %v ok", addr)

	return svr.Serve(ln)
}

func (svr *Server) Serve(l net.Listener) error {
	defer l.Close()
	for {
		rw, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			}
			return err
		}

		c := NewConn(svr, rw.(*net.TCPConn))
		go c.Serve()
	}
}

func ListenAndServe(addr string, factory Factory) error {
	server := &Server{Addr: addr, Factory: factory}
	server.Logger = factory.CreateLogger("server")
	return server.ListenAndServe()
}
