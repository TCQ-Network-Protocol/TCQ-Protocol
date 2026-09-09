//	Project: TCQ Network Protocol (Thread Controlled QUIC)
//	Author: Trần Nguyên Hiền (c)
//	Major: Electronic And Communication Engineering
//	Email: trannguyenhien29085@gmail.com
//	Date: 2/3/2026
//	Apache License 2.0
//
// ----------------------------------------------------------------
package quic

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"hash"
	"sync"

	"github.com/NguyenHien-8/tcq-network-protocol/internal/protocol"
)

type statelessResetter struct {
	mx sync.Mutex
	h  hash.Hash
}

// newStatelessRetter creates a new stateless reset generator.
// It is valid to use a nil key. In that case, a random key will be used.
// This makes is impossible for on-path attackers to shut down established connections.
func newStatelessResetter(key *StatelessResetKey) *statelessResetter {
	var h hash.Hash
	if key != nil {
		h = hmac.New(sha256.New, key[:])
	} else {
		b := make([]byte, 32)
		_, _ = rand.Read(b)
		h = hmac.New(sha256.New, b)
	}
	return &statelessResetter{h: h}
}

func (r *statelessResetter) GetStatelessResetToken(connID protocol.ConnectionID) protocol.StatelessResetToken {
	r.mx.Lock()
	defer r.mx.Unlock()

	var token protocol.StatelessResetToken
	r.h.Write(connID.Bytes())
	copy(token[:], r.h.Sum(nil))
	r.h.Reset()
	return token
}
