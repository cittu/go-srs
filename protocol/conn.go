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
	"net"
	"math/rand"
	"time"
	"runtime"
	"io"
	"errors"
	"github.com/winlinvip/go-srs/core"
)

var RtmpInChannelMsg = errors.New("put msg to channel failed")

type Conn struct {
	SrsId int
	Server *Server
	IoRw *net.TCPConn
	Logger core.Logger
	Rand *rand.Rand // the random to generate the handshake bytes.
	InChannel chan *RtmpMessage // the incoming messages channel
	OutChannel chan *RtmpMessage // the outgoing messages channel
	SendQuitChannel chan int // the quit signal for connection serve cycle
	Protocol *Protocol // the protocol stack.
	Stage Stage // the stage of connection.
	Request RtmpRequest // the request of client
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

		// the out channel is write by this goroutine only
		close(conn.OutChannel)

		// quit.
		conn.IoRw.Close()
		conn.Logger.Info("conn quit")
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

	// set stage to connect app.
	conn.Stage = conn.Server.Factory.NewConnectStage(conn)

	// pump and send message goroutine
	go conn.pumpMessage()
	go conn.sendMessage()

	// rtmp msg loop
	if err := conn.recvMessage(); err != nil {
		conn.Logger.Error("message cycle failed, err is %v", err)
		return
	}
	conn.Logger.Trace("serve conn ok")
}

func (conn *Conn) recvMessage() (err error) {
	for {
		select {
			// the send message goroutine will close this channel when error
		case <- conn.SendQuitChannel:
			return
			// when incoming message, process it.
			// the pump message goroutine will close this channel when error
		case msg, ok := <- conn.InChannel:
			if !ok {
				return
			}
			conn.Logger.Info("consume received msg %v", msg)
			if err = conn.Stage.ConsumeMessage(msg); err != nil {
				return
			}
			continue
		}
	}
}

func (conn *Conn) sendMessage() (err error) {
	defer func(){
		// any error for each connection msg pump must be recover
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			conn.Logger.Error("rtmp: panic send message %v\n%v\n%s", conn.IoRw.RemoteAddr(), err, buf)
		}

		// the quit channel is write by this goroutine only
		close(conn.SendQuitChannel)

		conn.Logger.Info("stop send rtmp messages")
	}()
	conn.Logger.Info("start send rtmp messages")

	for {
		select {
			// when got msg to send, send it immeidately.
		case msg,ok := <- conn.OutChannel:
			if !ok {
				return
			}
			conn.Logger.Info("send msg %v", msg)
			if err = conn.Protocol.SendMessage(msg); err != nil {
				return
			}
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

		// the in channel is write by this goroutine only
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

		select {
		case conn.InChannel <- msg:
			break
		default:
			conn.Logger.Warn("drop incoming msg for channel full")
			break
		}
	}
}

func (conn *Conn) EnqueueOutgoingMessage(msg *RtmpMessage) (err error) {
	select {
	case conn.OutChannel <- msg:
		break
	default:
		conn.Logger.Warn("drop message for channel full")
		break
	}
	return
}

func (conn *Conn) SetWindowAckSize(ackSize int) (err error) {
	var msg *RtmpMessage
	if msg,err = conn.Protocol.EncodeMessage(NewRtmpSetWindowAckSizePacket(ackSize), 0); err != nil {
		return
	}
	return conn.EnqueueOutgoingMessage(msg)
}

func (conn *Conn) SetPeerBandwidth(bandwidth, _type int) (err error) {
	var msg *RtmpMessage
	if msg,err = conn.Protocol.EncodeMessage(NewRtmpSetPeerBandwidthPacket(bandwidth, _type), 0); err != nil {
		return
	}
	return conn.EnqueueOutgoingMessage(msg)
}

func (conn *Conn) OnBwDone() (err error) {
	var msg *RtmpMessage
	if msg,err = conn.Protocol.EncodeMessage(NewRtmpOnBWDonePacket(), 0); err != nil {
		return
	}
	return conn.EnqueueOutgoingMessage(msg)
}

func (conn *Conn) ResponseConnectApp(objectEncoding int, serverIp string) (err error) {
	var msg *RtmpMessage
	if msg,err = conn.Protocol.EncodeMessage(NewRtmpConnectAppResPacket(objectEncoding, serverIp, conn.SrsId), 0); err != nil {
		return
	}
	return conn.EnqueueOutgoingMessage(msg)
}

func NewConn(svr *Server, conn *net.TCPConn) *Conn {
	v := &Conn{
		Server: svr,
		IoRw: conn,
		Rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// the srs id
	v.SrsId = svr.Factory.SrsId()
	v.Logger = svr.Factory.CreateLogger("conn", v.SrsId)

	// TODO: FIXME: channel with buffer
	v.InChannel = make(chan *RtmpMessage, 1024)
	v.OutChannel = make(chan *RtmpMessage, 1024)
	v.SendQuitChannel = make(chan int)

	// initialize the protocol stack.
	v.Protocol = NewProtocol(conn, v.Logger)

	// nil stage for handshake.
	v.Stage = nil

	return v
}
