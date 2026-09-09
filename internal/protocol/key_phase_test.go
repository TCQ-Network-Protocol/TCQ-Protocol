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

func TestKeyPhaseBitDefaultValue(t *testing.T) {
	var k KeyPhaseBit
	require.Equal(t, KeyPhaseUndefined, k)
}

func TestKeyPhaseStringRepresentation(t *testing.T) {
	require.Equal(t, "0", KeyPhaseZero.String())
	require.Equal(t, "1", KeyPhaseOne.String())
}

func TestKeyPhaseToBit(t *testing.T) {
	require.Equal(t, KeyPhaseZero, KeyPhase(0).Bit())
	require.Equal(t, KeyPhaseZero, KeyPhase(2).Bit())
	require.Equal(t, KeyPhaseZero, KeyPhase(4).Bit())
	require.Equal(t, KeyPhaseOne, KeyPhase(1).Bit())
	require.Equal(t, KeyPhaseOne, KeyPhase(3).Bit())
	require.Equal(t, KeyPhaseOne, KeyPhase(5).Bit())
}
