package protocol

import (
	"bytes"
	"encoding/binary"
	"io"
	"unsafe"

	"github.com/dayueba/minirpc/pool"
)

var bufferPool = pool.NewLimitedPool(512, 4096)

const headerLen = 9

// MessageStatusType is status of messages.
type MessageStatusType byte

const (
	// Normal is normal requests and responses.
	Normal MessageStatusType = iota
	// Error indicates some errors occur.
	Error
)

// SerializeType defines serialization type of payload.
type SerializeType byte

const (
	// SerializeNone uses raw []byte and don't serialize/deserialize
	SerializeNone SerializeType = iota
	// JSON for payload.
	JSON
	// ProtoBuffer for payload.
	ProtoBuffer
	// MsgPack for payload
	MsgPack
)

type Message struct {
	*Header
	ServicePath   string
	ServiceMethod string
	Payload       []byte
	data          []byte
}

func NewMessage() *Message {
	header := Header([headerLen]byte{})

	return &Message{
		Header: &header,
	}
}

type Header [headerLen]byte

// MessageStatusType returns the message status type.
func (h *Header) MessageStatusType() MessageStatusType {
	return MessageStatusType(h[0] & 0x01)
}

// SetMessageStatusType sets message status type.
func (h *Header) SetMessageStatusType(mt MessageStatusType) {
	h[0] = (h[0] &^ 0x01) | (byte(mt) & 0x01)
}

// SerializeType returns serialization type of payload.
func (h *Header) SerializeType() SerializeType {
	return SerializeType((h[0] & 0x06) >> 1)
}

// SetSerializeType sets the serialization type.
func (h *Header) SetSerializeType(st SerializeType) {
	h[0] = (h[0] &^ 0x06) | (byte(st) << 1)
}

// Seq returns sequence number of messages.
func (h *Header) Seq() uint64 {
	return binary.BigEndian.Uint64(h[1:])
}

// SetSeq set sequence number.
func (h *Header) SetSeq(seq uint64) {
	binary.BigEndian.PutUint64(h[1:], seq)
}

func (m *Message) Encode() *[]byte {
	//bytebuffer := bytebufferpool.Get()

	spL := len(m.ServicePath)
	smL := len(m.ServiceMethod)
	dataLen := (4 + spL) + (4 + smL) + (4 + len(m.Payload))
	// header + dataLen + spLen + sp + smLen + sm + payloadLen + payload
	l := headerLen + 4 + dataLen

	buffer := bufferPool.Get(l)

	// header
	copy(*buffer, m.Header[:])
	// totalLen
	binary.BigEndian.PutUint32((*buffer)[9:13], uint32(dataLen))
	// path
	binary.BigEndian.PutUint32((*buffer)[13:17], uint32(spL))
	copy((*buffer)[17:17+spL], StringToSliceByte(m.ServicePath))
	// method
	binary.BigEndian.PutUint32((*buffer)[17+spL:21+spL], uint32(smL))
	copy((*buffer)[21+spL:21+spL+smL], StringToSliceByte(m.ServiceMethod))
	// payload
	binary.BigEndian.PutUint32((*buffer)[21+spL+smL:25+spL+smL], uint32(len(m.Payload)))
	copy((*buffer)[25+spL+smL:], m.Payload)

	return buffer
}

func (m *Message) Encode2() []byte {
	spL := len(m.ServicePath)
	smL := len(m.ServiceMethod)
	dataLen := (4 + spL) + (4 + smL) + (4 + len(m.Payload))
	// header + dataLen + spLen + sp + smLen + sm + payloadLen + payload
	l := headerLen + 4 + dataLen

	buffer := bytes.NewBuffer(make([]byte, 0, l))
	// header
	buffer.Write(m.Header[:])

	// totalLen
	binary.Write(buffer, binary.BigEndian, uint32(dataLen))

	// path
	binary.Write(buffer, binary.BigEndian, uint32(spL))
	binary.Write(buffer, binary.BigEndian, StringToSliceByte(m.ServicePath))
	// method
	binary.Write(buffer, binary.BigEndian, uint32(smL))
	binary.Write(buffer, binary.BigEndian, StringToSliceByte(m.ServiceMethod))

	// payload
	binary.Write(buffer, binary.BigEndian, uint32(len(m.Payload)))
	binary.Write(buffer, binary.BigEndian, m.Payload)

	return buffer.Bytes()
}

func PutData(data *[]byte) {
	bufferPool.Put(data)
}

func StringToSliceByte(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

// Decode 非线程安全
func (m *Message) Decode(r io.Reader) error {
	// parse header
	_, err := io.ReadFull(r, m.Header[:])
	if err != nil {
		return err
	}

	// total
	lenData := make([]byte, 4)
	_, err = io.ReadFull(r, lenData)
	if err != nil {
		return err
	}
	l := binary.BigEndian.Uint32(lenData)
	totalL := int(l)
	if cap(m.data) >= totalL { // reuse data
		m.data = m.data[:totalL]
	} else {
		m.data = make([]byte, totalL)
	}
	data := m.data
	_, err = io.ReadFull(r, data)
	if err != nil {
		return err
	}

	n := 0
	// parse servicePath
	l = binary.BigEndian.Uint32(data[n:4])
	n = n + 4
	nEnd := n + int(l)
	m.ServicePath = SliceByteToString(data[n:nEnd])
	n = nEnd

	// parse serviceMethod
	l = binary.BigEndian.Uint32(data[n : n+4])
	n = n + 4
	nEnd = n + int(l)
	m.ServiceMethod = SliceByteToString(data[n:nEnd])
	n = nEnd

	// parse payload
	l = binary.BigEndian.Uint32(data[n : n+4])
	_ = l
	n = n + 4
	m.Payload = data[n:]

	return err
}

func SliceByteToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func (m *Message) Clone() *Message {
	header := *m.Header
	c := NewMessage()
	c.Header = &header
	c.ServicePath = m.ServicePath
	c.ServiceMethod = m.ServiceMethod
	return c
}

// Reset clean data of this message but keep allocated data
func (m *Message) Reset() {
	resetHeader(m.Header)
	m.Payload = []byte{}
	m.data = m.data[:0]
	m.ServicePath = ""
	m.ServiceMethod = ""
}

var (
	zeroHeaderArray Header
	zeroHeader      = zeroHeaderArray[1:]
)

func resetHeader(h *Header) {
	copy(h[:], zeroHeader)
}
