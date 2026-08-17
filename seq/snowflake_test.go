package seq

import (
	"strconv"
	"testing"

	"github.com/go-sdk/core/testx"
)

func TestNextID(t *testing.T) {
	const count = 1000

	seen := make(map[string]struct{}, count)
	var previous int64 = -1
	for range count {
		id := NextID()
		value, err := strconv.ParseInt(id, 10, 64)

		testx.NoError(t, err)
		testx.Greater(t, value, previous)
		testx.Equal(t, int64(56565), value&0xffff)
		_, exists := seen[id]
		testx.False(t, exists)

		seen[id] = struct{}{}
		previous = value
	}
}

func TestNextIDConcurrent(t *testing.T) {
	const (
		workers   = 8
		perWorker = 250
	)

	ids := make(chan string, workers*perWorker)
	for range workers {
		go func() {
			for range perWorker {
				ids <- NextID()
			}
		}()
	}

	seen := make(map[string]struct{}, workers*perWorker)
	for range workers * perWorker {
		id := <-ids
		_, exists := seen[id]
		testx.False(t, exists)
		seen[id] = struct{}{}
	}
}
