//	Project: TCQ Network Protocol (Thread Controlled QUIC)
//	Author: Trần Nguyên Hiền (c)
//	Major: Electronic And Communication Engineering
//	Email: trannguyenhien29085@gmail.com
//	Date: 2/3/2026
//	Apache License 2.0
//
// ----------------------------------------------------------------
package testdata

import (
	"crypto/tls"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCertificates(t *testing.T) {
	ln, err := tls.Listen("tcp", "localhost:4433", GetTLSConfig())
	require.NoError(t, err)

	go func() {
		conn, err := ln.Accept()
		require.NoError(t, err)
		defer conn.Close()
		_, err = conn.Write([]byte("foobar"))
		require.NoError(t, err)
	}()

	conn, err := tls.Dial("tcp", "localhost:4433", &tls.Config{RootCAs: GetRootCA()})
	require.NoError(t, err)
	data, err := io.ReadAll(conn)
	require.NoError(t, err)
	require.Equal(t, "foobar", string(data))
}
