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
	"net"
	"encoding/binary"
	"io"
	"fmt"
	"errors"
	"bytes"
	"math"
)

var RtmpChunkStart = errors.New("new chunk stream cid must be fresh")
var RtmpPacketSize = errors.New("chunk size should not changed")
var RtmpTcUrlNotString = errors.New("tcUrl of connect app must be string")
var RtmpRequestSchemaEmpty = errors.New("request schema is empty")
var RtmpRequestVhostEmpty = errors.New("request vhost is empty")
var RtmpRequestPortEmpty = errors.New("request port is empty")
var RtmpRequestAppEmpty = errors.New("request app is empty")

/**
* 6.1.2. Chunk Message Header
* There are four different formats for the chunk message header,
* selected by the "fmt" field in the chunk basic header.
*/
const (
	// 6.1.2.1. Type 0
	// Chunks of Type 0 are 11 bytes long. This type MUST be used at the
	// start of a chunk stream, and whenever the stream timestamp goes
	// backward (e.g., because of a backward seek).
	RTMP_FMT_TYPE0 = iota
	// 6.1.2.2. Type 1
	// Chunks of Type 1 are 7 bytes long. The message stream ID is not
	// included; this chunk takes the same stream ID as the preceding chunk.
	// Streams with variable-sized messages (for example, many video
	// formats) SHOULD use this format for the first chunk of each new
	// message after the first.
	RTMP_FMT_TYPE1
	// 6.1.2.3. Type 2
	// Chunks of Type 2 are 3 bytes long. Neither the stream ID nor the
	// message length is included; this chunk has the same stream ID and
	// message length as the preceding chunk. Streams with constant-sized
	// messages (for example, some audio and data formats) SHOULD use this
	// format for the first chunk of each message after the first.
	RTMP_FMT_TYPE2
	// 6.1.2.4. Type 3
	// Chunks of Type 3 have no header. Stream ID, message length and
	// timestamp delta are not present; chunks of this type take values from
	// the preceding chunk. When a single message is split into chunks, all
	// chunks of a message except the first one, SHOULD use this type. Refer
	// to example 2 in section 6.2.2. Stream consisting of messages of
	// exactly the same size, stream ID and spacing in time SHOULD use this
	// type for all chunks after chunk of Type 2. Refer to example 1 in
	// section 6.2.1. If the delta between the first message and the second
	// message is same as the time stamp of first message, then chunk of
	// type 3 would immediately follow the chunk of type 0 as there is no
	// need for a chunk of type 2 to register the delta. If Type 3 chunk
	// follows a Type 0 chunk, then timestamp delta for this Type 3 chunk is
	// the same as the timestamp of Type 0 chunk.
	RTMP_FMT_TYPE3
)

const (
	/**
    * the chunk stream id used for some under-layer message,
    * for example, the PC(protocol control) message.
    */
	RTMP_CID_ProtocolControl = 2 + iota
	/**
    * the AMF0/AMF3 command message, invoke method and return the result, over NetConnection.
    * generally use 0x03.
    */
	RTMP_CID_OverConnection
	/**
    * the AMF0/AMF3 command message, invoke method and return the result, over NetConnection,
    * the midst state(we guess).
    * rarely used, e.g. onStatus(NetStream.Play.Reset).
    */
	RTMP_CID_OverConnection2
	/**
    * the stream message(amf0/amf3), over NetStream.
    * generally use 0x05.
    */
	RTMP_CID_OverStream
	/**
    * the stream message(amf0/amf3), over NetStream, the midst state(we guess).
    * rarely used, e.g. play("mp4:mystram.f4v")
    */
	RTMP_CID_Video
	/**
    * the stream message(video), over NetStream
    * generally use 0x06.
    */
	RTMP_CID_Audio
	/**
    * the stream message(audio), over NetStream.
    * generally use 0x07.
    */
	RTMP_CID_OverStream2
)

const (
	/**
    * 6.1. Chunk Format
    * Extended timestamp: 0 or 4 bytes
    * This field MUST be sent when the normal timsestamp is set to
    * 0xffffff, it MUST NOT be sent if the normal timestamp is set to
    * anything else. So for values less than 0xffffff the normal
    * timestamp field SHOULD be used in which case the extended timestamp
    * MUST NOT be present. For values greater than or equal to 0xffffff
    * the normal timestamp field MUST NOT be used and MUST be set to
    * 0xffffff and the extended timestamp MUST be sent.
    */
	RTMP_EXTENDED_TIMESTAMP = 0xFFFFFF
	// 6. Chunking, RTMP protocol default chunk size.
	SRS_CONSTS_RTMP_PROTOCOL_CHUNK_SIZE = 128

	/**
    * 6. Chunking
    * The chunk size is configurable. It can be set using a control
    * message(Set Chunk Size) as described in section 7.1. The maximum
    * chunk size can be 65536 bytes and minimum 128 bytes. Larger values
    * reduce CPU usage, but also commit to larger writes that can delay
    * other content on lower bandwidth connections. Smaller chunks are not
    * good for high-bit rate streaming. Chunk size is maintained
    * independently for each direction.
    */
	SRS_CONSTS_RTMP_MIN_CHUNK_SIZE = 128
	SRS_CONSTS_RTMP_MAX_CHUNK_SIZE = 65536
)

/**
5. Protocol Control Messages
RTMP reserves message type IDs 1-7 for protocol control messages.
These messages contain information needed by the RTM Chunk Stream
protocol or RTMP itself. Protocol messages with IDs 1 & 2 are
reserved for usage with RTM Chunk Stream protocol. Protocol messages
with IDs 3-6 are reserved for usage of RTMP. Protocol message with ID
7 is used between edge server and origin server.
*/
const (
	RTMP_MSG_SetChunkSize = 1 + iota
	RTMP_MSG_AbortMessage
	RTMP_MSG_Acknowledgement
	RTMP_MSG_UserControlMessage
	RTMP_MSG_WindowAcknowledgementSize
	RTMP_MSG_SetPeerBandwidth
	RTMP_MSG_EdgeAndOriginServerCommand
)

const (
	/**
    3. Types of messages
    The server and the client send messages over the network to
    communicate with each other. The messages can be of any type which
    includes audio messages, video messages, command messages, shared
    object messages, data messages, and user control messages.
    3.1. Command message
    Command messages carry the AMF-encoded commands between the client
    and the server. These messages have been assigned message type value
    of 20 for AMF0 encoding and message type value of 17 for AMF3
    encoding. These messages are sent to perform some operations like
    connect, createStream, publish, play, pause on the peer. Command
    messages like onstatus, result etc. are used to inform the sender
    about the status of the requested commands. A command message
    consists of command name, transaction ID, and command object that
    contains related parameters. A client or a server can request Remote
    Procedure Calls (RPC) over streams that are communicated using the
    command messages to the peer.
    */
	RTMP_MSG_AMF3CommandMessage = 17 // 0x11
	RTMP_MSG_AMF0CommandMessage = 20 // 0x14
	/**
    3.2. Data message
    The client or the server sends this message to send Metadata or any
    user data to the peer. Metadata includes details about the
    data(audio, video etc.) like creation time, duration, theme and so
    on. These messages have been assigned message type value of 18 for
    AMF0 and message type value of 15 for AMF3.
    */
	RTMP_MSG_AMF0DataMessage = 18 // 0x12
	RTMP_MSG_AMF3DataMessage = 15 // 0x0F
	/**
    3.3. Shared object message
    A shared object is a Flash object (a collection of name value pairs)
    that are in synchronization across multiple clients, instances, and
    so on. The message types kMsgContainer=19 for AMF0 and
    kMsgContainerEx=16 for AMF3 are reserved for shared object events.
    Each message can contain multiple events.
    */
	RTMP_MSG_AMF3SharedObject = 16 // 0x10
	RTMP_MSG_AMF0SharedObject = 19 // 0x13
	/**
    3.4. Audio message
    The client or the server sends this message to send audio data to the
    peer. The message type value of 8 is reserved for audio messages.
    */
	RTMP_MSG_AudioMessage = 8 // 0x08
	/* *
    3.5. Video message
    The client or the server sends this message to send video data to the
    peer. The message type value of 9 is reserved for video messages.
    These messages are large and can delay the sending of other type of
    messages. To avoid such a situation, the video message is assigned
    the lowest priority.
    */
	RTMP_MSG_VideoMessage = 9 // 0x09
	/**
    3.6. Aggregate message
    An aggregate message is a single message that contains a list of submessages.
    The message type value of 22 is reserved for aggregate
    messages.
    */
	RTMP_MSG_AggregateMessage = 22 // 0x16
)

/**
* amf0 command message, command name macros
*/
const (
	RTMP_AMF0_COMMAND_CONNECT = "connect"
	RTMP_AMF0_COMMAND_CREATE_STREAM = "createStream"
	RTMP_AMF0_COMMAND_CLOSE_STREAM = "closeStream"
	RTMP_AMF0_COMMAND_PLAY = "play"
	RTMP_AMF0_COMMAND_PAUSE = "pause"
	RTMP_AMF0_COMMAND_ON_BW_DONE = "onBWDone"
	RTMP_AMF0_COMMAND_ON_STATUS = "onStatus"
	RTMP_AMF0_COMMAND_RESULT = "_result"
	RTMP_AMF0_COMMAND_ERROR = "_error"
	RTMP_AMF0_COMMAND_RELEASE_STREAM = "releaseStream"
	RTMP_AMF0_COMMAND_FC_PUBLISH = "FCPublish"
	RTMP_AMF0_COMMAND_UNPUBLISH = "FCUnpublish"
	RTMP_AMF0_COMMAND_PUBLISH = "publish"
	RTMP_AMF0_DATA_SAMPLE_ACCESS = "|RtmpSampleAccess"
	RTMP_AMF0_DATA_SET_DATAFRAME = "@setDataFrame"
	RTMP_AMF0_DATA_ON_METADATA = "onMetaData"
)

// the rtmp message header map to fmt.
var mhSizes = []byte{11, 7, 3, 0}

type AckWindowSize struct {
	Ack int
	Acked int
}

type Protocol struct {
	IoRw *net.TCPConn
	Logger core.Logger
	ChunkStreams map[int]*ChunkStream
	InChunkSize int
	OutChunkSize int
	/**
    * input ack size, when to send the acked packet.
    */
	InAckSize AckWindowSize
}

func NewProtocol(iorw *net.TCPConn, logger core.Logger) *Protocol {
	v := &Protocol{
		IoRw: iorw,
		Logger: logger,
		InChunkSize: SRS_CONSTS_RTMP_PROTOCOL_CHUNK_SIZE,
		OutChunkSize: SRS_CONSTS_RTMP_PROTOCOL_CHUNK_SIZE,
	}
	v.ChunkStreams = map[int]*ChunkStream{}
	return v
}

func (proto *Protocol) EncodeMessage(pkt RtmpPacket, streamId int) (msg *RtmpMessage, err error) {
	buffer := bytes.Buffer{}
	if err = pkt.Encode(&buffer, proto.Logger); err != nil {
		return
	}

	msg = NewRtmpMessage()
	msg.Payload = buffer.Bytes()
	msg.Header.PayloadLength = int32(len(msg.Payload))
	msg.Header.MessageType = int8(pkt.MessageType())
	msg.Header.StreamId = int32(streamId)
	msg.Header.PerferCid = pkt.PerferCid()
	return
}

func (proto *Protocol) DecodeMessage(msg *RtmpMessage) (pkt RtmpPacket, err error) {
	var b []byte
	if b,pkt,err = DiscoveryPacket(msg, proto.Logger); err != nil {
		return
	}

	if pkt == nil {
		proto.Logger.Info("null packet")
		return
	}

	proto.Logger.Info("disconvery rtmp packet ok")

	if err = pkt.Decode(bytes.NewBuffer(b), proto.Logger); err != nil {
		return
	}
	proto.Logger.Info("decode rtmp packet ok")

	return
}

func (proto *Protocol) SendMessage(msg *RtmpMessage) (err error) {
	hc0 := &bytes.Buffer{}
	var hc3 *bytes.Buffer = nil

	hasC3Chunk := (int(msg.Header.PayloadLength) > proto.OutChunkSize);
	if hasC3Chunk {
		hc3 = &bytes.Buffer{}
	}

	// write new chunk stream header, fmt is 0
	if err = hc0.WriteByte(byte(msg.Header.PerferCid & 0x3F)); err != nil {
		return
	}
	if hasC3Chunk {
		// write no message header chunk stream, fmt is 3
		// @remark, if perfer_cid > 0x3F, that is, use 2B/3B chunk header,
		// SRS will rollback to 1B chunk header.
		if err = hc3.WriteByte(byte(0xC0 | (msg.Header.PerferCid & 0x3F))); err != nil {
			return
		}
	}

	// chunk message header, 11 bytes
	// timestamp, 3bytes, big-endian
	timestamp := uint32(msg.Header.Timestamp) & 0x7fffffff
	if timestamp < RTMP_EXTENDED_TIMESTAMP {
		if _,err = hc0.Write([]byte{
			byte(timestamp >> 16),
			byte(timestamp >> 8),
			byte(timestamp),
		}); err != nil {
			return
		}
	} else {
		if _,err = hc0.Write([]byte{0xFF, 0xFF, 0xFF}); err != nil {
			return
		}
	}

	// message_length, 3bytes, big-endian
	if _,err = hc0.Write([]byte{
		byte(msg.Header.PayloadLength >> 16),
		byte(msg.Header.PayloadLength >> 8),
		byte(msg.Header.PayloadLength),
	}); err != nil {
		return
	}

	// message_type, 1bytes
	if err = hc0.WriteByte(byte(msg.Header.MessageType)); err != nil {
		return
	}

	// stream_id, 4bytes, little-endian
	if err = binary.Write(hc0, binary.LittleEndian, msg.Header.StreamId); err != nil {
		return
	}

	// for c0
	// chunk extended timestamp header, 0 or 4 bytes, big-endian
	//
	// for c3:
	// chunk extended timestamp header, 0 or 4 bytes, big-endian
	// 6.1.3. Extended Timestamp
	// This field is transmitted only when the normal time stamp in the
	// chunk message header is set to 0x00ffffff. If normal time stamp is
	// set to any value less than 0x00ffffff, this field MUST NOT be
	// present. This field MUST NOT be present if the timestamp field is not
	// present. Type 3 chunks MUST NOT have this field.
	// adobe changed for Type3 chunk:
	//        FMLE always sendout the extended-timestamp,
	//        must send the extended-timestamp to FMS,
	//        must send the extended-timestamp to flash-player.
	// @see: ngx_rtmp_prepare_message
	// @see: http://blog.csdn.net/win_lin/article/details/13363699
	// TODO: FIXME: extract to outer.
	if timestamp >= RTMP_EXTENDED_TIMESTAMP {
		if err = binary.Write(hc0, binary.BigEndian, timestamp); err != nil {
			return
		}
		if hasC3Chunk {
			if err = binary.Write(hc3, binary.BigEndian, timestamp); err != nil {
				return
			}
		}
	}

	var nbSent int32 = 0
	for nbSent < msg.Header.PayloadLength {
		if nbSent == 0 {
			if _,err = proto.IoRw.Write(hc0.Bytes()); err != nil {
				return
			}
			proto.Logger.Info("write %vB c0 header ok", hc0.Len())
		} else {
			if _,err = proto.IoRw.Write(hc3.Bytes()); err != nil {
				return
			}
			proto.Logger.Info("write %vB c3 header ok", hc3.Len())
		}

		payloadSize := msg.Header.PayloadLength - nbSent
		if payloadSize > int32(proto.OutChunkSize) {
			payloadSize = int32(proto.OutChunkSize)
		}
		if _,err = proto.IoRw.Write(msg.Payload[nbSent : nbSent + payloadSize]); err != nil {
			return
		}
		nbSent += payloadSize
		proto.Logger.Info("write %vB chunk payload ok, %v/%v", payloadSize, nbSent, msg.Header.PayloadLength)
	}

	// when sent message, filter it.
	if err = proto.onSendMessage(msg); err != nil {
		return
	}
	proto.Logger.Info("send msg ok")

	return
}

func (proto *Protocol) onSendMessage(msg *RtmpMessage) (err error) {
	var pkt RtmpPacket
	switch msg.Header.MessageType {
	case RTMP_MSG_AMF0CommandMessage:
		fallthrough
	case RTMP_MSG_AMF3CommandMessage:
		fallthrough
	case RTMP_MSG_SetChunkSize:
		if pkt,err = proto.DecodeMessage(msg); err != nil {
			proto.Logger.Error("decode packet from message payload failed.")
			return
		}
		proto.Logger.Info("decode packet from message payload ok.")
	default:
		return
	}

	switch pkt := pkt.(type) {
	case *RtmpSetChunkSizePacket:
		proto.OutChunkSize = int(pkt.ChunkSize)
		proto.Logger.Trace("out chunk size to %v", pkt.ChunkSize)
	// TODO: FIXME: implements it.
	// case *RtmConnectAppPacket
	// case *RtmCreateStreamPacket
	// case *RtmReleaseStreamPacket
	}

	return
}

func (proto *Protocol) PumpMessage() (msg *RtmpMessage, err error) {
	var fmt byte
	var cid int
	if fmt,cid,err = proto.readBasicHeader(); err != nil {
		return
	}
	proto.Logger.Info("read basic header fmt=%v, cid=%v", fmt, cid)

	var ok bool
	var chunk *ChunkStream
	if chunk,ok = proto.ChunkStreams[cid]; !ok {
		chunk = NewChunkStream(cid)
		proto.ChunkStreams[cid] = chunk
		proto.Logger.Info("create chunk stream cid=%v", cid)
	}

	var extsBuffer bytes.Buffer
	if err = proto.readMessageHeader(chunk, fmt, &extsBuffer); err != nil {
		return
	}
	proto.Logger.Info("read header ok. fmt=%v, message(type=%v, size=%v, time=%v, sid=%v), eb=%vB",
		fmt, chunk.Msg.Header.MessageType, chunk.Header.PayloadLength, chunk.Header.Timestamp,
		chunk.Header.StreamId, extsBuffer.Len())

	if msg,err = proto.readMessagePayload(chunk, &extsBuffer); err != nil {
		return
	}

	// not got an entire RTMP message, try next chunk.
	if msg == nil {
		return
	}

	// when got message filter it first.
	if err = proto.onRecvMessage(msg); err != nil {
		return
	}

	proto.Logger.Info("get entire message success. size=%v, message(type=%v, size=%v, time=%v, sid=%v)",
		len(msg.Payload), msg.Header.MessageType, msg.Header.PayloadLength, msg.Header.Timestamp, msg.Header.StreamId)

	return
}

func (proto *Protocol) onRecvMessage(msg *RtmpMessage) (err error) {
	// acknowledgement
	// TODO: FIXME: implements it.

	var pkt RtmpPacket
	switch msg.Header.MessageType {
	case RTMP_MSG_SetChunkSize:
		fallthrough
	case RTMP_MSG_UserControlMessage:
		fallthrough
	case RTMP_MSG_WindowAcknowledgementSize:
		if pkt,err = proto.DecodeMessage(msg); err != nil {
			proto.Logger.Error("decode packet from message payload failed.")
			return
		}
		proto.Logger.Info("decode packet from message payload success.")
	default:
		return
	}

	switch pkt := pkt.(type) {
	case *RtmpSetWindowAckSizePacket:
		if pkt.AckowledgementWindowSize > 0 {
			proto.InAckSize.Ack = int(pkt.AckowledgementWindowSize)
			// @remark, we ignore this message, for user noneed to care.
			// but it's important for dev, for client/server will block if required
			// ack msg not arrived.
			proto.Logger.Info("set ack window size to %v", pkt.AckowledgementWindowSize)
		} else {
			proto.Logger.Warn("ignored. set ack window size is %v", pkt.AckowledgementWindowSize)
		}
	case *RtmpSetChunkSizePacket:
		// for some server, the actual chunk size can greater than the max value(65536),
		// so we just warning the invalid chunk size, and actually use it is ok,
		// @see: https://github.com/winlinvip/simple-rtmp-server/issues/160
		if pkt.ChunkSize < SRS_CONSTS_RTMP_MIN_CHUNK_SIZE || pkt.ChunkSize > SRS_CONSTS_RTMP_MAX_CHUNK_SIZE {
			proto.Logger.Warn("accept chunk size %d, but should in [%v, %v], @see: https://github.com/winlinvip/simple-rtmp-server/issues/160",
				pkt.ChunkSize, SRS_CONSTS_RTMP_MIN_CHUNK_SIZE, SRS_CONSTS_RTMP_MAX_CHUNK_SIZE)
		}
		proto.InChunkSize = int(pkt.ChunkSize)
		proto.Logger.Trace("input chunk size to %v", pkt.ChunkSize)
	}

	return
}

/**
* 6.1.1. Chunk Basic Header
* The Chunk Basic Header encodes the chunk stream ID and the chunk
* type(represented by fmt field in the figure below). Chunk type
* determines the format of the encoded message header. Chunk Basic
* Header field may be 1, 2, or 3 bytes, depending on the chunk stream
* ID.
*
* The bits 0–5 (least significant) in the chunk basic header represent
* the chunk stream ID.
*
* Chunk stream IDs 2-63 can be encoded in the 1-byte version of this
* field.
*    0 1 2 3 4 5 6 7
*   +-+-+-+-+-+-+-+-+
*   |fmt|   cs id   |
*   +-+-+-+-+-+-+-+-+
*   Figure 6 Chunk basic header 1
*
* Chunk stream IDs 64-319 can be encoded in the 2-byte version of this
* field. ID is computed as (the second byte + 64).
*   0                   1
*   0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5
*   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*   |fmt|    0      | cs id - 64    |
*   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*   Figure 7 Chunk basic header 2
*
* Chunk stream IDs 64-65599 can be encoded in the 3-byte version of
* this field. ID is computed as ((the third byte)*256 + the second byte
* + 64).
*    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3
*   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*   |fmt|     1     |         cs id - 64            |
*   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*   Figure 8 Chunk basic header 3
*
* cs id: 6 bits
* fmt: 2 bits
* cs id - 64: 8 or 16 bits
*
* Chunk stream IDs with values 64-319 could be represented by both 2-
* byte version and 3-byte version of this field.
*/
func (proto *Protocol) readBasicHeader() (fmt byte, cid int, err error) {
	if err = binary.Read(proto.IoRw, binary.BigEndian, &fmt); err != nil {
		if err != io.EOF {
			proto.Logger.Error("read cid failed")
		}
		return
	}

	cid = int(fmt & 0x3f)
	fmt = (fmt >> 6) & 0x03
	proto.Logger.Info("read fmt=%v, cid=%v", fmt, cid)

	// 2-63, 1B chunk header
	if cid > 1 {
		proto.Logger.Info("1B basic header parsed. fmt=%d, cid=%d", fmt, cid);
		return
	}

	b0 := cid
	var b1 uint8

	// 64-319, 2B chunk header
	if err = binary.Read(proto.IoRw, binary.BigEndian, &b1); err != nil {
		if err != io.EOF {
			proto.Logger.Error("read 2B cid failed")
		}
		return
	}

	cid = 64
	cid += int(b1)
	if b0 == 0 {
		proto.Logger.Info("2B basic header parsed. fmt=%d, cid=%d, b0=%d, b1=%d", fmt, cid, b0, b1);
		return
	}

	// 64-65599, 3B chunk header
	var b2 uint8
	if err = binary.Read(proto.IoRw, binary.BigEndian, &b2); err != nil {
		if err != io.EOF {
			proto.Logger.Error("read 3B cid failed")
		}
		return
	}
	cid += int(b2) * 256
	proto.Logger.Info("3B basic header parsed. fmt=%d, cid=%d, b0=%d, b1=%d, b2=%d", fmt, cid, b0, b1, b2);

	return
}

/**
* parse the message header.
*   3bytes: timestamp delta,    fmt=0,1,2
*   3bytes: payload length,     fmt=0,1
*   1bytes: message type,       fmt=0,1
*   4bytes: stream id,          fmt=0
* where:
*   fmt=0, 0x0X
*   fmt=1, 0x4X
*   fmt=2, 0x8X
*   fmt=3, 0xCX
*/
func (proto *Protocol) readMessageHeader(chunk *ChunkStream, fmt byte, extsBuffer *bytes.Buffer) (err error) {
	/**
    * we should not assert anything about fmt, for the first packet.
    * (when first packet, the chunk->msg is NULL).
    * the fmt maybe 0/1/2/3, the FMLE will send a 0xC4 for some audio packet.
    * the previous packet is:
    *     04                // fmt=0, cid=4
    *     00 00 1a          // timestamp=26
    *     00 00 9d          // payload_length=157
    *     08                // message_type=8(audio)
    *     01 00 00 00       // stream_id=1
    * the current packet maybe:
    *     c4             // fmt=3, cid=4
    * it's ok, for the packet is audio, and timestamp delta is 26.
    * the current packet must be parsed as:
    *     fmt=0, cid=4
    *     timestamp=26+26=52
    *     payload_length=157
    *     message_type=8(audio)
    *     stream_id=1
    * so we must update the timestamp even fmt=3 for first packet.
    */
	// fresh packet used to update the timestamp even fmt=3 for first packet.
	// fresh packet always means the chunk is the first one of message.
	isFirstChunkOfMsg := bool(chunk.Msg == nil)

	// but, we can ensure that when a chunk stream is fresh,
	// the fmt must be 0, a new stream.
	if chunk.MsgCount == 0 && fmt != RTMP_FMT_TYPE0 {
		// for librtmp, if ping, it will send a fresh stream with fmt=1,
		// 0x42             where: fmt=1, cid=2, protocol contorl user-control message
		// 0x00 0x00 0x00   where: timestamp=0
		// 0x00 0x00 0x06   where: payload_length=6
		// 0x04             where: message_type=4(protocol control user-control message)
		// 0x00 0x06            where: event Ping(0x06)
		// 0x00 0x00 0x0d 0x0f  where: event data 4bytes ping timestamp.
		// @see: https://github.com/winlinvip/simple-rtmp-server/issues/98
		if chunk.Cid == RTMP_CID_ProtocolControl && fmt == RTMP_FMT_TYPE1 {
			proto.Logger.Warn("accept cid=2, fmt=1 to make librtmp happy.")
		} else {
			// must be a RTMP protocol level error.
			err = RtmpChunkStart
			proto.Logger.Error("chunk stream is fresh, fmt must be %d, actual is %d. cid=%d",
				RTMP_FMT_TYPE0, fmt, chunk.Cid)
			return
		}
	}

	// when exists cache msg, means got an partial message,
	// the fmt must not be type0 which means new message.
	if !isFirstChunkOfMsg && fmt == RTMP_FMT_TYPE0 {
		err = RtmpChunkStart
		proto.Logger.Error("chunk stream exists, fmt must not be %d, actual is %d.", RTMP_FMT_TYPE0, fmt)
		return
	}

	// create msg when new chunk stream start
	if isFirstChunkOfMsg {
		chunk.Msg = NewRtmpMessage()
		proto.Logger.Info("create message for new chunk, fmt=%d, cid=%d", fmt, chunk.Cid)
	}

	// read message header from socket to buffer.
	msgHeader := make([]byte, mhSizes[int(fmt)])
	proto.Logger.Info("calc chunk message header size. fmt=%d, mh_size=%d", fmt, len(msgHeader))

	if _,err = io.ReadFull(proto.IoRw, msgHeader); err != nil {
		if err != io.EOF {
			proto.Logger.Error("read %dB hreader failed", len(msgHeader))
		}
		return
	}
	proto.Logger.Info("read chunk header ok")

	extendedTimestamp := false
	/**
    * parse the message header.
    *   3bytes: timestamp delta,    fmt=0,1,2
    *   3bytes: payload length,     fmt=0,1
    *   1bytes: message type,       fmt=0,1
    *   4bytes: stream id,          fmt=0
    * where:
    *   fmt=0, 0x0X
    *   fmt=1, 0x4X
    *   fmt=2, 0x8X
    *   fmt=3, 0xCX
    */
	// see also: SrsProtocol::read_message_header
	if fmt <= RTMP_FMT_TYPE2 {
		b := msgHeader

		// fmt: 0
		// timestamp: 3 bytes
		// If the timestamp is greater than or equal to 16777215
		// (hexadecimal 0x00ffffff), this value MUST be 16777215, and the
		// ‘extended timestamp header’ MUST be present. Otherwise, this value
		// SHOULD be the entire timestamp.
		//
		// fmt: 1 or 2
		// timestamp delta: 3 bytes
		// If the delta is greater than or equal to 16777215 (hexadecimal
		// 0x00ffffff), this value MUST be 16777215, and the ‘extended
		// timestamp header’ MUST be present. Otherwise, this value SHOULD be
		// the entire delta.
		chunk.Header.TimestampDelta = int32(b[2]) | int32(b[1])<<8 | int32(b[0])<<16
		extendedTimestamp = (chunk.Header.TimestampDelta >= RTMP_EXTENDED_TIMESTAMP)
		if !extendedTimestamp {
			// Extended timestamp: 0 or 4 bytes
			// This field MUST be sent when the normal timsestamp is set to
			// 0xffffff, it MUST NOT be sent if the normal timestamp is set to
			// anything else. So for values less than 0xffffff the normal
			// timestamp field SHOULD be used in which case the extended timestamp
			// MUST NOT be present. For values greater than or equal to 0xffffff
			// the normal timestamp field MUST NOT be used and MUST be set to
			// 0xffffff and the extended timestamp MUST be sent.
			if fmt == RTMP_FMT_TYPE0 {
				// 6.1.2.1. Type 0
				// For a type-0 chunk, the absolute timestamp of the message is sent
				// here.
				chunk.Header.Timestamp = int64(chunk.Header.TimestampDelta)
			} else {
				// 6.1.2.2. Type 1
				// 6.1.2.3. Type 2
				// For a type-1 or type-2 chunk, the difference between the previous
				// chunk's timestamp and the current chunk's timestamp is sent here.
				chunk.Header.Timestamp += int64(chunk.Header.TimestampDelta)
			}
		}

		if fmt <= RTMP_FMT_TYPE1 {
			var payloadLength int32 = int32(b[5]) | int32(b[4])<<8 | int32(b[3])<<16

			// for a message, if msg exists in cache, the size must not changed.
			// always use the actual msg size to compare, for the cache payload length can changed,
			// for the fmt type1(stream_id not changed), user can change the payload
			// length(it's not allowed in the continue chunks).
			if !isFirstChunkOfMsg && chunk.Header.PayloadLength != payloadLength {
				err = RtmpPacketSize
				proto.Logger.Error("msg exists in chunk cache, size=%d cannot change to %d",
					chunk.Header.PayloadLength, payloadLength)
				return
			}
			chunk.Header.PayloadLength = payloadLength
			chunk.Header.MessageType = int8(b[6])

			if fmt == RTMP_FMT_TYPE0 {
				chunk.Header.StreamId = int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24
				proto.Logger.Info("header read completed. fmt=%v, mh_size=%v, ext_time=%v, time=%v, payload=%v, type=%v, sid=%v",
					fmt, len(b), extendedTimestamp, chunk.Header.Timestamp, chunk.Header.PayloadLength,
					chunk.Header.MessageType, chunk.Header.StreamId)
			} else {
				proto.Logger.Info("header read completed. fmt=%v, mh_size=%v, ext_time=%v, time=%v, payload=%v, type=%v",
					fmt, len(b), extendedTimestamp, chunk.Header.Timestamp, chunk.Header.PayloadLength,
					chunk.Header.MessageType)
			}
		} else {
			proto.Logger.Info("header read completed. fmt=%v, mh_size=%v, ext_time=%v, time=%v",
				fmt, len(b), extendedTimestamp, chunk.Header.Timestamp)
		}
	} else {
		// update the timestamp even fmt=3 for first chunk packet
		if isFirstChunkOfMsg && !extendedTimestamp {
			chunk.Header.Timestamp += int64(chunk.Header.TimestampDelta)
		}
		proto.Logger.Info("header read completed. fmt=%v, mh_size=%v, ext_time=%v, time=%v",
			fmt, len(msgHeader), extendedTimestamp, chunk.Header.Timestamp)
	}

	// read extended-timestamp
	if extendedTimestamp {
		b := make([]byte, 4)
		if _,err = io.ReadFull(proto.IoRw, b); err != nil {
			if err != io.EOF {
				proto.Logger.Error("read extended timestamp failed.")
			}
			return
		}

		var timestamp int32 = int32(b[3]) | int32(b[2])<<8 | int32(b[1])<<16 | int32(b[0])<<24
		// always use 31bits timestamp, for some server may use 32bits extended timestamp.
		// @see https://github.com/winlinvip/simple-rtmp-server/issues/111
		timestamp &= 0x7fffffff

		/**
        * RTMP specification and ffmpeg/librtmp is false,
        * but, adobe changed the specification, so flash/FMLE/FMS always true.
        * default to true to support flash/FMLE/FMS.
        *
        * ffmpeg/librtmp may donot send this filed, need to detect the value.
        * @see also: http://blog.csdn.net/win_lin/article/details/13363699
        * compare to the chunk timestamp, which is set by chunk message header
        * type 0,1 or 2.
        *
        * @remark, nginx send the extended-timestamp in sequence-header,
        * and timestamp delta in continue C1 chunks, and so compatible with ffmpeg,
        * that is, there is no continue chunks and extended-timestamp in nginx-rtmp.
        *
        * @remark, srs always send the extended-timestamp, to keep simple,
        * and compatible with adobe products.
        */
		var chunkTimestamp int32 = int32(chunk.Header.Timestamp)

		/**
        * if chunk_timestamp<=0, the chunk previous packet has no extended-timestamp,
        * always use the extended timestamp.
        */
		/**
        * about the is_first_chunk_of_msg.
        * @remark, for the first chunk of message, always use the extended timestamp.
        */
		if !isFirstChunkOfMsg && chunkTimestamp > 0 && chunkTimestamp != timestamp {
			if _,err = extsBuffer.Write(b); err != nil {
				return err
			}
			proto.Logger.Info("no 4bytes extended timestamp in the continued chunk");
		} else {
			chunk.Header.Timestamp = int64(timestamp)
		}
		proto.Logger.Info("header read ext_time completed. time=%v", chunk.Header.Timestamp)
	}

	// the extended-timestamp must be unsigned-int,
	//         24bits timestamp: 0xffffff = 16777215ms = 16777.215s = 4.66h
	//         32bits timestamp: 0xffffffff = 4294967295ms = 4294967.295s = 1193.046h = 49.71d
	// because the rtmp protocol says the 32bits timestamp is about "50 days":
	//         3. Byte Order, Alignment, and Time Format
	//                Because timestamps are generally only 32 bits long, they will roll
	//                over after fewer than 50 days.
	//
	// but, its sample says the timestamp is 31bits:
	//         An application could assume, for example, that all
	//        adjacent timestamps are within 2^31 milliseconds of each other, so
	//        10000 comes after 4000000000, while 3000000000 comes before
	//        4000000000.
	// and flv specification says timestamp is 31bits:
	//        Extension of the Timestamp field to form a SI32 value. This
	//        field represents the upper 8 bits, while the previous
	//        Timestamp field represents the lower 24 bits of the time in
	//        milliseconds.
	// in a word, 31bits timestamp is ok.
	// convert extended timestamp to 31bits.
	chunk.Header.Timestamp &= 0x7fffffff

	// valid message, the payload_length is 24bits,
	// so it should never be negative.
	core.AssertGreaterOrEquals(int64(chunk.Header.PayloadLength), 0)

	// copy header to msg
	chunk.Msg.Header = chunk.Header

	// increase the msg count, the chunk stream can accept fmt=1/2/3 message now.
	chunk.MsgCount++

	return
}

func (proto *Protocol) readMessagePayload(chunk *ChunkStream, extsBuffer *bytes.Buffer) (msg *RtmpMessage, err error) {
	// empty message
	if chunk.Header.PayloadLength <= 0 {
		proto.Logger.Trace("get an empty RTMP message(type=%v, size=%v, time=%v, sid=%v)",
			chunk.Header.MessageType, chunk.Header.PayloadLength, chunk.Header.Timestamp, chunk.Header.StreamId)
		msg = chunk.Msg
		chunk.Msg = nil
		return
	}

	// the chunk payload size.
	payloadSize := int(chunk.Header.PayloadLength) - len(chunk.Msg.Payload) - extsBuffer.Len()
	payloadSize = int(math.Min(float64(payloadSize), float64(proto.InChunkSize)))
	proto.Logger.Info("chunk payload size is %v, ext=%v, message_size=%v, received_size=%v, in_chunk_size=%v",
		payloadSize, extsBuffer.Len(), chunk.Header.PayloadLength, len(chunk.Msg.Payload), proto.InChunkSize)

	// create msg payload if not initialized
	if len(chunk.Msg.Payload) == 0 {
		chunk.Msg.Payload = make([]byte, 0, chunk.Header.PayloadLength)
		proto.Logger.Info("create payload for RTMP message. size=%v", chunk.Header.PayloadLength)
	}

	// the ext buffer read for extended timestamp, it's actually the payload.
	if extsBuffer.Len() > 0 {
		chunk.Msg.Payload = append(chunk.Msg.Payload, extsBuffer.Bytes()...)
		proto.Logger.Info("copy ext buffer to payload, size=%v", extsBuffer.Len())
	}

	// read payload to buffer
	b := make([]byte, payloadSize)
	if _,err = io.ReadFull(proto.IoRw, b); err != nil {
		if err != io.EOF {
			proto.Logger.Error("read payload failed, size=%v, read=%v", chunk.Msg.Header.PayloadLength, payloadSize)
		}
		return
	}
	chunk.Msg.Payload = append(chunk.Msg.Payload, b...)
	proto.Logger.Info("chunk payload read completed. payload_size=%v", payloadSize)

	// got entire RTMP message?
	if len(chunk.Msg.Payload) == int(chunk.Msg.Header.PayloadLength) {
		msg = chunk.Msg
		chunk.Msg = nil
		proto.Logger.Info("get entire RTMP message(type=%v, size=%v, time=%v, sid=%v)",
			chunk.Header.MessageType, chunk.Header.PayloadLength, chunk.Header.Timestamp, chunk.Header.StreamId)
		return
	}
	proto.Logger.Info("get partial RTMP message(type=%v, size=%v, time=%v, sid=%v), partial size=%v",
		chunk.Header.MessageType, chunk.Header.PayloadLength, chunk.Header.Timestamp, chunk.Header.StreamId,
		len(chunk.Msg.Payload))

	return
}

// the chunk stream
type ChunkStream struct {
	Cid int
	Header RtmpMessageHeader
	Msg *RtmpMessage
	MsgCount int
}

func NewChunkStream(cid int) *ChunkStream {
	v := &ChunkStream{}
	v.Cid = cid
	v.Header.PerferCid = cid
	return v
}

func (cs *ChunkStream) String() string {
	return fmt.Sprintf("%v", cs.Cid)
}

// the message header
type RtmpMessageHeader struct {
	/**
    * 3bytes.
    * Three-byte field that contains a timestamp delta of the message.
    * @remark, only used for decoding message from chunk stream.
    */
	TimestampDelta int32
	/**
    * 3bytes.
    * Three-byte field that represents the size of the payload in bytes.
    * It is set in big-endian format.
    */
	PayloadLength int32
	/**
    * 1byte.
    * One byte field to represent the message type. A range of type IDs
    * (1-7) are reserved for protocol control messages.
    */
	MessageType int8
	/**
    * 4bytes.
    * Four-byte field that identifies the stream of the message. These
    * bytes are set in little-endian format.
    */
	StreamId int32

	/**
    * Four-byte field that contains a timestamp of the message.
    * The 4 bytes are packed in the big-endian order.
    * @remark, used as calc timestamp when decode and encode time.
    * @remark, we use 64bits for large time for jitter detect and hls.
    */
	Timestamp int64
	/**
    * get the perfered cid(chunk stream id) which sendout over.
    * set at decoding, and canbe used for directly send message,
    * for example, dispatch to all connections.
    */
	PerferCid int
}

func (mh *RtmpMessageHeader) IsAudio() bool {
	return mh.MessageType == RTMP_MSG_AudioMessage
}

func (mh *RtmpMessageHeader) IsVideo() bool {
	return mh.MessageType == RTMP_MSG_VideoMessage
}

func (mh *RtmpMessageHeader) IsAmf0Command() bool {
	return mh.MessageType == RTMP_MSG_AMF0CommandMessage
}

func (mh *RtmpMessageHeader) IsAmf0Data() bool {
	return mh.MessageType == RTMP_MSG_AMF0DataMessage
}

func (mh *RtmpMessageHeader) IsAmf3Command() bool {
	return mh.MessageType == RTMP_MSG_AMF3CommandMessage
}

func (mh *RtmpMessageHeader) IsAmf3Data() bool {
	return mh.MessageType == RTMP_MSG_AMF3DataMessage
}

func (mh *RtmpMessageHeader) IsWindowAckledgementSize() bool {
	return mh.MessageType == RTMP_MSG_WindowAcknowledgementSize
}

func (mh *RtmpMessageHeader) IsAckledgement() bool {
	return mh.MessageType == RTMP_MSG_Acknowledgement
}

func (mh *RtmpMessageHeader) IsSetChunkSize() bool {
	return mh.MessageType == RTMP_MSG_SetChunkSize
}

func (mh *RtmpMessageHeader) IsUserControlMessage() bool {
	return mh.MessageType == RTMP_MSG_UserControlMessage
}

func (mh *RtmpMessageHeader) IsSetPeerBandwidth() bool {
	return mh.MessageType == RTMP_MSG_SetPeerBandwidth
}

func (mh *RtmpMessageHeader) IsAggregate() bool {
	return mh.MessageType == RTMP_MSG_AggregateMessage
}

// the rtmp protocol level message packet
type RtmpMessage struct {
	Header RtmpMessageHeader
	Payload []byte
}

func NewRtmpMessage() *RtmpMessage {
	return &RtmpMessage{
		Payload: []byte{},
	}
}

func (msg *RtmpMessage) String() string {
	return fmt.Sprintf("Message(%v,%v,%v)",
		msg.Header.MessageType, msg.Header.Timestamp, msg.Header.PayloadLength)
}

/**
* the original request from client.
*/
type RtmpRequest struct {
	/**
    * tcUrl: rtmp://request_vhost:port/app/stream
    * support pass vhost in query string, such as:
    *    rtmp://ip:port/app?vhost=request_vhost/stream
    *    rtmp://ip:port/app...vhost...request_vhost/stream
    */
	TcUrl string
	PageUrl string
	SwfUrl string
	ObjectEncoding int

	// data discovery from request.
	// discovery from tcUrl and play/publish.
	Schema string
	// the vhost in tcUrl.
	Vhost string
	// the host in tcUrl.
	Host string
	// the port in tcUrl.
	Port int
	// the app in tcUrl, without param.
	App string
	// the param in tcUrl(app).
	Param string
	// the stream in play/publish
	Stream string
	// for play live stream,
	// used to specified the stop when exceed the duration.
	// @see https://github.com/winlinvip/simple-rtmp-server/issues/45
	// in ms.
	Duration float64
	// the token in the connect request,
	// used for edge traverse to origin authentication,
	// @see https://github.com/winlinvip/simple-rtmp-server/issues/104
	Args *Amf0Object
}

func (req *RtmpRequest) StreamUrl() string {
	return fmt.Sprintf("%s/%s/%s", req.Vhost, req.App, req.Stream)
}

func (req *RtmpRequest) FormatArgs() string {
	if req.Args != nil {
		return "(obj)"
	} else {
		return "null"
	}
}

func (req *RtmpRequest) UpdateAuth(r *RtmpRequest) {
	// TODO: FIXME: implements it.
}

func (req *RtmpRequest) Parse(cc, args *Amf0Object, logger core.Logger) (err error) {
	if v,ok := cc.GetString("tcUrl"); ok {
		req.TcUrl = string(v)
	} else {
		logger.Error("invalid request, must specifies the tcUrl.")
		return RtmpTcUrlNotString
	}

	if v,ok := cc.GetString("pageUrl"); ok {
		req.PageUrl = string(v)
	}
	if v,ok := cc.GetString("swfUrl"); ok {
		req.SwfUrl = string(v)
	}
	if v,ok := cc.GetNumber("objectEncoding"); ok {
		req.ObjectEncoding = int(float64(v))
	}

	if args != nil {
		req.Args = args
		logger.Info("copy edge traverse to origin auth args.")
	}
	logger.Info("get connect app message params success.")

	req.Schema,req.Host,req.Vhost,req.App,req.Port,req.Param,err = DiscoveryTcUrl(req.TcUrl, logger)
	logger.Info("tcUrl=%v parsed", req.TcUrl)

	return
}

func (req *RtmpRequest) Validate(logger core.Logger) (err error) {
	if req.Schema == "" {
		logger.Error("request schema is empty")
		return RtmpRequestSchemaEmpty
	}
	if req.Vhost == "" {
		logger.Error("request vhost is empty")
		return RtmpRequestVhostEmpty
	}
	if req.Port <= 0 {
		logger.Error("request port is not positive")
		return RtmpRequestPortEmpty
	}
	if req.App == "" {
		logger.Error("request app is empty")
		return RtmpRequestAppEmpty
	}

	logger.Info("request validate ok")

	return
}
