package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessage(t *testing.T) {
	msg := NewMessage()
	status := Error
	serializeType := MsgPack
	var seq uint64 = 10000
	msg.SetMessageStatusType(status)
	msg.SetSeq(seq)
	msg.SetSerializeType(serializeType)

	assert.Equal(t, status, msg.MessageStatusType(), "message status should be equal")
	assert.Equal(t, seq, msg.Seq(), "seq should be equal")
	assert.Equal(t, serializeType, msg.SerializeType(), "serializeType should be equal")
}

func TestMessage_Encode(t *testing.T) {
	msg := NewMessage()
	msg.ServicePath = "123"
	msg.ServiceMethod = "456"
	msg.Payload = []byte("789")
	bytes1 := msg.Encode()
	bytes2 := msg.Encode2()
	assert.Equal(t, len(*bytes1), len(bytes2))
}

func BenchmarkMessage_Encode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		msg := NewMessage()
		msg.ServicePath = "123"
		msg.ServiceMethod = "456"
		msg.Payload = []byte("789")
		data := msg.Encode()
		PutData(data)
	}
}

func BenchmarkMessage_Encode2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		msg := NewMessage()
		msg.ServicePath = "123"
		msg.ServiceMethod = "456"
		msg.Payload = []byte("789")
		msg.Encode2()
	}
}
