//
// functions related to parsing gerber files
// Apertures support
package apertures

import (
	"blockapertures"
	"errors"
	. "gerberbasetypes"
	"strconv"
	"strings"
	. "xy"
	//	"amprocessor"
)




type Aperture struct {
	Code         int
	SourceString string
	Type         GerberApType
	XSize        float64
	YSize        float64
	Diameter     float64
	HoleDiameter float64
	Vertices     int
	RotAngle     float64
	BlockPtr     *blockapertures.BlockAperture
//	MacroPtr     *amprocessor.ApertureMacro
}


func (apert *Aperture) GetCode() int {
	return apert.Code
}

func (apert *Aperture) Init(sourceString string, fs *FormatSpec) error {
	sourceString = strings.TrimSpace(sourceString)
	var err error = nil
	apert.SourceString = sourceString

	apertureCodePosition := strings.IndexAny(sourceString, "CROP")
	apert.Code, err = strconv.Atoi(sourceString[:apertureCodePosition])
	if err == nil {
		var tmpVal float64
		tmpSplitted := strings.Split(sourceString[apertureCodePosition+2:], "X")
		for j := range tmpSplitted {
			tmpSplitted[j] = strings.TrimSpace(tmpSplitted[j])
		}
		switch sourceString[apertureCodePosition] {
		case 'C':
			apert.Type = AptypeCircle
			if len(tmpSplitted) == 1 || len(tmpSplitted) == 2 {
				for i, s := range tmpSplitted {
					tmpVal, err = strconv.ParseFloat(s, 32)
					if err == nil {
						switch i {
						case 0:
							apert.Diameter = float64(tmpVal)
						case 1:
							apert.HoleDiameter = float64(tmpVal)
						}
					}
				}
			} else {
				err = errors.New("bad number of parameters for circle aperture")
			}
		case 'R':
			apert.Type = AptypeRectangle
			if len(tmpSplitted) == 2 || len(tmpSplitted) == 3 {
				for i, s := range tmpSplitted {
					tmpVal, err = strconv.ParseFloat(s, 32)
					if err == nil {
						switch i {
						case 0:
							apert.XSize = float64(tmpVal)
						case 1:
							apert.YSize = float64(tmpVal)
						case 2:
							apert.HoleDiameter = float64(tmpVal)
						}
					}
				}
			} else {
				err = errors.New("bad number of parameters for rectangle aperture")
			}
		case 'O':
			apert.Type = AptypeObround
			if len(tmpSplitted) == 2 || len(tmpSplitted) == 3 {
				for i, s := range tmpSplitted {
					tmpVal, err = strconv.ParseFloat(s, 32)
					if err == nil {
						switch i {
						case 0:
							apert.XSize = float64(tmpVal)
						case 1:
							apert.YSize = float64(tmpVal)
						case 2:
							apert.HoleDiameter = float64(tmpVal)
						}
					}
				}
			} else {
				err = errors.New("bad number of parameters for obround aperture")
			}
		case 'P':
			apert.Type = AptypePoly
			if len(tmpSplitted) >= 2 && len(tmpSplitted) < 5 {
				for i, s := range tmpSplitted {
					tmpVal, err = strconv.ParseFloat(s, 32)
					if err == nil {
						switch i {
						case 0:
							apert.Diameter = float64(tmpVal) // OuterDiameter
						case 1:
							apert.Vertices = int(tmpVal)
						case 2:
							apert.RotAngle = float64(tmpVal)
						case 3:
							apert.HoleDiameter = float64(tmpVal)
						}
					}
				}
			} else {
				err = errors.New("bad number of parameters for polygon aperture")
			}
		default:
			err = errors.New("aperture macro is not supported")
			goto fExit
		}
	} else {
		err = errors.New("bad aperture number ")
	}

	apert.HoleDiameter *= fs.ReadMU()
	apert.Diameter *= fs.ReadMU()
	apert.YSize *= fs.ReadMU()
	apert.XSize *= fs.ReadMU()

fExit:
	return err
}
