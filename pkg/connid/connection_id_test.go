package connid

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestParseConnectionId(t *testing.T) {
	str := "0QNFX42SPVYVVQ0203ZG00800812A09Q05BK0E9G74JA8"
	dt, _ := time.Parse(time.RFC3339Nano, "2021-07-20T00:40:39.789456000+08:00")
	ip := net.ParseIP("fdbd:dc02:ff:1:2:225:137:157")
	port := 12345
	machineIdBuf := make([]byte, 18)
	copy(machineIdBuf[:16], ip[:16])
	bigEndian.PutUint16(machineIdBuf[16:18], uint16(port))

	connId, err := ParseConnectionId(str)
	assert.Nil(t, err)
	assert.Equal(t, uint64(dt.Truncate(time.Millisecond).UnixNano()), connId.Msec*1e6)
	assert.Equal(t, uint16(12345), connId.Incr)
	assert.Equal(t, uint16(2345), connId.Rand)
	assert.Equal(t, uint8(0), connId.Version)
	assert.Equal(t, AddressV6, connId.MachineIdType)
	assert.Equal(t, b32Enc.EncodeToString(machineIdBuf), connId.MachineId)
}
