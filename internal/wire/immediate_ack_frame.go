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
	"github.com/NguyenHien-8/tcq-network-protocol/internal/protocol"
	"github.com/NguyenHien-8/tcq-network-protocol/quicvarint"
)

// An ImmediateAckFrame is an IMMEDIATE_ACK frame
type ImmediateAckFrame struct{}

func (f *ImmediateAckFrame) Append(b []byte, _ protocol.Version) ([]byte, error) {
	return quicvarint.Append(b, uint64(FrameTypeImmediateAck)), nil
}

// Length of a written frame
func (f *ImmediateAckFrame) Length(_ protocol.Version) protocol.ByteCount {
	return protocol.ByteCount(quicvarint.Len(uint64(FrameTypeImmediateAck)))
}
