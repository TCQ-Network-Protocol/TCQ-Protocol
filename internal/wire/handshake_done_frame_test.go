//	Project: TCQ Network Protocol (Thread Controlled QUIC)
//	Author: Trần Nguyên Hiền (c)
//	Major: Electronic And Communication Engineering
//	Email: trannguyenhien29085@gmail.com
//	Date: 2/3/2026
//	Apache License 2.0
//
// ----------------------------------------------------------------
package wire

import (
	"testing"

	"github.com/NguyenHien-8/tcq-network-protocol/internal/protocol"
	"github.com/stretchr/testify/require"
)

func TestWriteHandshakeDoneSampleFrame(t *testing.T) {
	frame := HandshakeDoneFrame{}
	b, err := frame.Append(nil, protocol.Version1)
	require.NoError(t, err)
	require.Equal(t, []byte{byte(FrameTypeHandshakeDone)}, b)
	require.Equal(t, protocol.ByteCount(1), frame.Length(protocol.Version1))
}
