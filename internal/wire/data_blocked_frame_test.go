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
	"io"
	"testing"

	"github.com/NguyenHien-8/tcq-network-protocol/internal/protocol"
	"github.com/NguyenHien-8/tcq-network-protocol/quicvarint"

	"github.com/stretchr/testify/require"
)

func TestParseDataBlocked(t *testing.T) {
	data := encodeVarInt(0x12345678)
	frame, l, err := parseDataBlockedFrame(data, protocol.Version1)
	require.NoError(t, err)
	require.Equal(t, protocol.ByteCount(0x12345678), frame.MaximumData)
	require.Equal(t, len(data), l)
}

func TestParseDataBlockedErrorsOnEOFs(t *testing.T) {
	data := encodeVarInt(0x12345678)
	_, l, err := parseDataBlockedFrame(data, protocol.Version1)
	require.NoError(t, err)
	require.Equal(t, len(data), l)
	for i := range data {
		_, _, err := parseDataBlockedFrame(data[:i], protocol.Version1)
		require.Equal(t, io.EOF, err)
	}
}

func TestWriteDataBlocked(t *testing.T) {
	frame := DataBlockedFrame{MaximumData: 0xdeadbeef}
	b, err := frame.Append(nil, protocol.Version1)
	require.NoError(t, err)
	expected := []byte{byte(FrameTypeDataBlocked)}
	expected = append(expected, encodeVarInt(0xdeadbeef)...)
	require.Equal(t, expected, b)
	require.Equal(t, protocol.ByteCount(1+quicvarint.Len(uint64(frame.MaximumData))), frame.Length(protocol.Version1))
}
