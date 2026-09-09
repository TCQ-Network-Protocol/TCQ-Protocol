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

	"github.com/stretchr/testify/require"
)

func TestParseRetireConnectionID(t *testing.T) {
	data := encodeVarInt(0xdeadbeef) // sequence number
	frame, l, err := parseRetireConnectionIDFrame(data, protocol.Version1)
	require.NoError(t, err)
	require.Equal(t, uint64(0xdeadbeef), frame.SequenceNumber)
	require.Equal(t, len(data), l)
}

func TestParseRetireConnectionIDErrorsOnEOFs(t *testing.T) {
	data := encodeVarInt(0xdeadbeef) // sequence number
	_, l, err := parseRetireConnectionIDFrame(data, protocol.Version1)
	require.NoError(t, err)
	require.Equal(t, len(data), l)
	for i := range data {
		_, _, err := parseRetireConnectionIDFrame(data[:i], protocol.Version1)
		require.Equal(t, io.EOF, err)
	}
}

func TestWriteRetireConnectionID(t *testing.T) {
	frame := &RetireConnectionIDFrame{SequenceNumber: 0x1337}
	b, err := frame.Append(nil, protocol.Version1)
	require.NoError(t, err)
	expected := []byte{byte(FrameTypeRetireConnectionID)}
	expected = append(expected, encodeVarInt(0x1337)...)
	require.Equal(t, expected, b)
	require.Len(t, b, int(frame.Length(protocol.Version1)))
}
