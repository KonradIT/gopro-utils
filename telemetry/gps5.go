package telemetry

import (
	"encoding/binary"
	"fmt"
)

// GPS sentence with lat/lon/alt/speed/3d speed.
type GPS5 struct {
	Latitude  float64 `json:"lat"`   // degrees lat
	Longitude float64 `json:"lon"`   // degrees lon
	Altitude  float64 `json:"alt"`   // meters above wgs84 ellipsoid ?
	Speed     float64 `json:"spd"`   // m/s
	Speed3D   float64 `json:"spd3d"` // m/s, standard error?
	TS        int64   `json:"utc"`
}

func uint32ToInt32Float64(value uint32, scale float64) float64 {
	return float64(int32(value)) / scale //nolint:gosec // unsigned to signed is fine for now.
}

func (gps *GPS5) Parse(bytes []byte, scale *SCAL) error {
	if len(bytes) != 20 {
		return fmt.Errorf("gps5: %w", ErrInvalidTelemLength)
	}

	gps.Latitude = uint32ToInt32Float64(binary.BigEndian.Uint32(bytes[0:4]), float64(scale.Values[0]))
	gps.Longitude = uint32ToInt32Float64(binary.BigEndian.Uint32(bytes[4:8]), float64(scale.Values[1]))

	// convert from mm
	gps.Altitude = uint32ToInt32Float64(binary.BigEndian.Uint32(bytes[8:12]), float64(scale.Values[2]))

	// convert from mm/s
	gps.Speed = uint32ToInt32Float64(binary.BigEndian.Uint32(bytes[12:16]), float64(scale.Values[3]))

	// convert from mm/s
	gps.Speed3D = uint32ToInt32Float64(binary.BigEndian.Uint32(bytes[16:20]), float64(scale.Values[4]))

	return nil
}
