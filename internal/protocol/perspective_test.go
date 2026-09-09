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

func TestPerspectiveOpposite(t *testing.T) {
	require.Equal(t, PerspectiveServer, PerspectiveClient.Opposite())
	require.Equal(t, PerspectiveClient, PerspectiveServer.Opposite())
}

func TestPerspectiveStringer(t *testing.T) {
	require.Equal(t, "client", PerspectiveClient.String())
	require.Equal(t, "server", PerspectiveServer.String())
	require.Equal(t, "invalid perspective", Perspective(0).String())
}
