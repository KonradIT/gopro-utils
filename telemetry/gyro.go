package telemetry

import (
	"encoding/binary"
	"fmt"
)

// 3-axis Gyroscope data in rad/s.
type GYRO struct {
	X float64
	Y float64
	Z float64
}

func (gyro *GYRO) Parse(bytes []byte, scale *SCAL) error {
	if len(bytes) != 6 {
		return fmt.Errorf("gyro: %w", ErrInvalidTelemLength)
	}

	gyro.X = uint16ToInt16Float64(binary.BigEndian.Uint16(bytes[0:2]), float64(scale.Values[0]))
	gyro.Y = uint16ToInt16Float64(binary.BigEndian.Uint16(bytes[2:4]), float64(scale.Values[0]))
	gyro.Z = uint16ToInt16Float64(binary.BigEndian.Uint16(bytes[4:6]), float64(scale.Values[0]))

	return nil
}
