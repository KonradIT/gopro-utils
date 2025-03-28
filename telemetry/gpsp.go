package telemetry

import (
	"encoding/binary"
	"fmt"
)

// GPS position accuracy in cm.
type GPSP struct {
	Accuracy uint16
}

func (gpsp *GPSP) Parse(bytes []byte) error {
	if len(bytes) != 2 {
		return fmt.Errorf("gpsp: %w", ErrInvalidTelemLength)
	}

	gpsp.Accuracy = binary.BigEndian.Uint16(bytes[0:2])

	return nil
}
