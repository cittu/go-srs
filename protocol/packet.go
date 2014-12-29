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
    "github.com/winlinvip/go-srs/core"
    "bytes"
)

func DiscoveryPacket(msg *RtmpMessage, logger core.Logger) (b []byte, pkt RtmpPacket, err error) {
    header := msg.Header
    b = msg.Payload

    if msg == nil || len(msg.Payload) == 0 {
        logger.Info("ignore empty msg")
        return
    }

    // decode specified packet type
    if header.IsAmf0Command() || header.IsAmf3Command() || header.IsAmf0Data() || header.IsAmf3Data() {
        logger.Info("start to decode AMF0/AMF3 command message.")

        // skip 1bytes to decode the amf3 command.
        if header.IsAmf3Command() {
            b = b[1:]
            logger.Info("skip 1bytes to decode AMF3 command")
        }

        // amf0 command message.
        // need to read the command name.
        var command Amf0String
        if command,err = ParseAmf0String(bytes.NewBuffer(b)); err != nil {
            logger.Error("decode AMF0/AMF3 command name failed.")
            return
        }
        logger.Info("AMF0/AMF3 command message, command_name=%v", command)

        // result/error packet
        if command == RTMP_AMF0_COMMAND_RESULT || command == RTMP_AMF0_COMMAND_ERROR {
        }

        // decode command object.
        switch command {
        case RTMP_AMF0_COMMAND_CONNECT:
            logger.Info("decode the AMF0/AMF3 command(connect vhost/app message).")
            pkt = NewRtmpConnectAppPacket()
        }
    } else if header.IsUserControlMessage() {
    } else if header.IsWindowAckledgementSize() {
    } else if header.IsSetChunkSize() {
    } else {
    }

    return
}

// the rtmp packet, decoded from rtmp message payload.
type RtmpPacket interface {
    Decode(buffer *bytes.Buffer, logger core.Logger) error
}

type RtmpCommonCallPacket struct {
    /**
    * Name of the remote procedure that is called.
    */
    CommandName Amf0String
    /**
    * If a response is expected we give a transaction Id. Else we pass a value of 0
    */
    TransactionId Amf0Number
}

func (ccp *RtmpCommonCallPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if ccp.CommandName,err = ParseAmf0String(buffer); err != nil {
        return
    }
    if ccp.TransactionId,err = ParseAmf0Number(buffer); err != nil {
        return
    }
    return
}

/**
* 4.1.1. connect
* The client sends the connect command to the server to request
* connection to a server application instance.
*/
type RtmpConnectAppPacket struct {
    RtmpCommonCallPacket
    /**
    * Command information object which has the name-value pairs.
    * @remark: alloc in packet constructor, user can directly use it,
    *       user should never alloc it again which will cause memory leak.
    * @remark, never be nil.
    */
    CommandObject *Amf0Object
    /**
    * Any optional information
    * @remark, optional, init to and maybe nil.
    */
    Arguments *Amf0Object
}

func NewRtmpConnectAppPacket() *RtmpConnectAppPacket {
    return &RtmpConnectAppPacket{
        CommandObject: NewAmf0Object(),
        Arguments: nil,
    }
}

func (cap *RtmpConnectAppPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = cap.RtmpCommonCallPacket.Decode(buffer, logger); err != nil {
        return
    }

    // some client donot send id=1.0, so we only warn user if not match.
    if cap.TransactionId != 1.0 {
        logger.Warn("connect should be 1.0, actual is %v", cap.TransactionId)
    }

    if err = cap.CommandObject.Decode(buffer); err != nil {
        logger.Error("amf0 decode connect command_object failed.")
        return
    }

    if buffer.Len() > 0 {
        // see: https://github.com/winlinvip/simple-rtmp-server/issues/186
        // the args maybe any amf0, for instance, a string. we should drop if not object.
        var any Amf0Any
        if any,err = ParseAmf0Any(buffer); err != nil {
            logger.Error("amf0 decode connect args failed")
            return
        }

        // drop when not an AMF0 object.
        if any.(*Amf0Object) == nil {
            logger.Warn("drop the args, see: '4.1.1. connect'")
        } else {
            cap.Arguments = any.(*Amf0Object)
        }
    }

    logger.Info("amf0 decode connect packet success")

    return
}

/**
* 4.1.2. Call
* The call method of the NetConnection object runs remote procedure
* calls (RPC) at the receiving end. The called RPC name is passed as a
* parameter to the call command.
*/
type RtmpCallPacket struct {
    RtmpCommonCallPacket
    /**
    * If there exists any command info this
    * is set, else this is set to null type.
    * @remark, optional, init to and maybe nil.
    */
    CommandObject Amf0Any
    /**
    * Any optional arguments to be provided
    * @remark, optional, init to and maybe nil.
    */
    Arguments Amf0Any
}

func (cp *RtmpCallPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    return
}
