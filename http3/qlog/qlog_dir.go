//	Project: TCQ Network Protocol (Thread Controlled QUIC)
//	Author: Trần Nguyên Hiền (c)
//	Major: Electronic And Communication Engineering
//	Email: trannguyenhien29085@gmail.com
//	Date: 2/3/2026
//	MIT Licence
//
// ----------------------------------------------------------------
package qlog

import (
	"context"

	quic "github.com/NguyenHien-8/tcq-network-protocol"
	"github.com/NguyenHien-8/tcq-network-protocol/qlog"
	"github.com/NguyenHien-8/tcq-network-protocol/qlogwriter"
)

const EventSchema = "urn:ietf:params:qlog:events:http3-12"

func DefaultConnectionTracer(ctx context.Context, isClient bool, connID quic.ConnectionID) qlogwriter.Trace {
	return qlog.DefaultConnectionTracerWithSchemas(ctx, isClient, connID, []string{qlog.EventSchema, EventSchema})
}
