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
    "github.com/cittu/go-srs/core"
    "bytes"
    "encoding/binary"
    "errors"
)

var RtmpMsgSwaspRead = errors.New("decode ack window size failed.")
var RtmpMsgSetPeerBandwidthRead = errors.New("decode set peer bandwidth failed.")
var RtmpMsgSetChunkSizeRead = errors.New("decode set chunk size failed.")
var RtmpMsgUserControlRead = errors.New("decode user control failed.")

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

// 3.7. User Control message
const (
    // generally, 4bytes event-data
    /**
    * The server sends this event to notify the client
    * that a stream has become functional and can be
    * used for communication. By default, this event
    * is sent on ID 0 after the application connect
    * command is successfully received from the
    * client. The event data is 4-byte and represents
    * the stream ID of the stream that became
    * functional.
    */
    SrcPCUCStreamBegin              = 0x00

    /**
    * The server sends this event to notify the client
    * that the playback of data is over as requested
    * on this stream. No more data is sent without
    * issuing additional commands. The client discards
    * the messages received for the stream. The
    * 4 bytes of event data represent the ID of the
    * stream on which playback has ended.
    */
    SrcPCUCStreamEOF                = 0x01

    /**
    * The server sends this event to notify the client
    * that there is no more data on the stream. If the
    * server does not detect any message for a time
    * period, it can notify the subscribed clients
    * that the stream is dry. The 4 bytes of event
    * data represent the stream ID of the dry stream.
    */
    SrcPCUCStreamDry                = 0x02

    /**
    * The client sends this event to inform the server
    * of the buffer size (in milliseconds) that is
    * used to buffer any data coming over a stream.
    * This event is sent before the server starts
    * processing the stream. The first 4 bytes of the
    * event data represent the stream ID and the next
    * 4 bytes represent the buffer length, in
    * milliseconds.
    */
    SrcPCUCSetBufferLength          = 0x03 // 8bytes event-data

    /**
    * The server sends this event to notify the client
    * that the stream is a recorded stream. The
    * 4 bytes event data represent the stream ID of
    * the recorded stream.
    */
    SrcPCUCStreamIsRecorded         = 0x04

    /**
    * The server sends this event to test whether the
    * client is reachable. Event data is a 4-byte
    * timestamp, representing the local server time
    * when the server dispatched the command. The
    * client responds with kMsgPingResponse on
    * receiving kMsgPingRequest.
    */
    SrcPCUCPingRequest              = 0x06

    /**
    * The client sends this event to the server in
    * response to the ping request. The event data is
    * a 4-byte timestamp, which was received with the
    * kMsgPingRequest request.
    */
    SrcPCUCPingResponse             = 0x07
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
        case RTMP_AMF0_COMMAND_CREATE_STREAM:
            logger.Info("decode the AMF0/AMF3 command(createStream message).")
            pkt = NewRtmpCreateStreamPacket()
        case RTMP_AMF0_COMMAND_PLAY:
            logger.Info("decode the AMF0/AMF3 command(paly message).")
            pkt = NewRtmpPlayPacket()
        case RTMP_AMF0_COMMAND_RELEASE_STREAM:
            logger.Info("decode the AMF0/AMF3 command(FMLE releaseStream message).")
            pkt = NewRtmpReleaseStreamPacket()
        case RTMP_AMF0_COMMAND_FC_PUBLISH:
            logger.Info("decode the AMF0/AMF3 command(FMLE FCPublish message).")
            pkt = NewRtmpFcPublishPacket()
        case RTMP_AMF0_COMMAND_UNPUBLISH:
            logger.Info("decode the AMF0/AMF3 command(FMLE FCUnpublish message).")
            pkt = NewRtmpFcUnPublishPacket()
        case RTMP_AMF0_COMMAND_PUBLISH:
            logger.Info("decode the AMF0/AMF3 command(publish message).")
            pkt = NewRtmpPublishPacket()
        default:
            if header.IsAmf0Command() || header.IsAmf3Command() {
                logger.Info("decode the AMF0/AMF3 call message.")
                pkt = NewRtmpCallPacket()
            } else {
                logger.Info("drop the AMF0/AMF3 command message, command_name=%v", command)
            }
        }
    } else if header.IsUserControlMessage() {
        logger.Info("start to decode user control message.")
        pkt = NewRtmpUserControlPacket()
    } else if header.IsWindowAckledgementSize() {
        logger.Info("start to decode set ack window size message.")
        pkt = NewRtmpSetWindowAckSizePacket(0)
    } else if header.IsSetChunkSize() {
        logger.Info("start to decode set chunk size message.")
        pkt = NewRtmpSetChunkSizePacket()
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
    if err = pkt.rtmpCommonCallPacket.Encode(buffer, logger); err != nil {
        return
    }
    if err = pkt.CommandObject.Encode(buffer); err != nil {
        return
    }
    if pkt.Arguments != nil {
        if err = pkt.Arguments.Encode(buffer); err != nil {
            return
        }
    }
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

func NewRtmpConnectAppResPacket() RtmpPacket {
    v := &RtmpConnectAppResPacket{
        Props: NewAmf0Object(),
        Info: NewAmf0Object(),
    }
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

func NewRtmpCallPacket() RtmpPacket {
    return &RtmpCallPacket{}
}

func (pkt *RtmpCallPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Decode(buffer, logger); err != nil {
        return
    }
    if buffer.Len() > 0 {
        if pkt.CommandObject, err = DecodeAmf0Any(buffer); err != nil {
            return
        }
    }
    if buffer.Len() > 0 {
        if pkt.Arguments,err = DecodeAmf0Any(buffer); err != nil {
            return
        }
    }
    return
}

func (pkt *RtmpCallPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Encode(buffer, logger); err != nil {
        return
    }
    if pkt.CommandObject != nil {
        if err = EncodeAmf0Any(buffer, pkt.CommandObject); err != nil {
            return
        }
    }
    if pkt.Arguments != nil {
        if err = EncodeAmf0Any(buffer, pkt.Arguments); err != nil {
            return
        }
    }
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
    if err = binary.Write(buffer, binary.BigEndian, pkt.AckowledgementWindowSize); err != nil {
        return
    }
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
        return RtmpMsgSetPeerBandwidthRead
    }
    if err = binary.Read(buffer, binary.BigEndian, &pkt.Type); err != nil {
        return RtmpMsgSetPeerBandwidthRead
    }
    return
}

func (pkt *RtmpSetPeerBandwidthPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = binary.Write(buffer, binary.BigEndian, pkt.Bandwidth); err != nil {
        return
    }
    if err = binary.Write(buffer, binary.BigEndian, pkt.Type); err != nil {
        return
    }
    return
}

func (pkt *RtmpSetPeerBandwidthPacket) MessageType() byte {
    return RTMP_MSG_SetPeerBandwidth
}

func (pkt *RtmpSetPeerBandwidthPacket) PerferCid() int {
    return RTMP_CID_ProtocolControl
}

/**
* 7.1. Set Chunk Size
* Protocol control message 1, Set Chunk Size, is used to notify the
* peer about the new maximum chunk size.
*/
type RtmpSetChunkSizePacket struct {
    /**
    * The maximum chunk size can be 65536 bytes. The chunk size is
    * maintained independently for each direction.
    */
    ChunkSize int32
}

func NewRtmpSetChunkSizePacket() RtmpPacket {
    return &RtmpSetChunkSizePacket{
        ChunkSize: SRS_CONSTS_RTMP_PROTOCOL_CHUNK_SIZE,
    }
}

func (pkt *RtmpSetChunkSizePacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = binary.Read(buffer, binary.BigEndian, &pkt.ChunkSize); err != nil {
        return RtmpMsgSetChunkSizeRead
    }
    return
}

func (pkt *RtmpSetChunkSizePacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = binary.Write(buffer, binary.BigEndian, pkt.ChunkSize); err != nil {
        return
    }
    return
}

func (pkt *RtmpSetChunkSizePacket) MessageType() byte {
    return RTMP_MSG_SetChunkSize
}

func (pkt *RtmpSetChunkSizePacket) PerferCid() int {
    return RTMP_CID_ProtocolControl
}

/**
* 5.4. User Control Message (4)
*
* for the EventData is 4bytes.
* Stream Begin(=0)              4-bytes stream ID
* Stream EOF(=1)                4-bytes stream ID
* StreamDry(=2)                 4-bytes stream ID
* SetBufferLength(=3)           8-bytes 4bytes stream ID, 4bytes buffer length.
* StreamIsRecorded(=4)          4-bytes stream ID
* PingRequest(=6)               4-bytes timestamp local server time
* PingResponse(=7)              4-bytes timestamp received ping request.
*
* 3.7. User Control message
* +------------------------------+-------------------------
* | Event Type ( 2- bytes ) | Event Data
* +------------------------------+-------------------------
* Figure 5 Pay load for the ‘User Control Message’.
*/
type RtmpUserControlPacket struct {
    /**
    * Event type is followed by Event data.
    * @see: SrcPCUCEventType
    */
    EventType int16
    EventData int32
    /**
    * 4bytes if event_type is SetBufferLength; otherwise 0.
    */
    ExtraData int32
}

func NewRtmpUserControlPacket() RtmpPacket {
    return &RtmpUserControlPacket{}
}

func (pkt *RtmpUserControlPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = binary.Read(buffer, binary.BigEndian, &pkt.EventType); err != nil {
        return RtmpMsgUserControlRead
    }
    if err = binary.Read(buffer, binary.BigEndian, &pkt.EventData); err != nil {
        return RtmpMsgUserControlRead
    }
    if pkt.EventType == SrcPCUCSetBufferLength {
        if err = binary.Read(buffer, binary.BigEndian, &pkt.ExtraData); err != nil {
            return
        }
    }
    return
}

func (pkt *RtmpUserControlPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = binary.Write(buffer, binary.BigEndian, pkt.EventType); err != nil {
        return
    }
    if err = binary.Write(buffer, binary.BigEndian, pkt.EventData); err != nil {
        return
    }
    if pkt.EventType == SrcPCUCSetBufferLength {
        if err = binary.Write(buffer, binary.BigEndian, pkt.ExtraData); err != nil {
            return
        }
    }
    return
}

func (pkt *RtmpUserControlPacket) MessageType() byte {
    return RTMP_MSG_UserControlMessage
}

func (pkt *RtmpUserControlPacket) PerferCid() int {
    return RTMP_CID_ProtocolControl
}

/**
* 4.1.3. createStream
* The client sends this command to the server to create a logical
* channel for message communication The publishing of audio, video, and
* metadata is carried out over stream channel created using the
* createStream command.
*/
type RtmpCreateStreamPacket struct {
    rtmpCommonCallPacket
    /**
    * If there exists any command info this is set, else this is set to null type.
    * @remark, never be NULL, an AMF0 null instance.
    */
    CommandObject Amf0Null
}

func NewRtmpCreateStreamPacket() RtmpPacket {
    v := &RtmpCreateStreamPacket{}
    v.CommandName = Amf0String(RTMP_AMF0_COMMAND_CREATE_STREAM)
    v.TransactionId = Amf0Number(2.0)
    return v
}

func (pkt *RtmpCreateStreamPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Decode(buffer, logger); err != nil {
        return
    }
    if err = DecodeAmf0Null(buffer); err != nil {
        return
    }
    return
}

func (pkt *RtmpCreateStreamPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Encode(buffer, logger); err != nil {
        return
    }
    if err = EncodeAmf0Null(buffer); err != nil {
        return
    }
    return
}

func (pkt *RtmpCreateStreamPacket) MessageType() byte {
    return RTMP_MSG_AMF0CommandMessage
}

func (pkt *RtmpCreateStreamPacket) PerferCid() int {
    return RTMP_CID_OverConnection
}

/**
* response for SrsCreateStreamPacket.
*/
type RtmpCreateStreamResPacket struct {
    rtmpCommonCallPacket
    /**
    * If there exists any command info this is set, else this is set to null type.
    * @remark, never be NULL, an AMF0 null instance.
    */
    CommandObject Amf0Null
    /**
    * The return value is either a stream ID or an error information object.
    */
    StreamId Amf0Number
}

func NewRtmpCreateStreamResPacket(transactionId, streamId float64) RtmpPacket {
    v := &RtmpCreateStreamResPacket{}
    v.CommandName = Amf0String(RTMP_AMF0_COMMAND_RESULT)
    v.TransactionId = Amf0Number(transactionId)
    v.StreamId = Amf0Number(streamId)
    return v
}

func (pkt *RtmpCreateStreamResPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Decode(buffer, logger); err != nil {
        return
    }
    if err = DecodeAmf0Null(buffer); err != nil {
        return
    }
    if pkt.StreamId,err = DecodeAmf0Number(buffer); err != nil {
        return
    }
    return
}

func (pkt *RtmpCreateStreamResPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Encode(buffer, logger); err != nil {
        return
    }
    if err = EncodeAmf0Null(buffer); err != nil {
        return
    }
    if err = EncodeAmf0Number(buffer, pkt.StreamId); err != nil {
        return
    }
    return
}

func (pkt *RtmpCreateStreamResPacket) MessageType() byte {
    return RTMP_MSG_AMF0CommandMessage
}

func (pkt *RtmpCreateStreamResPacket) PerferCid() int {
    return RTMP_CID_OverConnection
}

/**
* 4.2.1. play
* The client sends this command to the server to play a stream.
*/
type RtmpPlayPacket struct {
    rtmpCommonCallPacket
    /**
    * Command information does not exist. Set to null type.
    * @remark, never be NULL, an AMF0 null instance.
    */
    CommandObject Amf0Null
    /**
    * Name of the stream to play.
    * To play video (FLV) files, specify the name of the stream without a file
    *       extension (for example, "sample").
    * To play back MP3 or ID3 tags, you must precede the stream name with mp3:
    *       (for example, "mp3:sample".)
    * To play H.264/AAC files, you must precede the stream name with mp4: and specify the
    *       file extension. For example, to play the file sample.m4v, specify
    *       "mp4:sample.m4v"
    */
    StreamName Amf0String
    /**
    * An optional parameter that specifies the start time in seconds.
    * The default value is -2, which means the subscriber first tries to play the live
    *       stream specified in the Stream Name field. If a live stream of that name is
    *       not found, it plays the recorded stream specified in the Stream Name field.
    * If you pass -1 in the Start field, only the live stream specified in the Stream
    *       Name field is played.
    * If you pass 0 or a positive number in the Start field, a recorded stream specified
    *       in the Stream Name field is played beginning from the time specified in the
    *       Start field.
    * If no recorded stream is found, the next item in the playlist is played.
    */
    Start Amf0Number
    /**
    * An optional parameter that specifies the duration of playback in seconds.
    * The default value is -1. The -1 value means a live stream is played until it is no
    *       longer available or a recorded stream is played until it ends.
    * If u pass 0, it plays the single frame since the time specified in the Start field
    *       from the beginning of a recorded stream. It is assumed that the value specified
    *       in the Start field is equal to or greater than 0.
    * If you pass a positive number, it plays a live stream for the time period specified
    *       in the Duration field. After that it becomes available or plays a recorded
    *       stream for the time specified in the Duration field. (If a stream ends before the
    *       time specified in the Duration field, playback ends when the stream ends.)
    * If you pass a negative number other than -1 in the Duration field, it interprets the
    *       value as if it were -1.
    */
    Duration Amf0Number
    /**
    * An optional Boolean value or number that specifies whether to flush any
    * previous playlist.
    */
    Reset Amf0Boolean
}

func NewRtmpPlayPacket() RtmpPacket {
    v := &RtmpPlayPacket{}
    v.CommandName = Amf0String(RTMP_AMF0_COMMAND_PLAY)
    v.TransactionId = Amf0Number(0.0)
    v.Start = Amf0Number(-2)
    v.Duration = Amf0Number(-1)
    v.Reset = Amf0Boolean(true)
    return v
}

func (pkt *RtmpPlayPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Decode(buffer, logger); err != nil {
        return
    }
    if err = DecodeAmf0Null(buffer); err != nil {
        return
    }
    if pkt.StreamName,err = DecodeAmf0String(buffer); err != nil {
        return
    }
    if buffer.Len() > 0 {
        if pkt.Start,err = DecodeAmf0Number(buffer); err != nil {
            return
        }
    }
    if buffer.Len() > 0 {
        if pkt.Duration,err = DecodeAmf0Number(buffer); err != nil {
            return
        }
    }
    if buffer.Len() > 0 {
        if pkt.Reset,err = DecodeAmf0Boolean(buffer); err != nil {
            return
        }
    }
    return
}

func (pkt *RtmpPlayPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Encode(buffer, logger); err != nil {
        return
    }
    if err = EncodeAmf0Null(buffer); err != nil {
        return
    }
    if err = EncodeAmf0String(buffer, pkt.StreamName); err != nil {
        return
    }
    if err = EncodeAmf0Number(buffer, pkt.Start); err != nil {
        return
    }
    if err = EncodeAmf0Number(buffer, pkt.Duration); err != nil {
        return
    }
    if err = EncodeAmf0Boolean(buffer, pkt.Reset); err != nil {
        return
    }
    return
}

func (pkt *RtmpPlayPacket) MessageType() byte {
    return RTMP_MSG_AMF0CommandMessage
}

func (pkt *RtmpPlayPacket) PerferCid() int {
    return RTMP_CID_OverStream
}

/**
* FMLE start publish: ReleaseStream
*/
type RtmpReleaseStreamPacket struct {
    rtmpCommonCallPacket
    /**
    * If there exists any command info this is set, else this is set to null type.
    * @remark, never be NULL, an AMF0 null instance.
    */
    CommandObject Amf0Null
    /**
    * the stream name to start publish or release.
    */
    StreamName Amf0String
}

func NewRtmpReleaseStreamPacket() RtmpPacket {
    v := &RtmpReleaseStreamPacket{}
    v.CommandName = Amf0String(RTMP_AMF0_COMMAND_RELEASE_STREAM)
    v.TransactionId = Amf0Number(0.0)
    return v
}

func (pkt *RtmpReleaseStreamPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Decode(buffer, logger); err != nil {
        return
    }
    if err = DecodeAmf0Null(buffer); err != nil {
        return
    }
    if pkt.StreamName,err = DecodeAmf0String(buffer); err != nil {
        return
    }
    return
}

func (pkt *RtmpReleaseStreamPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Encode(buffer, logger); err != nil {
        return
    }
    if err = EncodeAmf0Null(buffer); err != nil {
        return
    }
    if err = EncodeAmf0String(buffer, pkt.StreamName); err != nil {
        return
    }
    return
}

func (pkt *RtmpReleaseStreamPacket) MessageType() byte {
    return RTMP_MSG_AMF0CommandMessage
}

func (pkt *RtmpReleaseStreamPacket) PerferCid() int {
    return RTMP_CID_OverConnection
}

/**
* FMLE start publish: FCPublish
*/
type RtmpFcPublishPacket struct {
    RtmpReleaseStreamPacket
}

func NewRtmpFcPublishPacket() RtmpPacket {
    v := &RtmpFcPublishPacket{}
    v.CommandName = Amf0String(RTMP_AMF0_COMMAND_FC_PUBLISH)
    v.TransactionId = Amf0Number(0.0)
    return v
}

/**
* FMLE stop publish: FCUnpublish
*/
type RtmpFcUnPublishPacket struct {
    RtmpReleaseStreamPacket
}

func NewRtmpFcUnPublishPacket() RtmpPacket {
    v := &RtmpFcUnPublishPacket{}
    v.CommandName = Amf0String(RTMP_AMF0_COMMAND_UNPUBLISH)
    v.TransactionId = Amf0Number(0.0)
    return v
}

/**
* FMLE/flash publish
* 4.2.6. Publish
* The client sends the publish command to publish a named stream to the
* server. Using this name, any client can play this stream and receive
* the published audio, video, and data messages.
*/
type RtmpPublishPacket struct {
    rtmpCommonCallPacket
    /**
    * Command information object does not exist. Set to null type.
    * @remark, never be NULL, an AMF0 null instance.
    */
    CommandObject Amf0Null
    /**
    * Name with which the stream is published.
    */
    StreamName Amf0String
    /**
    * Type of publishing. Set to “live”, “record”, or “append”.
    *   record: The stream is published and the data is recorded to a new file.The file
    *           is stored on the server in a subdirectory within the directory that
    *           contains the server application. If the file already exists, it is
    *           overwritten.
    *   append: The stream is published and the data is appended to a file. If no file
    *           is found, it is created.
    *   live: Live data is published without recording it in a file.
    * @remark, SRS only support live.
    * @remark, optional, default to live.
    */
    StreamType Amf0String
}

func NewRtmpPublishPacket() RtmpPacket {
    v := &RtmpPublishPacket{}
    v.CommandName = Amf0String(RTMP_AMF0_COMMAND_PUBLISH)
    v.TransactionId = Amf0Number(0.0)
    v.StreamType = Amf0String("live")
    return v
}

func (pkt *RtmpPublishPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Decode(buffer, logger); err != nil {
        return
    }
    if err = DecodeAmf0Null(buffer); err != nil {
        return
    }
    if pkt.StreamName,err = DecodeAmf0String(buffer); err != nil {
        return
    }
    if buffer.Len() > 0 {
        if pkt.StreamType,err = DecodeAmf0String(buffer); err != nil {
            return
        }
    }
    return
}

func (pkt *RtmpPublishPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Encode(buffer, logger); err != nil {
        return
    }
    if err = EncodeAmf0Null(buffer); err != nil {
        return
    }
    if err = EncodeAmf0String(buffer, pkt.StreamName); err != nil {
        return
    }
    if err = EncodeAmf0String(buffer, pkt.StreamType); err != nil {
        return
    }
    return
}

func (pkt *RtmpPublishPacket) MessageType() byte {
    return RTMP_MSG_AMF0CommandMessage
}

func (pkt *RtmpPublishPacket) PerferCid() int {
    return RTMP_CID_OverStream
}

/**
* onStatus command, AMF0 Call
* @remark, user must set the stream_id by SrsCommonMessage.set_packet().
*/
type RtmpOnStatusCallPacket struct {
    rtmpCommonCallPacket
    /**
    * Command information does not exist. Set to null type.
    * @remark, never be NULL, an AMF0 null instance.
    */
    CommandObject Amf0Null
    /**
    * Name-value pairs that describe the response from the server.
    * ‘code’,‘level’, ‘description’ are names of few among such information.
    * @remark, never be NULL, an AMF0 object instance.
    */
    Data *Amf0Object
}

func NewRtmpOnStatusCallPacket() RtmpPacket {
    v := &RtmpOnStatusCallPacket{}
    v.CommandName = Amf0String(RTMP_AMF0_COMMAND_ON_STATUS)
    v.TransactionId = Amf0Number(0.0)
    v.Data = NewAmf0Object()
    return v
}

func (pkt *RtmpOnStatusCallPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Decode(buffer, logger); err != nil {
        return
    }
    if err = DecodeAmf0Null(buffer); err != nil {
        return
    }
    if err = pkt.Data.Decode(buffer); err != nil {
        return
    }
    return
}

func (pkt *RtmpOnStatusCallPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = pkt.rtmpCommonCallPacket.Encode(buffer, logger); err != nil {
        return
    }
    if err = EncodeAmf0Null(buffer); err != nil {
        return
    }
    if err = pkt.Data.Encode(buffer); err != nil {
        return
    }
    return
}

func (pkt *RtmpOnStatusCallPacket) MessageType() byte {
    return RTMP_MSG_AMF0CommandMessage
}

func (pkt *RtmpOnStatusCallPacket) PerferCid() int {
    return RTMP_CID_OverStream
}

/**
* AMF0Data RtmpSampleAccess
* @remark, user must set the stream_id by SrsCommonMessage.set_packet().
*/
type RtmpSampleAccessPacket struct {
    /**
    * Name of command. Set to "|RtmpSampleAccess".
    */
    CommandName Amf0String
    /**
    * whether allow access the sample of video.
    * @see: https://github.com/winlinvip/simple-rtmp-server/issues/49
    * @see: http://help.adobe.com/en_US/FlashPlatform/reference/actionscript/3/flash/net/NetStream.html#videoSampleAccess
    */
    VideoSampleAccess Amf0Boolean
    /**
    * whether allow access the sample of audio.
    * @see: https://github.com/winlinvip/simple-rtmp-server/issues/49
    * @see: http://help.adobe.com/en_US/FlashPlatform/reference/actionscript/3/flash/net/NetStream.html#audioSampleAccess
    */
    AudioSampleAccess Amf0Boolean
}

func NewRtmpSampleAccessPacket() RtmpPacket {
    v := &RtmpSampleAccessPacket{}
    v.CommandName = Amf0String(RTMP_AMF0_DATA_SAMPLE_ACCESS)
    return v
}

func (pkt *RtmpSampleAccessPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if pkt.CommandName,err = DecodeAmf0String(buffer); err != nil {
        return
    }
    if pkt.VideoSampleAccess,err = DecodeAmf0Boolean(buffer); err != nil {
        return
    }
    if pkt.AudioSampleAccess,err = DecodeAmf0Boolean(buffer); err != nil {
        return
    }
    return
}

func (pkt *RtmpSampleAccessPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = EncodeAmf0String(buffer, pkt.CommandName); err != nil {
        return
    }
    if err = EncodeAmf0Boolean(buffer, pkt.VideoSampleAccess); err != nil {
        return
    }
    if err = EncodeAmf0Boolean(buffer, pkt.AudioSampleAccess); err != nil {
        return
    }
    return
}

func (pkt *RtmpSampleAccessPacket) MessageType() byte {
    return RTMP_MSG_AMF0DataMessage
}

func (pkt *RtmpSampleAccessPacket) PerferCid() int {
    return RTMP_CID_OverStream
}

/**
* onStatus data, AMF0 Data
* @remark, user must set the stream_id by SrsCommonMessage.set_packet().
*/
type RtmpOnStatusDataPacket struct {
    /**
    * Name of command. Set to "onStatus"
    */
    CommandName Amf0String
    /**
    * Name-value pairs that describe the response from the server.
    * ‘code’, are names of few among such information.
    * @remark, never be NULL, an AMF0 object instance.
    */
    Data *Amf0Object
}

func NewRtmpOnStatusDataPacket() RtmpPacket {
    v := &RtmpOnStatusDataPacket{}
    v.CommandName = Amf0String(RTMP_AMF0_COMMAND_ON_STATUS)
    v.Data = NewAmf0Object()
    return v
}

func (pkt *RtmpOnStatusDataPacket) Decode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if pkt.CommandName,err = DecodeAmf0String(buffer); err != nil {
        return
    }
    if err = pkt.Data.Decode(buffer); err != nil {
        return
    }
    return
}

func (pkt *RtmpOnStatusDataPacket) Encode(buffer *bytes.Buffer, logger core.Logger) (err error) {
    if err = EncodeAmf0String(buffer, pkt.CommandName); err != nil {
        return
    }
    if err = pkt.Data.Encode(buffer); err != nil {
        return
    }
    return
}

func (pkt *RtmpOnStatusDataPacket) MessageType() byte {
    return RTMP_MSG_AMF0DataMessage
}

func (pkt *RtmpOnStatusDataPacket) PerferCid() int {
    return RTMP_CID_OverStream
}
