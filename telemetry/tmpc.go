package telemetry

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Temperature in °C.
type TMPC struct {
	Temp float32
}

func (temp *TMPC) Parse(bytes []byte) error {
	if len(bytes) != 4 {
		return fmt.Errorf("tmpc: %w", ErrInvalidTelemLength)
	}

	bits := binary.BigEndian.Uint32(bytes[0:4])

	temp.Temp = math.Float32frombits(bits)

	return nil
}
