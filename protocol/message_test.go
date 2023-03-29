package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessage(t *testing.T) {
	msg := NewMessage()
	status := Normal
	var seq uint64 = 10000
	msg.SetMessageStatusType(status)
	msg.SetSeq(seq)

	assert.Equal(t, status, msg.MessageStatusType(), "message status should be equal")
	assert.Equal(t, seq, msg.Seq(), "seq should be equal")
}
