//	Project: TCQ Network Protocol (Thread Controlled QUIC)
//	Author: Trần Nguyên Hiền (c)
//	Major: Electronic And Communication Engineering
//	Email: trannguyenhien29085@gmail.com
//	Date: 2/3/2026
//	Apache License 2.0
//
// ----------------------------------------------------------------
package handshake

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/NguyenHien-8/tcq-network-protocol/internal/wire"
	"github.com/NguyenHien-8/tcq-network-protocol/quicvarint"
)

const sessionTicketRevision = 5

type sessionTicket struct {
	Parameters *wire.TransportParameters
}

func (t *sessionTicket) Marshal() []byte {
	b := make([]byte, 0, 256)
	b = quicvarint.Append(b, sessionTicketRevision)
	return t.Parameters.MarshalForSessionTicket(b)
}

func (t *sessionTicket) Unmarshal(b []byte) error {
	rev, l, err := quicvarint.Parse(b)
	if err != nil {
		return errors.New("failed to read session ticket revision")
	}
	b = b[l:]
	if rev != sessionTicketRevision {
		return fmt.Errorf("unknown session ticket revision: %d", rev)
	}
	var tp wire.TransportParameters
	if err := tp.UnmarshalFromSessionTicket(b); err != nil {
		return fmt.Errorf("unmarshaling transport parameters from session ticket failed: %s", err.Error())
	}
	t.Parameters = &tp
	return nil
}

const extraPrefix = "quic-go1"

func addSessionStateExtraPrefix(b []byte) []byte {
	return append([]byte(extraPrefix), b...)
}

func findSessionStateExtraData(extras [][]byte) []byte {
	prefix := []byte(extraPrefix)
	for _, extra := range extras {
		if len(extra) < len(prefix) || !bytes.Equal(prefix, extra[:len(prefix)]) {
			continue
		}
		return extra[len(prefix):]
	}
	return nil
}
