package seq

import (
	"encoding/hex"
	"testing"

	"github.com/google/uuid"

	"github.com/go-sdk/core/testx"
)

func TestUUID(t *testing.T) {
	value := UUID()
	parsed, err := uuid.Parse(value)

	testx.NoError(t, err)
	testx.Equal(t, uuid.Version(7), parsed.Version())
	testx.Equal(t, uuid.RFC4122, parsed.Variant())
	testx.Equal(t, value, parsed.String())
}

func TestUUIDShort(t *testing.T) {
	value := UUIDShort()

	testx.Len(t, value, 32)
	testx.NotContains(t, value, "-")
	_, err := hex.DecodeString(value)
	testx.NoError(t, err)

	formatted := value[:8] + "-" + value[8:12] + "-" + value[12:16] + "-" + value[16:20] + "-" + value[20:]
	parsed, err := uuid.Parse(formatted)
	testx.NoError(t, err)
	testx.Equal(t, uuid.Version(7), parsed.Version())
}

func TestUUIDConcurrent(t *testing.T) {
	const (
		workers   = 8
		perWorker = 250
	)

	values := make(chan string, workers*perWorker)
	for range workers {
		go func() {
			for range perWorker {
				values <- UUID()
			}
		}()
	}

	seen := make(map[string]struct{}, workers*perWorker)
	for range workers * perWorker {
		value := <-values
		_, exists := seen[value]
		testx.False(t, exists)
		seen[value] = struct{}{}
	}
}
