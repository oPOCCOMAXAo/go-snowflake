package snowflake

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNextMonotonic(t *testing.T) {
	g, err := New(Config{
		MachineID: 10,
	})
	require.NoError(t, err)

	out := make([]uint64, 10000)

	for i := range out {
		out[i] = g.Next()
	}

	// ensure they are all distinct and increasing
	for i := range out[1:] {
		if out[i] >= out[i+1] {
			t.Fatal("bad entries:", out[i], out[i+1])
		}
	}
}

var blackhole uint64 //nolint:gochecknoglobals // to make sure the g.Next calls are not removed

func BenchmarkNext(b *testing.B) {
	g, err := New(Config{
		MachineID: 10,
	})
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		blackhole += g.Next()
	}
}

func BenchmarkNextParallel(b *testing.B) {
	g, err := New(Config{
		MachineID: 1,
	})
	require.NoError(b, err)

	b.RunParallel(func(pb *testing.PB) {
		var lblackhole uint64
		for pb.Next() {
			lblackhole += g.Next()
		}

		atomic.AddUint64(&blackhole, lblackhole)
	})
}
