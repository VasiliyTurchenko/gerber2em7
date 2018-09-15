//
// functions related to parsing gerber files
// Apertures support
package render

import (
	"fmt"
	. "gerberbasetypes"
	"strconv"

	//	. "xy"
	//	"amprocessor"
)


type BlockAperture struct {
	StartStringNum int
	Code           int
	BodyStrings    []string
	StepsPtr       []*State
}

//Print info
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
	MacroPtr     *ApertureMacro
}


func (apert *Aperture) GetCode() int {
	return apert.Code
}

func (apert *Aperture) String() string {
	retVal := "Aperture:\t" + strconv.Itoa(apert.Code) + "\t"
	retVal = retVal + apert.SourceString + "\n"
	if apert.Type != AptypeMacro {
		names := []string{"Type", "XSize", "YSize", "Diameter", "HoleDiameter", "#Vertices", "Rot.angle"}
		values := []interface{}{apert.Type, apert.XSize, apert.YSize, apert.Diameter, apert.HoleDiameter,
			apert.Vertices, apert.RotAngle}
		retVal = retVal + ArrayInfo(values, names)
	}
	if apert.BlockPtr != nil {
		retVal = retVal + ArrayInfo([]interface{}{apert.BlockPtr}, []string{"BlockAperture"})
	} else {
		retVal = retVal + ArrayInfo([]interface{}{"<empty>"}, []string{"BlockAperture"})
	}
	if apert.MacroPtr != nil {
		retVal = retVal + ArrayInfo([]interface{}{apert.MacroPtr}, []string{"Macro"})
	} else {
		retVal = retVal + ArrayInfo([]interface{}{"<empty>"}, []string{"Macro"})
	}
	return retVal
}

func (apert *Aperture) Render(xC int, yC int, render *Render) {
	if apert.Type == AptypeMacro {
		apert.MacroPtr.Render(xC, yC, render)
	} else {

	}
}

// %ADD10C,0.0650*%
//     ^--------^
//func (apert *Aperture) Init(sourceString string, scale float64) error {
//	sourceString = strings.TrimSpace(sourceString)
//	var err error = nil
//	apert.SourceString = sourceString
//
//	apertureCodePosition := strings.IndexAny(sourceString, "CROP")
//	apert.Code, err = strconv.Atoi(sourceString[:apertureCodePosition])
//	if err == nil {
//		var tmpVal float64
//		tmpSplitted := strings.Split(sourceString[apertureCodePosition+2:], "X")
//		for j := range tmpSplitted {
//			tmpSplitted[j] = strings.TrimSpace(tmpSplitted[j])
//		}
//		switch sourceString[apertureCodePosition] {
//		case 'C':
//			apert.Type = AptypeCircle
//			if len(tmpSplitted) == 1 || len(tmpSplitted) == 2 {
//				for i, s := range tmpSplitted {
//					tmpVal, err = strconv.ParseFloat(s, 64)
//					if err == nil {
//						switch i {
//						case 0:
//							apert.Diameter = float64(tmpVal)
//						case 1:
//							apert.HoleDiameter = float64(tmpVal)
//						}
//					}
//				}
//			} else {
//				err = errors.New("bad number of parameters for circle aperture")
//			}
//		case 'R':
//			apert.Type = AptypeRectangle
//			if len(tmpSplitted) == 2 || len(tmpSplitted) == 3 {
//				for i, s := range tmpSplitted {
//					tmpVal, err = strconv.ParseFloat(s, 64)
//					if err == nil {
//						switch i {
//						case 0:
//							apert.XSize = float64(tmpVal)
//						case 1:
//							apert.YSize = float64(tmpVal)
//						case 2:
//							apert.HoleDiameter = float64(tmpVal)
//						}
//					}
//				}
//			} else {
//				err = errors.New("bad number of parameters for rectangle aperture")
//			}
//		case 'O':
//			apert.Type = AptypeObround
//			if len(tmpSplitted) == 2 || len(tmpSplitted) == 3 {
//				for i, s := range tmpSplitted {
//					tmpVal, err = strconv.ParseFloat(s, 64)
//					if err == nil {
//						switch i {
//						case 0:
//							apert.XSize = float64(tmpVal)
//						case 1:
//							apert.YSize = float64(tmpVal)
//						case 2:
//							apert.HoleDiameter = float64(tmpVal)
//						}
//					}
//				}
//			} else {
//				err = errors.New("bad number of parameters for obround aperture")
//			}
//		case 'P':
//			apert.Type = AptypePoly
//			if len(tmpSplitted) >= 2 && len(tmpSplitted) < 5 {
//				for i, s := range tmpSplitted {
//					tmpVal, err = strconv.ParseFloat(s, 64)
//					if err == nil {
//						switch i {
//						case 0:
//							apert.Diameter = float64(tmpVal) // OuterDiameter
//						case 1:
//							apert.Vertices = int(tmpVal)
//						case 2:
//							apert.RotAngle = float64(tmpVal)
//						case 3:
//							apert.HoleDiameter = float64(tmpVal)
//						}
//					}
//				}
//			} else {
//				err = errors.New("bad number of parameters for polygon aperture")
//			}
//		default:
//			err = errors.New("aperture macro is not supported")
//			goto fExit
//		}
//	} else {
//		err = errors.New("bad aperture number ")
//	}
//
//	apert.HoleDiameter *= scale
//	apert.Diameter *= scale
//	apert.YSize *= scale
//	apert.XSize *= scale
//
//fExit:
//	return err
//}
//
