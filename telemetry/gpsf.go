package telemetry

import (
	"encoding/binary"
	"fmt"
)

// GPS Fix.
type GPSF struct {
	F uint32
}

func (gpsf *GPSF) Parse(bytes []byte) error {
	if len(bytes) != 4 {
		return fmt.Errorf("gpsf: %w", ErrInvalidTelemLength)
	}

	gpsf.F = binary.BigEndian.Uint32(bytes[0:4])

	return nil
}
