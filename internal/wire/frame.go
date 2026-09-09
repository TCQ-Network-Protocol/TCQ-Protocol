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
)

// A Frame in QUIC
type Frame interface {
	Append(b []byte, version protocol.Version) ([]byte, error)
	Length(version protocol.Version) protocol.ByteCount
}

// IsProbingFrame returns true if the frame is a probing frame.
// See section 9.1 of RFC 9000.
func IsProbingFrame(f Frame) bool {
	switch f.(type) {
	case *PathChallengeFrame, *PathResponseFrame, *NewConnectionIDFrame:
		return true
	}
	return false
}

// IsProbingFrameType returns true if the FrameType is a probing frame.
// See section 9.1 of RFC 9000.
func IsProbingFrameType(f FrameType) bool {
	//nolint:exhaustive // PATH_CHALLENGE, PATH_RESPONSE and NEW_CONNECTION_ID are the only probing frames
	switch f {
	case FrameTypePathChallenge, FrameTypePathResponse, FrameTypeNewConnectionID:
		return true
	default:
		return false
	}
}
