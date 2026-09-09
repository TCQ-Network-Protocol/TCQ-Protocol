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
	"github.com/NguyenHien-8/tcq-network-protocol/quicvarint"
	"github.com/stretchr/testify/require"
)

func TestImmediateAckFrame(t *testing.T) {
	frame := ImmediateAckFrame{}
	b, err := frame.Append(nil, protocol.Version1)
	require.NoError(t, err)

	val, l, err := quicvarint.Parse(b)
	require.NoError(t, err)
	require.Equal(t, uint64(FrameTypeImmediateAck), val)
	require.Equal(t, len(b), l)

	require.Len(t, b, int(frame.Length(protocol.Version1)))
}
