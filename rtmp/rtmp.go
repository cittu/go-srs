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
	"net"
	"encoding/binary"
	"io"
	"fmt"
	"errors"
	"bytes"
	"math"
)

const (
	RTMP_FMT_TYPE0 = iota
	RTMP_FMT_TYPE1
	RTMP_FMT_TYPE2
	RTMP_FMT_TYPE3
)
const (
	RTMP_CID_ProtocolControl = 2 + iota
	RTMP_CID_OverConnection
	RTMP_CID_OverConnection2
	RTMP_CID_OverStream
	RTMP_CID_Video
	RTMP_CID_Audio
	RTMP_CID_OverStream2
)
const (
	RTMP_EXTENDED_TIMESTAMP = 0xFFFFFF
	SRS_CONSTS_RTMP_PROTOCOL_CHUNK_SIZE = 128
)

// the rtmp message header map to fmt.
var mhSizes = []byte{11, 7, 3, 0}

var RtmpChunkStart = errors.New("new chunk stream cid must be fresh")
var RtmpPacketSize = errors.New("chunk size should not changed")

type Protocol struct {
	IoRw *net.TCPConn
	Logger core.Logger
	ChunkStreams map[int]*ChunkStream
	InChunkSize int
}

func NewProtocol(iorw *net.TCPConn, logger core.Logger) *Protocol {
	v := &Protocol{
		IoRw: iorw,
		Logger: logger,
		InChunkSize: SRS_CONSTS_RTMP_PROTOCOL_CHUNK_SIZE,
	}
	v.ChunkStreams = map[int]*ChunkStream{}
	return v
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

	proto.Logger.Info("get entire message success. size=%v, message(type=%v, size=%v, time=%v, sid=%v)",
		len(msg.Payload), msg.Header.MessageType, msg.Header.PayloadLength, msg.Header.Timestamp, msg.Header.StreamId)

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

type ChunkStream struct {
	Cid int
	Header RtmpMessageHeader
	Msg *RtmpMessage
	MsgCount int
}

func NewChunkStream(cid int) *ChunkStream {
	return &ChunkStream{
		Cid: cid,
	}
}

func (cs *ChunkStream) String() string {
	return fmt.Sprintf("%v", cs.Cid)
}

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

