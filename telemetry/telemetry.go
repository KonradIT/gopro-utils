package telemetry

import (
	"errors"
	"time"

	"github.com/paulmach/orb"
	geo "github.com/paulmach/orb/geo"
)

var ErrInvalidTelemLength = errors.New("invalid telemetry length")

// Represents one second of telemetry data.
type Telem struct {
	Accl        []ACCL
	Gps         []GPS5
	Gyro        []GYRO
	GpsFix      GPSF
	GpsAccuracy GPSP
	Time        GPSU
	Temp        TMPC
}

// GPS data might have a generated timestamp and derived track.
type TelemOut struct {
	*GPS5

	GpsAccuracy uint16  `json:"gps_accuracy,omitempty"`
	GpsFix      uint32  `json:"gps_fix,omitempty"`
	Temp        float32 `json:"temp,omitempty"`
	Track       float64 `json:"track,omitempty"`
}

var (
	pp                    = orb.Point{10, 10}
	lastGoodTrack float64 = 0
)

// zeroes out the telem struct.
func (t *Telem) Clear() {
	t.Accl = t.Accl[:0]
	t.Gps = t.Gps[:0]
	t.Gyro = t.Gyro[:0]
	t.Time.Time = time.Time{}
}

// determines if the telem has data.
func (t *Telem) IsZero() bool {
	// hack.
	return t.Time.Time.IsZero()
}

// try to populate a timestamp for every GPS row. probably bogus.
func (t *Telem) FillTimes(until time.Time) error {
	gpsLen := len(t.Gps)
	diff := until.Sub(t.Time.Time)

	offset := diff.Seconds() / float64(gpsLen)

	for i := range t.Gps {
		dur := time.Duration(float64(i)*offset*1000) * time.Millisecond
		ts := t.Time.Time.Add(dur)
		t.Gps[i].TS = ts.UnixNano() / 1000
	}

	return nil
}

func (t *Telem) Json() []TelemOut {
	var out []TelemOut

	for i := range t.Gps {
		jobj := TelemOut{&t.Gps[i], 0, 0, 0, 0}
		if i == 0 {
			jobj.GpsAccuracy = t.GpsAccuracy.Accuracy
			jobj.GpsFix = t.GpsFix.F
			jobj.Temp = t.Temp.Temp
		}

		p := orb.Point{jobj.GPS5.Longitude, jobj.GPS5.Latitude}
		jobj.Track = geo.Bearing(pp, p)
		pp = p

		if jobj.Track < 0 {
			jobj.Track = 360 + jobj.Track
		}

		// only set the track if speed is over 1 m/s
		// if it's slower (eg, stopped) it will drift all over with the location
		if jobj.GPS5.Speed > 1 {
			lastGoodTrack = jobj.Track
		} else {
			jobj.Track = lastGoodTrack
		}

		out = append(out, jobj)
	}

	return out
}
