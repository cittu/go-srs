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
	"fmt"
	"os"
)

var RtmpInChannelMsg = errors.New("put msg to channel failed")
var RtmpControlRepublish = errors.New("encoder republish stream")

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
	StreamId int // current using stream id.
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
	for {
		if err := conn.recvMessage(); err != nil {
			if err == RtmpControlRepublish {
				conn.Stage = conn.Server.Factory.NewIdenfityStage(conn)
				conn.Logger.Trace("control message(unpublish) accept, retry stream service.")
				continue
			}

			conn.Logger.Error("message cycle failed, err is %v", err)
			break
		}
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

func (conn *Conn) ResponseCreateStream(transactionId float64, streamId int) (err error) {
	var msg *RtmpMessage
	if msg,err = conn.Protocol.EncodeMessage(NewRtmpCreateStreamResPacket(transactionId, float64(streamId)), 0); err != nil {
		return
	}
	return conn.EnqueueOutgoingMessage(msg)
}

func (conn *Conn) OnBwDone() (err error) {
	pkt := NewRtmpCallPacket().(*RtmpCallPacket)
	pkt.CommandName = Amf0String(RTMP_AMF0_COMMAND_ON_BW_DONE)
	pkt.CommandObject = Amf0Null(0)

	var msg *RtmpMessage
	if msg,err = conn.Protocol.EncodeMessage(pkt, 0); err != nil {
		return
	}
	return conn.EnqueueOutgoingMessage(msg)
}

func (conn *Conn) ResponseConnectApp(objectEncoding int, serverIp string) (err error) {
	v := NewRtmpConnectAppResPacket().(*RtmpConnectAppResPacket)
	v.CommandName = Amf0String(RTMP_AMF0_COMMAND_RESULT)
	v.TransactionId = Amf0Number(1.0)

	v.Props.Set("fmsVer", Amf0String(fmt.Sprintf("FMS/%v", RTMP_SIG_FMS_VER)))
	v.Props.Set("capabilities", Amf0Number(127))
	v.Props.Set("mode", Amf0Number(1))

	v.Info.Set(StatusLevel, Amf0String(StatusLevelStatus))
	v.Info.Set(StatusCode, Amf0String(StatusCodeConnectSuccess))
	v.Info.Set(StatusDescription, Amf0String("Connection succeeded"))
	v.Info.Set("objectEncoding", Amf0Number(objectEncoding))

	data := NewAmf0EcmaArray()
	v.Info.Set("data", data)
	data.Set("version", Amf0String(RTMP_SIG_FMS_VER))
	data.Set("srs_sig", Amf0String(core.RTMP_SIG_SRS_KEY))
	data.Set("srs_server", Amf0String(fmt.Sprintf("%v %v (%v)",
			core.RTMP_SIG_SRS_KEY, core.RTMP_SIG_SRS_VERSION, core.RTMP_SIG_SRS_URL_SHORT)))
	data.Set("srs_license", Amf0String(core.RTMP_SIG_SRS_LICENSE))
	data.Set("srs_role", Amf0String(core.RTMP_SIG_SRS_ROLE))
	data.Set("srs_url", Amf0String(core.RTMP_SIG_SRS_URL))
	data.Set("srs_version", Amf0String(core.RTMP_SIG_SRS_VERSION))
	data.Set("srs_site", Amf0String(core.RTMP_SIG_SRS_WEB))
	data.Set("srs_email", Amf0String(core.RTMP_SIG_SRS_EMAIL))
	data.Set("srs_copyright", Amf0String(core.RTMP_SIG_SRS_COPYRIGHT))
	data.Set("srs_primary", Amf0String(core.RTMP_SIG_SRS_PRIMARY))

	if serverIp != "" {
		data.Set("srs_server_ip", Amf0String(serverIp))
	}
	// for edge to directly get the id of client.
	data.Set("srs_pid", Amf0Number(os.Getpid()))
	data.Set("srs_id", Amf0Number(conn.SrsId))

	var msg *RtmpMessage
	if msg,err = conn.Protocol.EncodeMessage(v, 0); err != nil {
		return
	}
	return conn.EnqueueOutgoingMessage(msg)
}

func (conn *Conn) ResponseReleaseStream(transactionId float64) (err error) {
	pkt := NewRtmpCallPacket().(*RtmpCallPacket)
	pkt.CommandName = Amf0String(RTMP_AMF0_COMMAND_RESULT)
	pkt.TransactionId = Amf0Number(transactionId)
	pkt.CommandObject = Amf0Null(0)
	pkt.Arguments = Amf0Undefined(0)

	var msg *RtmpMessage
	if msg,err = conn.Protocol.EncodeMessage(pkt, 0); err != nil {
		return
	}
	return conn.EnqueueOutgoingMessage(msg)
}

func (conn *Conn) ResponseFcPublish(transactionId float64) (err error) {
	pkt := NewRtmpCallPacket().(*RtmpCallPacket)
	pkt.CommandName = Amf0String(RTMP_AMF0_COMMAND_RESULT)
	pkt.TransactionId = Amf0Number(transactionId)
	pkt.CommandObject = Amf0Null(0)
	pkt.Arguments = Amf0Undefined(0)

	var msg *RtmpMessage
	if msg,err = conn.Protocol.EncodeMessage(pkt, 0); err != nil {
		return
	}
	return conn.EnqueueOutgoingMessage(msg)
}

func (conn *Conn) ResponsePublish(transactionId float64, streamId int) (err error) {
	pkt := NewRtmpCallPacket().(*RtmpCallPacket)
	pkt.CommandName = Amf0String(RTMP_AMF0_COMMAND_ON_FC_PUBLISH)
	pkt.TransactionId = Amf0Number(0.0)
	pkt.CommandObject = Amf0Null(0)

	obj := NewAmf0Object()
	pkt.Arguments = obj
	obj.Set(StatusCode, Amf0String(StatusCodePublishStart))
	obj.Set(StatusDescription, Amf0String("Started publising stream."))

	var msg *RtmpMessage
	if msg,err = conn.Protocol.EncodeMessage(pkt, streamId); err != nil {
		return
	}
	return conn.EnqueueOutgoingMessage(msg)
}

func (conn *Conn) OnStatus(streamId int) (err error) {
	pkt := NewRtmpOnStatusCallPacket().(*RtmpOnStatusCallPacket)
	pkt.Data.Set(StatusLevel, Amf0String(StatusLevelStatus))
	pkt.Data.Set(StatusCode, Amf0String(StatusCodePublishStart))
	pkt.Data.Set(StatusDescription, Amf0String("Started publishing stream."))
	pkt.Data.Set(StatusClientId, Amf0String(RTMP_SIG_CLIENT_ID))

	var msg *RtmpMessage
	if msg,err = conn.Protocol.EncodeMessage(pkt, streamId); err != nil {
		return
	}
	return conn.EnqueueOutgoingMessage(msg)
}

func NewConn(svr *Server, conn *net.TCPConn) *Conn {
	v := &Conn{
		Server: svr,
		IoRw: conn,
		Rand: rand.New(rand.NewSource(time.Now().UnixNano())),
		StreamId: SRS_DEFAULT_SID - 1, // we always increase it when create stream.
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
