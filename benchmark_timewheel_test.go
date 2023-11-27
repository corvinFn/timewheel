package timewheel

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func BenchmarkTimeWheelTest(b *testing.B) {
	const delay = 10 * time.Millisecond
	num := int32(0)
	tw, _ := NewTimeWheel(delay, 3600)
	tw.activate()
	defer tw.stop()
	f := func() {
		atomic.AddInt32(&num, 1)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tw.addTimer(delay, f, false)
	}
	time.Sleep(3 * delay)
	require.EqualValues(b, b.N, atomic.LoadInt32(&num))
}

func BenchmarkTimeWheelParallelTest(b *testing.B) {
	const delay = 10 * time.Millisecond
	num := int32(0)
	tw, _ := NewTimeWheel(delay, 3600)
	tw.activate()
	defer tw.stop()
	f := func() {
		atomic.AddInt32(&num, 1)
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tw.addTimer(delay, f, false)
		}
	})
	time.Sleep(3 * delay)
	require.EqualValues(b, b.N, atomic.LoadInt32(&num))
}
