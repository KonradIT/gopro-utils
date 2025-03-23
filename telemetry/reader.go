package telemetry

import (
	"io"
	"slices"
)

type Parser interface {
	Parse(bytes []byte) error
}

var (
	_ Parser = &GPSU{}
	_ Parser = &GPSP{}
	_ Parser = &GPSF{}
	_ Parser = &TMPC{}
)

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

		labelStr := string(label)

		if !slices.Contains(labels, labelStr) {
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
		if labelStr == "EMPT" {
			_, err := io.CopyN(io.Discard, f, 4)
			if err != nil {
				return nil, err
			}

			continue
		}

		// extract the size and length
		valSize := int64(desc[1])
		numValues := (int64(desc[2]) << 8) | int64(desc[3])
		length := valSize * numValues

		// uncomment to see label, type, size and length
		// fmt.Printf("%s (%c) of size %v and len %v\n", label, desc[0], val_size, length)

		if labelStr == "SCAL" {
			value := make([]byte, valSize*numValues)

			read, err = f.Read(value)
			if err == io.EOF || read == 0 {
				return nil, err
			}

			// clear the scales
			s.Values = s.Values[:0]

			err := s.Parse(value, valSize)
			if err != nil {
				return nil, err
			}
		} else {
			value := make([]byte, valSize)

			for range numValues {
				read, err := f.Read(value)
				if err == io.EOF || read == 0 {
					return nil, err
				}

				// I think DVID is the payload boundary; this might be a bad assumption
				switch labelStr {
				case "DVID":
					// XXX: I think this might skip the first sentence
					return t, nil

				case "GPS5":
					g := GPS5{}
					err := g.Parse(value, &s)
					if err != nil {
						return nil, err
					}
					t.Gps = append(t.Gps, g)

				case "GPSU":
					g := GPSU{}
					err := g.Parse(value)
					if err != nil {
						return nil, err
					}
					t.Time = g

				case "ACCL":
					a := ACCL{}
					err := a.Parse(value, &s)
					if err != nil {
						return nil, err
					}
					t.Accl = append(t.Accl, a)

				case "TMPC":
					tmp := TMPC{}
					err := tmp.Parse(value)
					if err != nil {
						return nil, err
					}
					t.Temp = tmp

				case "TSMP":
					tsmp := TSMP{}
					tsmp.Parse(value)

				case "GYRO":
					g := GYRO{}
					err := g.Parse(value, &s)
					if err != nil {
						return nil, err
					}
					t.Gyro = append(t.Gyro, g)

				case "GPSP":
					g := GPSP{}
					err := g.Parse(value)
					if err != nil {
						return nil, err
					}
					t.GpsAccuracy = g

				case "GPSF":
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
			_, err := io.CopyN(io.Discard, f, seek)
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}
