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
