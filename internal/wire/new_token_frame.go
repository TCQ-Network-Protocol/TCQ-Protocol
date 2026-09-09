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
	"errors"
	"io"

	"github.com/NguyenHien-8/tcq-network-protocol/internal/protocol"
	"github.com/NguyenHien-8/tcq-network-protocol/quicvarint"
)

// A NewTokenFrame is a NEW_TOKEN frame
type NewTokenFrame struct {
	Token []byte
}

func parseNewTokenFrame(b []byte, _ protocol.Version) (*NewTokenFrame, int, error) {
	tokenLen, l, err := quicvarint.Parse(b)
	if err != nil {
		return nil, 0, replaceUnexpectedEOF(err)
	}
	b = b[l:]
	if tokenLen == 0 {
		return nil, 0, errors.New("token must not be empty")
	}
	if uint64(len(b)) < tokenLen {
		return nil, 0, io.EOF
	}
	token := make([]byte, int(tokenLen))
	copy(token, b)
	return &NewTokenFrame{Token: token}, l + int(tokenLen), nil
}

func (f *NewTokenFrame) Append(b []byte, _ protocol.Version) ([]byte, error) {
	b = append(b, byte(FrameTypeNewToken))
	b = quicvarint.Append(b, uint64(len(f.Token)))
	b = append(b, f.Token...)
	return b, nil
}

// Length of a written frame
func (f *NewTokenFrame) Length(protocol.Version) protocol.ByteCount {
	return 1 + protocol.ByteCount(quicvarint.Len(uint64(len(f.Token)))+len(f.Token))
}
