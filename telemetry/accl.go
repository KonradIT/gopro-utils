package telemetry

import (
	"encoding/binary"
	"fmt"
)

// Accelerometer in m/s for XYZ.
type ACCL struct {
	X float64
	Y float64
	Z float64
}

func (accl *ACCL) Parse(bytes []byte, scale *SCAL) error {
	if len(bytes) != 6 {
		return fmt.Errorf("accl: %w", ErrInvalidTelemLength)
	}

	accl.X = uint16ToInt16Float64(binary.BigEndian.Uint16(bytes[0:2]), float64(scale.Values[0]))
	accl.Y = uint16ToInt16Float64(binary.BigEndian.Uint16(bytes[2:4]), float64(scale.Values[0]))
	accl.Z = uint16ToInt16Float64(binary.BigEndian.Uint16(bytes[4:6]), float64(scale.Values[0]))

	return nil
}

func uint16ToInt16Float64(value uint16, scale float64) float64 {
	return float64(int16(value)) / scale //nolint:gosec // unsigned to signed is fine for now.
}
