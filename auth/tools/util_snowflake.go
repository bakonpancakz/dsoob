package tools

import (
	"sync/atomic"
	"time"
)

// Format: 42 bits timestamp (seconds) | 22 bits sequence

var (
	maxSequence int64 = (1 << 22) - 1 // 22 bits for sequence
	sequence    atomic.Int64
	timestamp   atomic.Int64
)

// GenerateSnowflake returns an int64-safe unique ID
func GenerateSnowflake() int64 {
	now := time.Now().Unix()

	// Reset sequence if time has advanced
	if now != timestamp.Load() {
		sequence.Store(0)
		timestamp.Store(now)
	} else {
		seq := sequence.Add(1)
		if seq > maxSequence {
			// Sequence overflowed: wait for next second
			for now <= timestamp.Load() {
				time.Sleep(time.Millisecond * 10)
				now = time.Now().Unix()
			}
			sequence.Store(0)
			timestamp.Store(now)
			seq = 0
		}
	}

	id := ((now - EPOCH_SECONDS) << 22) | sequence.Load()
	return id
}
