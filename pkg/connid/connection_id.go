package connid

import (
	"encoding/base32"
	"encoding/binary"
	"errors"
)

type MachineIdType uint8

const (
	Random    MachineIdType = 0
	AddressV4 MachineIdType = 1
	AddressV6 MachineIdType = 2
)

var ErrInvalidConnectionId = errors.New("invalid connection id")

var (
	crockfordEncoding = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	b32Enc            = base32.NewEncoding(crockfordEncoding).WithPadding(base32.NoPadding)
	bigEndian         = binary.BigEndian
)

func beReadU48(b []byte) uint64 {
	_ = b[5]
	return uint64(b[5]) | uint64(b[4])<<8 | uint64(b[3])<<16 | uint64(b[2])<<24 |
		uint64(b[1])<<32 | uint64(b[0])<<40
}

type ConnectionId struct {
	Msec          uint64
	MachineIdType MachineIdType
	MachineId     string
	Incr          uint16
	Rand          uint16
	Version       uint8
}

func ParseConnectionId(s string) (ConnectionId, error) {
	id := ConnectionId{}
	if len(s) != 45 {
		return id, ErrInvalidConnectionId
	}
	buf, err := b32Enc.DecodeString(s)
	if err != nil {
		return id, ErrInvalidConnectionId
	}
	msecAndMachineIdType := beReadU48(buf[:6])
	id.Msec = msecAndMachineIdType >> 2
	id.MachineIdType = MachineIdType(msecAndMachineIdType & 0x3)
	id.MachineId = b32Enc.EncodeToString(buf[6:24])
	id.Incr = bigEndian.Uint16(buf[24:26])
	randAndVer := bigEndian.Uint16(buf[26:28])
	id.Rand = randAndVer >> 2
	id.Version = uint8(randAndVer & 0x3)
	return id, nil
}
