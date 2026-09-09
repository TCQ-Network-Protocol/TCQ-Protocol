//	Project: TCQ Network Protocol (Thread Controlled QUIC)
//	Author: Trần Nguyên Hiền (c)
//	Major: Electronic And Communication Engineering
//	Email: trannguyenhien29085@gmail.com
//	Date: 2/3/2026
//	Apache License 2.0
//
// ----------------------------------------------------------------
package congestion

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBandwidthFromDelta(t *testing.T) {
	require.Equal(t, 1000*BytesPerSecond, BandwidthFromDelta(1, time.Millisecond))
}
