package telemetry

import (
	"fmt"
	"time"
)

// GPS-acquired timestamp.
type GPSU struct {
	Time time.Time
}

func (gpsu *GPSU) Parse(bytes []byte) error {
	if len(bytes) != 16 {
		return fmt.Errorf("gpsu: %w", ErrInvalidTelemLength)
	}

	t, err := time.Parse("060102150405", string(bytes))
	if err != nil {
		return err
	}

	gpsu.Time = t

	return nil
}
