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

	"github.com/stretchr/testify/require"
)

func TestAckRangeLength(t *testing.T) {
	require.EqualValues(t, 1, AckRange{Smallest: 10, Largest: 10}.Len())
	require.EqualValues(t, 4, AckRange{Smallest: 10, Largest: 13}.Len())
}
