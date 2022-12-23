package telemetry

import (
	"io"
	"io/ioutil"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Read(f io.Reader) (*TELEM, error) {

	// Labels come from https://exiftool.org/TagNames/GoPro.html
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

	label := make([]byte, 4, 4) // 4 byte ascii label of data
	desc := make([]byte, 4, 4)  // 4 byte description of length of data

	// keep a copy of the scale to apply to subsequent sentences
	s := SCAL{}

	// the full telemetry for this period
	t := &TELEM{}

	for {
		// pick out the label
		read, err := f.Read(label)
		if err == io.EOF || read == 0 {
			return nil, err
		}

		label_string := string(label)

		if !stringInSlice(label_string, labels) {
			continue
		}

		// pick out the label description
		read, err = f.Read(desc)
		if err == io.EOF || read == 0 {
			break
		}

		// first byte is zero, there is no length
		if 0x0 == desc[0] {
			continue
		}

		// skip empty packets
		if "EMPT" == label_string {
			io.CopyN(ioutil.Discard, f, 4)
			continue
		}

		// extract the size and length
		val_size := int64(desc[1])
		num_values := (int64(desc[2]) << 8) | int64(desc[3])
		length := val_size * num_values

		// uncomment to see label, type, size and length
		//fmt.Printf("%s (%c) of size %v and len %v\n", label, desc[0], val_size, length)

		if "SCAL" == label_string {
			value := make([]byte, val_size*num_values, val_size*num_values)
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
				if "DVID" == label_string {

					// XXX: I think this might skip the first sentence
					return t, nil
				} else if "GPS5" == label_string {
					g := GPS5{}
					g.Parse(value, &s)
					t.Gps = append(t.Gps, g)
				} else if "GPSU" == label_string {
					g := GPSU{}
					err := g.Parse(value)
					if err != nil {
						return nil, err
					}
					t.Time = g
				} else if "ACCL" == label_string {
					a := ACCL{}
					err := a.Parse(value, &s)
					if err != nil {
						return nil, err
					}
					t.Accl = append(t.Accl, a)
				} else if "TMPC" == label_string {
					tmp := TMPC{}
					tmp.Parse(value)
					t.Temp = tmp
				} else if "TSMP" == label_string {
					tsmp := TSMP{}
					tsmp.Parse(value, &s)
				} else if "GYRO" == label_string {
					g := GYRO{}
					err := g.Parse(value, &s)
					if err != nil {
						return nil, err
					}
					t.Gyro = append(t.Gyro, g)
				} else if "GPSP" == label_string {
					g := GPSP{}
					err := g.Parse(value)
					if err != nil {
						return nil, err
					}
					t.GpsAccuracy = g
				} else if "GPSF" == label_string {
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
			io.CopyN(ioutil.Discard, f, seek)
		}
	}

	return nil, nil
}
