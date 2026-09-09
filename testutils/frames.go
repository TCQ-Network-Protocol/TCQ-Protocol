//	Project: TCQ Network Protocol (Thread Controlled QUIC)
//	Author: Trần Nguyên Hiền (c)
//	Major: Electronic And Communication Engineering
//	Email: trannguyenhien29085@gmail.com
//	Date: 2/3/2026
//	Apache License 2.0
//
// ----------------------------------------------------------------
package testutils

import "github.com/NguyenHien-8/tcq-network-protocol/internal/wire"

type (
	Frame                   = wire.Frame
	AckFrame                = wire.AckFrame
	ConnectionCloseFrame    = wire.ConnectionCloseFrame
	CryptoFrame             = wire.CryptoFrame
	DataBlockedFrame        = wire.DataBlockedFrame
	HandshakeDoneFrame      = wire.HandshakeDoneFrame
	MaxDataFrame            = wire.MaxDataFrame
	MaxStreamDataFrame      = wire.MaxStreamDataFrame
	MaxStreamsFrame         = wire.MaxStreamsFrame
	NewConnectionIDFrame    = wire.NewConnectionIDFrame
	NewTokenFrame           = wire.NewTokenFrame
	PathChallengeFrame      = wire.PathChallengeFrame
	PathResponseFrame       = wire.PathResponseFrame
	PingFrame               = wire.PingFrame
	ResetStreamFrame        = wire.ResetStreamFrame
	RetireConnectionIDFrame = wire.RetireConnectionIDFrame
	StopSendingFrame        = wire.StopSendingFrame
	StreamDataBlockedFrame  = wire.StreamDataBlockedFrame
	StreamFrame             = wire.StreamFrame
	StreamsBlockedFrame     = wire.StreamsBlockedFrame
)
