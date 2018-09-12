//
// functions related to parsing gerber files
// Apertures support
package gerbparser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Apertures
const GerberApertureDef = "%ADD"
const GerberApertureMacroDef = "%AM"
const GerberApertureBlockDef = "%AB"
const GerberApertureBlockDefEnd = "%AB*%"

type GerberApType int

const (
	AptypeCircle GerberApType = iota + 1
	AptypeRectangle
	AptypeObround
	AptypePoly
	AptypeMacro
	AptypeBlock
)

func (ga GerberApType) String() string {
	switch ga {
	case AptypeCircle:
		return "circle aperture"
	case AptypeRectangle:
		return "rectangle aperture"
	case AptypeObround:
		return "obround (box) aperture"
	case AptypePoly:
		return "polygon aperture"
	case AptypeMacro:
		return "macro aperture"
	case AptypeBlock:
		return "block aperture"
	default:
	}
	return "Unknown aperture type"

}

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
	BlockPtr     *BlockAperture
	MacroPtr     *MacroAperture
}

type BlockAperture struct {
	StartStringNum int
	Code           int
	BodyStrings    []string
	StepsPtr       []*State
}

func (ba *BlockAperture) Print() {
	fmt.Println("\n***** Block aperture *****")
	fmt.Println("\tBlock aperture code:", ba.Code)
	fmt.Println("\tSource strings:")
	for b := range ba.BodyStrings {
		fmt.Println("\t\t", b, "  ", ba.BodyStrings[b])
	}
	fmt.Println("\tResulting steps:")
	for b := range ba.StepsPtr {
		ba.StepsPtr[b].Print()
	}
}

type MacroAperture struct {
	dummy int
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
