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
    "encoding/binary"
    "errors"
    "fmt"
    "os"
)

var RtmpMsgSwaspRead = errors.New("decode ack window size failed.")
var RtmpMsgSpbpRead = errors.New("decode set peer bandwidth failed.")

const (
    /**
    * the signature for packets to client.
    */
    RTMP_SIG_FMS_VER = "3,5,3,888"
    RTMP_SIG_AMF0_VER = 0
    RTMP_SIG_CLIENT_ID = "ASAICiss"

    /**
    * onStatus consts.
    */
    StatusLevel = "level"
    StatusCode = "code"
    StatusDescription = "description"
    StatusDetails = "details"
    StatusClientId = "clientid"
    // status value
    StatusLevelStatus = "status"
    // status error
    StatusLevelError = "error"
    // code value
    StatusCodeConnectSuccess = "NetConnection.Connect.Success"
    StatusCodeConnectRejected = "NetConnection.Connect.Rejected"
    StatusCodeStreamReset = "NetStream.Play.Reset"
    StatusCodeStreamStart = "NetStream.Play.Start"
    StatusCodeStreamPause = "NetStream.Pause.Notify"
    StatusCodeStreamUnpause = "NetStream.Unpause.Notify"
    StatusCodePublishStart = "NetStream.Publish.Start"
    StatusCodeDataStart = "NetStream.Data.Start"
    StatusCodeUnpublishSuccess = "NetStream.Unpublish.Success"

    // FMLE
    RTMP_AMF0_COMMAND_ON_FC_PUBLISH = "onFCPublish"
    RTMP_AMF0_COMMAND_ON_FC_UNPUBLISH = "onFCUnpublish"

    // default stream id for response the createStream request.
    SRS_DEFAULT_SID = 1
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
        if command,err = DecodeAmf0String(bytes.NewBuffer(b)); err != nil {
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
    // decode methods
    Decode(buffer *bytes.Buffer, logger core.Logger) error
    // encode methods
    Encode(buffer *bytes.Buffer, logger core.Logger) error
    MessageType() byte
    PerferCid() int
}

type rtmpCommonCallPacket struct {
    /**
    * Name of the remote procedure that is called.
    */
    CommandName Amf0String
    /**
    * If a response is expected we give a transaction Id. Else we pass a value of 0
    */
    TransactionId Amf0Number
}

func (pkt *rtmpCommonCallPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if pkt.CommandName,err = DecodeAmf0String(buffer); err != nil {
        return
    }
    if pkt.TransactionId,err = DecodeAmf0Number(buffer); err != nil {
        return
    }
    return
}

func (pkt *rtmpCommonCallPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = EncodeAmf0String(buffer, pkt.CommandName); err != nil {
        return
    }
    if err = EncodeAmf0Number(buffer, pkt.TransactionId); err != nil {
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
    rtmpCommonCallPacket
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

func NewRtmpConnectAppPacket() RtmpPacket {
    return &RtmpConnectAppPacket{
        CommandObject: NewAmf0Object(),
        Arguments: nil,
    }
}

func (pkt *RtmpConnectAppPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Decode(buffer, logger); err != nil {
        return
    }

    // some client donot send id=1.0, so we only warn user if not match.
    if pkt.TransactionId != 1.0 {
        logger.Warn("connect should be 1.0, actual is %v", pkt.TransactionId)
    }

    if err = pkt.CommandObject.Decode(buffer); err != nil {
        logger.Error("amf0 decode connect command_object failed.")
        return
    }

    if buffer.Len() > 0 {
        // see: https://github.com/winlinvip/simple-rtmp-server/issues/186
        // the args maybe any amf0, for instance, a string. we should drop if not object.
        var any Amf0Any
        if any,err = DecodeAmf0Any(buffer); err != nil {
            logger.Error("amf0 decode connect args failed")
            return
        }

        // drop when not an AMF0 object.
        if any.(*Amf0Object) == nil {
            logger.Warn("drop the args, see: '4.1.1. connect'")
        } else {
            pkt.Arguments = any.(*Amf0Object)
        }
    }

    logger.Info("amf0 decode connect packet success")

    return
}

func (pkt *RtmpConnectAppPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    return
}

func (pkt *RtmpConnectAppPacket) MessageType() byte {
    return RTMP_MSG_AMF0CommandMessage
}

func (pkt *RtmpConnectAppPacket) PerferCid() int {
    return RTMP_CID_OverConnection
}

/**
* response for SrsConnectAppPacket.
*/
type RtmpConnectAppResPacket struct {
    rtmpCommonCallPacket
    /**
    * Name-value pairs that describe the properties(fmsver etc.) of the connection.
    * @remark, never be NULL.
    */
    Props *Amf0Object
    /**
    * Name-value pairs that describe the response from|the server. ‘code’,
    * ‘level’, ‘description’ are names of few among such information.
    * @remark, never be NULL.
    */
    Info *Amf0Object
}

func NewRtmpConnectAppResPacket(objectEncoding int, serverIp string, srsId int) RtmpPacket {
    v := &RtmpConnectAppResPacket{
        Props: NewAmf0Object(),
        Info: NewAmf0Object(),
    }
    v.Props.Set("fmsVer", Amf0String(fmt.Sprintf("FMS/%v", RTMP_SIG_FMS_VER)))
    v.Props.Set("capabilities", Amf0Number(127))
    v.Props.Set("mode", Amf0Number(1))

    v.Props.Set(StatusLevel, Amf0String(StatusLevelStatus))
    v.Props.Set(StatusCode, Amf0String(StatusCodeConnectSuccess))
    v.Props.Set(StatusDescription, Amf0String("Connection succeeded"))
    v.Props.Set("objectEncoding", Amf0Number(objectEncoding))

    data := NewAmf0EcmaArray()
    v.Props.Set("data", data)
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
    data.Set("srs_authors", Amf0String(core.RTMP_SIG_SRS_AUTHROS))

    if serverIp != "" {
        data.Set("srs_server_ip", Amf0String(serverIp))
    }
    // for edge to directly get the id of client.
    data.Set("srs_pid", Amf0Number(os.Getpid()))
    data.Set("srs_id", Amf0Number(srsId))

    return v
}

func (pkt *RtmpConnectAppResPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Decode(buffer, logger); err != nil {
        return
    }
    if err = pkt.Props.Decode(buffer); err != nil {
        return
    }
    if err = pkt.Info.Decode(buffer); err != nil {
        return
    }
    return
}

func (pkt *RtmpConnectAppResPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Encode(buffer, logger); err != nil {
        return
    }
    if err = pkt.Props.Encode(buffer); err != nil {
        return
    }
    if err = pkt.Info.Encode(buffer); err != nil {
        return
    }
    return
}

func (pkt *RtmpConnectAppResPacket) MessageType() byte {
    return RTMP_MSG_AMF0CommandMessage
}

func (pkt *RtmpConnectAppResPacket) PerferCid() int {
    return RTMP_CID_OverConnection
}

/**
* 4.1.2. Call
* The call method of the NetConnection object runs remote procedure
* calls (RPC) at the receiving end. The called RPC name is passed as a
* parameter to the call command.
*/
type RtmpCallPacket struct {
    rtmpCommonCallPacket
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

func (pkt *RtmpCallPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    return
}

func (pkt *RtmpCallPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    return
}

func (pkt *RtmpCallPacket) MessageType() byte {
    return RTMP_MSG_AMF0CommandMessage
}

func (pkt *RtmpCallPacket) PerferCid() int {
    return RTMP_CID_OverConnection
}

/**
* 5.5. Window Acknowledgement Size (5)
* The client or the server sends this message to inform the peer which
* window size to use when sending acknowledgment.
*/
type RtmpSetWindowAckSizePacket struct {
    AckowledgementWindowSize int32
}

func NewRtmpSetWindowAckSizePacket(ackSize int) RtmpPacket {
    return &RtmpSetWindowAckSizePacket{
        AckowledgementWindowSize: int32(ackSize),
    }
}

func (pkt *RtmpSetWindowAckSizePacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = binary.Read(buffer, binary.BigEndian, &pkt.AckowledgementWindowSize); err != nil {
        return RtmpMsgSwaspRead
    }
    return
}

func (pkt *RtmpSetWindowAckSizePacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    return
}

func (pkt *RtmpSetWindowAckSizePacket) MessageType() byte {
    return RTMP_MSG_WindowAcknowledgementSize
}

func (pkt *RtmpSetWindowAckSizePacket) PerferCid() int {
    return RTMP_CID_ProtocolControl
}

/**
* 5.6. Set Peer Bandwidth (6)
* The client or the server sends this message to update the output
* bandwidth of the peer.
*/
type RtmpSetPeerBandwidthPacket struct {
    Bandwidth int32
    // @see: SrsPeerBandwidthType
    Type byte
}

func NewRtmpSetPeerBandwidthPacket(bandwidth, _type int) RtmpPacket {
    return &RtmpSetPeerBandwidthPacket{
        Bandwidth: int32(bandwidth),
        Type: byte(_type),
    }
}

func (pkt *RtmpSetPeerBandwidthPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = binary.Read(buffer, binary.BigEndian, &pkt.Bandwidth); err != nil {
        return RtmpMsgSpbpRead
    }
    if err = binary.Read(buffer, binary.BigEndian, &pkt.Type); err != nil {
        return RtmpMsgSpbpRead
    }
    return
}

func (pkt *RtmpSetPeerBandwidthPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    return
}

func (pkt *RtmpSetPeerBandwidthPacket) MessageType() byte {
    return RTMP_MSG_SetPeerBandwidth
}

func (pkt *RtmpSetPeerBandwidthPacket) PerferCid() int {
    return RTMP_CID_ProtocolControl
}

/**
* when bandwidth test done, notice client.
*/
type RtmpOnBWDonePacket struct {
    rtmpCommonCallPacket
    /**
    * Command information does not exist. Set to null type.
    * @remark, never be NULL, an AMF0 null instance.
    */
    Args Amf0Null
}

func NewRtmpOnBWDonePacket() RtmpPacket {
    v := &RtmpOnBWDonePacket{}
    v.CommandName = Amf0String(RTMP_AMF0_COMMAND_ON_BW_DONE)
    return v
}

func (pkt *RtmpOnBWDonePacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Decode(buffer, logger); err != nil {
        return
    }
    if err = DecodeAmf0Null(buffer); err != nil {
        return
    }
    return
}

func (pkt *RtmpOnBWDonePacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    return
}

func (pkt *RtmpOnBWDonePacket) MessageType() byte {
    return RTMP_MSG_AMF0CommandMessage
}

func (pkt *RtmpOnBWDonePacket) PerferCid() int {
    return RTMP_CID_OverConnection
}
