//  Project: TCQ Network Protocol (Thread Controlled QUIC)
//  Author: Trần Nguyên Hiền (c)
//  Major: Electronic And Communication Engineering
//  Email: trannguyenhien29085@gmail.com
//  Date: 2/3/2026
//	Apache License 2.0
//
// ----------------------------------------------------------------
//go:build gomock || generate

package quic

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/NguyenHien-8/tcq-network-protocol -destination mock_send_conn_test.go github.com/NguyenHien-8/tcq-network-protocol SendConn"
type SendConn = sendConn

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/NguyenHien-8/tcq-network-protocol -destination mock_raw_conn_test.go github.com/NguyenHien-8/tcq-network-protocol RawConn"
type RawConn = rawConn

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/NguyenHien-8/tcq-network-protocol -destination mock_sender_test.go github.com/NguyenHien-8/tcq-network-protocol Sender"
type Sender = sender

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/NguyenHien-8/tcq-network-protocol -destination mock_stream_sender_test.go github.com/NguyenHien-8/tcq-network-protocol StreamSender"
type StreamSender = streamSender

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/NguyenHien-8/tcq-network-protocol -destination mock_stream_control_frame_getter_test.go github.com/NguyenHien-8/tcq-network-protocol StreamControlFrameGetter"
type StreamControlFrameGetter = streamControlFrameGetter

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/NguyenHien-8/tcq-network-protocol -destination mock_stream_frame_getter_test.go github.com/NguyenHien-8/tcq-network-protocol StreamFrameGetter"
type StreamFrameGetter = streamFrameGetter

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/NguyenHien-8/tcq-network-protocol -destination mock_frame_source_test.go github.com/NguyenHien-8/tcq-network-protocol FrameSource"
type FrameSource = frameSource

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/NguyenHien-8/tcq-network-protocol -destination mock_ack_frame_source_test.go github.com/NguyenHien-8/tcq-network-protocol AckFrameSource"
type AckFrameSource = ackFrameSource

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/NguyenHien-8/tcq-network-protocol -destination mock_sealing_manager_test.go github.com/NguyenHien-8/tcq-network-protocol SealingManager"
type SealingManager = sealingManager

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/NguyenHien-8/tcq-network-protocol -destination mock_unpacker_test.go github.com/NguyenHien-8/tcq-network-protocol Unpacker"
type Unpacker = unpacker

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/NguyenHien-8/tcq-network-protocol -destination mock_packer_test.go github.com/NguyenHien-8/tcq-network-protocol Packer"
type Packer = packer

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/NguyenHien-8/tcq-network-protocol -destination mock_mtu_discoverer_test.go github.com/NguyenHien-8/tcq-network-protocol MTUDiscoverer"
type MTUDiscoverer = mtuDiscoverer

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/NguyenHien-8/tcq-network-protocol -destination mock_conn_runner_test.go github.com/NguyenHien-8/tcq-network-protocol ConnRunner"
type ConnRunner = connRunner

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/NguyenHien-8/tcq-network-protocol -destination mock_packet_handler_test.go github.com/NguyenHien-8/tcq-network-protocol PacketHandler"
type PacketHandler = packetHandler

//go:generate sh -c "go tool mockgen -typed -package quic -self_package github.com/NguyenHien-8/tcq-network-protocol -self_package github.com/NguyenHien-8/tcq-network-protocol -destination mock_packetconn_test.go net PacketConn"
