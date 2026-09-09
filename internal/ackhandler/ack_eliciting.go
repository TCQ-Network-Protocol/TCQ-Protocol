//	Project: TCQ Network Protocol (Thread Controlled QUIC)
//	Author: Trần Nguyên Hiền (c)
//	Major: Electronic And Communication Engineering
//	Email: trannguyenhien29085@gmail.com
//	Date: 2/3/2026
//	Apache License 2.0
//
// ----------------------------------------------------------------
package ackhandler

import "github.com/NguyenHien-8/tcq-network-protocol/internal/wire"

// IsFrameTypeAckEliciting returns true if the frame is ack-eliciting.
func IsFrameTypeAckEliciting(t wire.FrameType) bool {
	//nolint:exhaustive // The default case catches the rest.
	switch t {
	case wire.FrameTypeAck, wire.FrameTypeAckECN:
		return false
	case wire.FrameTypeConnectionClose, wire.FrameTypeApplicationClose:
		return false
	default:
		return true
	}
}

// IsFrameAckEliciting returns true if the frame is ack-eliciting.
func IsFrameAckEliciting(f wire.Frame) bool {
	_, isAck := f.(*wire.AckFrame)
	_, isConnectionClose := f.(*wire.ConnectionCloseFrame)
	return !isAck && !isConnectionClose
}

// HasAckElicitingFrames returns true if at least one frame is ack-eliciting.
func HasAckElicitingFrames(fs []Frame) bool {
	for _, f := range fs {
		if IsFrameAckEliciting(f.Frame) {
			return true
		}
	}
	return false
}
