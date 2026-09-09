//	Project: TCQ Network Protocol (Thread Controlled QUIC)
//	Author: Trần Nguyên Hiền (c)
//	Major: Electronic And Communication Engineering
//	Email: trannguyenhien29085@gmail.com
//	Date: 2/3/2026
//	Apache License 2.0
//
// ----------------------------------------------------------------
package qlog

import (
	"testing"

	"github.com/NguyenHien-8/tcq-network-protocol/internal/protocol"

	"github.com/stretchr/testify/require"
)

func TestEncryptionLevelToPacketType(t *testing.T) {
	require.Equal(t, "initial", string(EncryptionLevelToPacketType(protocol.EncryptionInitial)))
	require.Equal(t, "handshake", string(EncryptionLevelToPacketType(protocol.EncryptionHandshake)))
	require.Equal(t, "0RTT", string(EncryptionLevelToPacketType(protocol.Encryption0RTT)))
	require.Equal(t, "1RTT", string(EncryptionLevelToPacketType(protocol.Encryption1RTT)))
}

func TestCalculateDatagramID(t *testing.T) {
	require.Equal(t, DatagramID(0xcbf43926), CalculateDatagramID([]byte("123456789")))
}
