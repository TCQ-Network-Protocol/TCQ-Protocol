//	Project: TCQ Network Protocol (Thread Controlled QUIC)
//	Author: Trần Nguyên Hiền (c)
//	Major: Electronic And Communication Engineering
//	Email: trannguyenhien29085@gmail.com
//	Date: 2/3/2026
//	Apache License 2.0
//
// ----------------------------------------------------------------
package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPacketQueueCapacities(t *testing.T) {
	// Ensure that the session can queue more packets than the 0-RTT queue
	require.Greater(t, MaxConnUnprocessedPackets, Max0RTTQueueLen)
	require.Greater(t, MaxUndecryptablePackets, Max0RTTQueueLen)
}
