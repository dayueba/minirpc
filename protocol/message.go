package protocol

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"unsafe"
)

const headerLen = 9

// MessageStatusType is status of messages.
type MessageStatusType byte

const (
	// Normal is normal requests and responses.
	Normal MessageStatusType = iota
	// Error indicates some errors occur.
	Error
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

// Header 9个字节，第一个字节存储元数据，目前只有一个bit存储该请求成功还是失败
// 后面8字节存储seq
type Header [headerLen]byte

// MessageStatusType returns the message status type.
func (h *Header) MessageStatusType() MessageStatusType {
	return MessageStatusType(h[0] & 0x03)
}

// SetMessageStatusType sets message status type.
func (h *Header) SetMessageStatusType(mt MessageStatusType) {
	h[2] = (h[0] &^ 0x03) | (byte(mt) & 0x03)
}

// Seq returns sequence number of messages.
func (h *Header) Seq() uint64 {
	return binary.BigEndian.Uint64(h[1:])
}

// SetSeq set sequence number.
func (h *Header) SetSeq(seq uint64) {
	binary.BigEndian.PutUint64(h[1:], seq)
}

func (m *Message) Encode() ([]byte, error) {
	spL := len(m.ServicePath)
	smL := len(m.ServiceMethod)
	dataLen := (4 + spL) + (4 + smL) + (4 + len(m.Payload))
	// header + dataLen + spLen + sp + smLen + sm + payloadLen + payload
	l := headerLen + 4 + dataLen
	buffer := bytes.NewBuffer(make([]byte, 0, l))
	// header
	buffer.Write(m.Header[:])
	// totalLen
	if err := binary.Write(buffer, binary.BigEndian, uint32(dataLen)); err != nil {
		return nil, err
	}
	// path
	if err := binary.Write(buffer, binary.BigEndian, uint32(spL)); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, StringToSliceByte(m.ServicePath)); err != nil {
		return nil, err
	}
	// method
	if err := binary.Write(buffer, binary.BigEndian, uint32(smL)); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, StringToSliceByte(m.ServiceMethod)); err != nil {
		return nil, err
	}
	// payload
	if err := binary.Write(buffer, binary.BigEndian, uint32(len(m.Payload))); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, m.Payload); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func StringToSliceByte(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func (m *Message) Decode(conn net.Conn) error {
	// parse header
	_, err := io.ReadFull(conn, m.Header[:])
	if err != nil {
		return err
	}

	// total
	lenData := make([]byte, 4)
	_, err = io.ReadFull(conn, lenData)
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
	_, err = io.ReadFull(conn, data)
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
