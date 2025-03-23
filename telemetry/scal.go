package telemetry

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var ErrInvalidSCALLength = errors.New("invalid SCAL length")

// Scale - contains slice of multipliers for subsequent data.
type SCAL struct {
	Values []int
}

func (scale *SCAL) Parse(bytes []byte, size int64) error {
	s := int(size)

	if len(bytes)%s != 0 {
		return fmt.Errorf("scal: %w", ErrInvalidTelemLength)
	}

	if s == 2 {
		for i := 0; i < len(bytes); i += s {
			scale.Values = append(scale.Values, int(binary.BigEndian.Uint16(bytes[i:i+s])))
		}
	} else if s == 4 {
		for i := 0; i < len(bytes); i += s {
			scale.Values = append(scale.Values, int(binary.BigEndian.Uint32(bytes[i:i+s])))
		}
	} else {
		return fmt.Errorf("scal: %w", ErrInvalidSCALLength)
	}

	return nil
}
