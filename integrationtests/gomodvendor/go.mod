module test

go 1.25

// The version doesn't matter here, as we're replacing it with the currently checked out code anyway.
require github.com/NguyenHien-8/tcq-network-protocol v0.1.0

require (
	github.com/NguyenHien-8/qpack-http3 v0.1.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)

replace github.com/NguyenHien-8/tcq-network-protocol => ../../
