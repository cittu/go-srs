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
    "github.com/winlinvip/go-srs/core"
    "github.com/winlinvip/go-srs/protocol"
    "errors"
    "net"
)

var FinalStage = errors.New("rtmp final stage")

type commonStage struct {
    logger core.Logger
    conn *protocol.Conn
}

type connectStage struct {
    commonStage
}

func (stage *connectStage) ConsumeMessage(msg *protocol.RtmpMessage) (err error) {
    // always expect the connect app message.
    if !msg.Header.IsAmf0Command() && !msg.Header.IsAmf3Command() {
        return
    }

    var pkt protocol.RtmpPacket
    if pkt,err = stage.conn.Protocol.DecodeMessage(msg); err != nil {
        return
    }

    // got connect app packet
    if pkt,ok := pkt.(*protocol.RtmpConnectAppPacket); ok {
        req := &stage.conn.Request
        if err = req.Parse(pkt.CommandObject, pkt.Arguments, stage.logger); err != nil {
            stage.logger.Error("parse request from connect app packet failed.")
            return
        }
        stage.logger.Info("rtmp connect app success")

        // discovery vhost, resolve the vhost from config
        // TODO: FIXME: implements it

        // check the request paramaters.
        if err = req.Validate(stage.logger); err != nil {
            return
        }
        stage.logger.Info("discovery app success. schema=%v, vhost=%v, port=%v, app=%v",
            req.Schema, req.Vhost, req.Port, req.App)

        // check vhost
        // TODO: FIXME: implements it
        stage.logger.Info("check vhost success.")

        stage.logger.Trace("connect app, tcUrl=%v, pageUrl=%v, swfUrl=%v, schema=%v, vhost=%v, port=%v, app=%v, args=%v",
            req.TcUrl, req.PageUrl, req.SwfUrl, req.Schema, req.Vhost, req.Port, req.App, req.FormatArgs())

        // show client identity
        si := SrsInfo{}
        si.Parse(req.Args)
        if si.SrsPid > 0 {
            stage.logger.Trace("edge-srs ip=%v, version=%v, pid=%v, id=%v",
                si.SrsServerIp, si.SrsVersion, si.SrsPid, si.SrsId)
        }

        if err = stage.conn.SetWindowAckSize(2.5 * 1000 * 1000); err != nil {
            stage.logger.Error("set window acknowledgement size failed.")
            return
        }
        stage.logger.Info("set window acknowledgement size success")

        if err = stage.conn.SetPeerBandwidth(2.5 * 1000 * 1000, 2); err != nil {
            stage.logger.Error("set peer bandwidth failed.")
            return
        }
        stage.logger.Info("set peer bandwidth success")

        // get the ip which client connected.
        var iorw *net.TCPConn = stage.conn.IoRw
        localIp := iorw.LocalAddr().String()

        // do bandwidth test if connect to the vhost which is for bandwidth check.
        // TODO: FIXME: implements it

        // do token traverse before serve it.
        // @see https://github.com/winlinvip/simple-rtmp-server/pull/239
        // TODO: FIXME: implements it

        // response the client connect ok.
        if err = stage.conn.ResponseConnectApp(req.ObjectEncoding, localIp); err != nil {
            stage.logger.Error("response connect app failed.")
            return
        }
        stage.logger.Info("response connect app success")

        if err = stage.conn.OnBwDone(); err != nil {
            stage.logger.Error("on_bw_done failed.")
            return
        }
        stage.logger.Info("on_bw_done success")

        // use next stage.
        stage.conn.Stage = NewIdentifyClientStage(stage.conn)
    }

    return
}

type identifyClientStage struct {
    commonStage
}

func NewIdentifyClientStage(conn *protocol.Conn) protocol.Stage {
    return &identifyClientStage{
        commonStage:commonStage{
            logger:conn.Logger,
            conn: conn,
        },
    }
}

func (stage *identifyClientStage) ConsumeMessage(msg *protocol.RtmpMessage) (err error) {
    h := &msg.Header
    if h.IsAckledgement() || h.IsSetChunkSize() || h.IsWindowAckledgementSize() || h.IsUserControlMessage() {
        stage.logger.Info("ignore the ack/setChunkSize/windowAck/userControl msg")
        return
    }
    if !h.IsAmf0Command() && !h.IsAmf3Command() {
        stage.logger.Trace("identify ignore messages except AMF0/AMF3 command message, type=%d", h.MessageType)
        return
    }

    var pkt protocol.RtmpPacket
    if pkt,err = stage.conn.Protocol.DecodeMessage(msg); err != nil {
        stage.logger.Error("identify decode message failed")
        return
    }

    if pkt,ok := pkt.(*protocol.RtmpCreateStreamPacket); ok {
        stage.logger.Info("identify client by create stream, play or flash publish.")
        if err = stage.conn.ResponseCreateStream(float64(pkt.TransactionId), stage.conn.StreamId); err != nil {
            stage.logger.Error("send createStream response message failed.")
            return
        }
        stage.logger.Info("send createStream response message success.")

        // use next stage.
        stage.conn.Stage = NewIdentifyClientCreateStreamStage(stage.conn)
    }

    // use next stage.
    stage.conn.Stage = NewFinalStage(stage.conn)
    return
}

type identifyClientCreateStreamStage struct {
    commonStage
}

func NewIdentifyClientCreateStreamStage(conn *protocol.Conn) protocol.Stage {
    return &identifyClientCreateStreamStage{
        commonStage:commonStage{
            logger:conn.Logger,
            conn: conn,
        },
    }
}

func (stage *identifyClientCreateStreamStage) ConsumeMessage(msg *protocol.RtmpMessage) (err error) {
    return FinalStage
}

type finalStage struct {
    commonStage
}

func NewFinalStage(conn *protocol.Conn) protocol.Stage {
    return &finalStage{
        commonStage:commonStage{
            logger:conn.Logger,
            conn: conn,
        },
    }
}

func (stage *finalStage) ConsumeMessage(msg *protocol.RtmpMessage) (err error) {
    return FinalStage
}
