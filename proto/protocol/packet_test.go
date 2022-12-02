package protocol

import (
	"bytes"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPacketConcatenation(t *testing.T) {
	pkt1 := &Packet{
		Control: nil,
		SeqId:   0,
		Flag:    1234,
		Command: 453814,
		Headers: []*KVEntry{
			{Key: "abc", Value: "123"},
			{Key: "bcd", Value: "234"},
		},
		Payload: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
	}
	buf1, err1 := proto.Marshal(pkt1)
	assert.Nil(t, err1)

	pkt1.Control = []byte{100, 101, 102, 103, 104}
	pkt1.SeqId = 9876543219
	buf2, err2 := proto.Marshal(pkt1)
	assert.Nil(t, err2)
	assert.True(t, bytes.HasSuffix(buf2, buf1))

	pkt3 := &Packet{
		Control: []byte{100, 101, 102, 103, 104},
		SeqId:   9876543219,
	}
	buf3, err3 := proto.Marshal(pkt3)
	assert.Nil(t, err3)
	assert.True(t, bytes.HasPrefix(buf2, buf3))
}
