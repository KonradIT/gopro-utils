package telemetry

import (
	"encoding/binary"
	"fmt"
)

// Total number of samples.
type TSMP struct {
	Samples uint32
}

func (t *TSMP) Parse(bytes []byte) error {
	if len(bytes) != 4 {
		return fmt.Errorf("tsmp: %w", ErrInvalidTelemLength)
	}

	t.Samples = binary.BigEndian.Uint32(bytes[0:4])

	return nil
}
