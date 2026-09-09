//	Project: TCQ Network Protocol (Thread Controlled QUIC)
//	Author: Trần Nguyên Hiền (c)
//	Major: Electronic And Communication Engineering
//	Email: trannguyenhien29085@gmail.com
//	Date: 2/3/2026
//	Apache License 2.0
//
// ----------------------------------------------------------------
package ringbuffer

import "testing"

func BenchmarkRingBuffer(b *testing.B) {
	r := RingBuffer[int]{}

	var val int
	for b.Loop() {
		r.PushBack(val)
		r.PopFront()
		val++
	}
}
