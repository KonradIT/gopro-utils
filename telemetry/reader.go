package telemetry

import (
	"io"
	"slices"
)

type Parser interface {
	Parse(bytes []byte) error
}

var _ Parser = &GPSU{}
var _ Parser = &GPSP{}
var _ Parser = &GPSF{}
var _ Parser = &TMPC{}

func Read(f io.Reader) (*Telem, error) {
	labels := []string{
		"ACCL",
		"AALP",
		"CORI",
		"DEVC",
		"DISP",
		"DVID",
		"DVNM",
		"EMPT",
		"GPRO",
		"GPSA",
		"GPS5",
		"GPSF",
		"GPSP",
		"GPSU",
		"GRAV",
		"GYRO",
		"HD5.",
		"IORI",
		"ISOG",
		"WNDM",
		"MWET",
		"SCAL",
		"SHUT",
		"SIUN",
		"STMP",
		"STNM",
		"STRM",
		"TICK",
		"TMPC",
		"TSMP",
		"UNIT",
		"VPTS",
		"TYPE",
		"FACE",
		"FCNM",
		"ISOE",
		"WBAL",
		"WRGB",
		"MAGN",
		"ALLD",
		"MTRX",
		"ORIN",
		"ORIO",
		"YAVG",
		"UNIF",
		"SCEN",
		"HUES",
		"SROT",
		"TIMO",
		"MSKP",
		"LRVO",
		"LRVS",
		"LSKP",
		"VERS",
		"FMWR",
		"LINF",
		"CINF",
		"CASN",
		"MINF",
		"MUID",
		"CMOD",
		"MTYP",
		"OREN",
		"DZOM",
		"DZST",
		"SMTR",
		"PRTN",
		"PTWB",
		"PTSH",
		"PTCL",
		"EXPT",
		"PIMX",
		"PIMN",
		"PTEV",
		"RATE",
		"VFOV",
		"ZFOV",
		"EISE",
		"EISA",
		"HLVL",
		"AUPT",
		"AUDO",
		"BROD",
		"BRID",
		"PVUL",
		"PRJT",
		"SOFF",
		"CLKS",
		"CDAT",
		"PRNA",
		"PRNU",
		"SCAP",
		"CDTM",
		"DUST",
		"VRES",
		"VFPS",
		"HSGT",
		"BITR",
		"MMOD",
		"RAMP",
		"TZON",
		"CLKC",
		"DZMX",
		"FSKP",
	}

	label := make([]byte, 4) // 4 byte ascii label of data
	desc := make([]byte, 4)  // 4 byte description of length of data

	// keep a copy of the scale to apply to subsequent sentences
	s := SCAL{}

	// the full telemetry for this period
	t := &Telem{}

	for {
		// pick out the label
		read, err := f.Read(label)
		if err == io.EOF || read == 0 {
			return nil, err
		}

		label_string := string(label)

		if !slices.Contains(labels, label_string) {
			continue
		}

		// pick out the label description
		read, err = f.Read(desc)
		if err == io.EOF || read == 0 {
			break
		}

		// first byte is zero, there is no length
		if desc[0] == 0x0 {
			continue
		}

		// skip empty packets
		if label_string == "EMPT" {
			io.CopyN(io.Discard, f, 4)
			continue
		}

		// extract the size and length
		val_size := int64(desc[1])
		num_values := (int64(desc[2]) << 8) | int64(desc[3])
		length := val_size * num_values

		// uncomment to see label, type, size and length
		//fmt.Printf("%s (%c) of size %v and len %v\n", label, desc[0], val_size, length)

		if label_string == "SCAL" {
			value := make([]byte, val_size*num_values)
			read, err = f.Read(value)
			if err == io.EOF || read == 0 {
				return nil, err
			}

			// clear the scales
			s.Values = s.Values[:0]

			err := s.Parse(value, val_size)
			if err != nil {
				return nil, err
			}
		} else {
			value := make([]byte, val_size)

			for i := int64(0); i < num_values; i++ {
				read, err := f.Read(value)
				if err == io.EOF || read == 0 {
					return nil, err
				}

				// I think DVID is the payload boundary; this might be a bad assumption
				if label_string == "DVID" {

					// XXX: I think this might skip the first sentence
					return t, nil
				} else if label_string == "GPS5" {
					g := GPS5{}
					g.Parse(value, &s)
					t.Gps = append(t.Gps, g)
				} else if label_string == "GPSU" {
					g := GPSU{}
					err := g.Parse(value)
					if err != nil {
						return nil, err
					}
					t.Time = g
				} else if label_string == "ACCL" {
					a := ACCL{}
					err := a.Parse(value, &s)
					if err != nil {
						return nil, err
					}
					t.Accl = append(t.Accl, a)
				} else if label_string == "TMPC" {
					tmp := TMPC{}
					tmp.Parse(value)
					t.Temp = tmp
				} else if label_string == "TSMP" {
					tsmp := TSMP{}
					tsmp.Parse(value, &s)
				} else if label_string == "GYRO" {
					g := GYRO{}
					err := g.Parse(value, &s)
					if err != nil {
						return nil, err
					}
					t.Gyro = append(t.Gyro, g)
				} else if label_string == "GPSP" {
					g := GPSP{}
					err := g.Parse(value)
					if err != nil {
						return nil, err
					}
					t.GpsAccuracy = g
				} else if label_string == "GPSF" {
					g := GPSF{}
					err := g.Parse(value)
					if err != nil {
						return nil, err
					}
					t.GpsFix = g
				}
			}
		}

		// pack into 4 bytes
		mod := length % 4
		if mod != 0 {
			seek := 4 - mod
			io.CopyN(io.Discard, f, seek)
		}
	}

	return nil, nil
}
