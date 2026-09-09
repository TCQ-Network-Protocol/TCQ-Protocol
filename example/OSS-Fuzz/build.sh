#!/bin/bash -eu

go mod download
compile_native_go_fuzzer_v2 github.com/NguyenHien-8/tcq-network-protocol/internal/wire FuzzFrames frame_fuzzer
compile_native_go_fuzzer_v2 github.com/NguyenHien-8/tcq-network-protocol/internal/wire FuzzTransportParameters transportparameter_fuzzer
compile_native_go_fuzzer_v2 github.com/NguyenHien-8/tcq-network-protocol/http3 FuzzFrameParser http3_frame_fuzzer
compile_native_go_fuzzer_v2 github.com/NguyenHien-8/tcq-network-protocol/internal/wire FuzzHeaderParser header_fuzzer
compile_native_go_fuzzer_v2 github.com/NguyenHien-8/tcq-network-protocol/internal/handshake FuzzHandshake handshake_fuzzer
compile_native_go_fuzzer_v2 github.com/NguyenHien-8/tcq-network-protocol/http3 FuzzHeaderParsing http3_header_parsing_fuzzer