//
// functions related to parsing gerber files
// Apertures support
package render

import (
	"errors"
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
			w := transformCoord(apert.XSize, render.XRes)
			h := transformCoord(apert.YSize, render.YRes)
			d := transformCoord(apert.Diameter, render.XRes)
			hd := transformCoord(apert.HoleDiameter, render.XRes)
			switch apert.Type {
			case AptypeRectangle:
				render.DrawFilledRectangle(xC, yC, w, h, render.ApColor)
			case AptypeCircle:
				render.DrawDonut(xC, yC, d, hd, render.ApColor)
			case AptypeObround:
				if w == h {
					render.DrawDonut(xC, yC, w, hd, render.ApColor)
				} else {
					render.DrawObRound(xC, yC, w, h, 0, render.ObRoundColor)
				}
			case AptypePoly:
				render.DrawDonut(xC, yC, d, hd, render.MissedColor)
				fmt.Println("Polygonal apertures ain't supported.")
			default:
				checkError(errors.New("bad aperture type found"), 5011)
				break
			}
	}
}

func (apert *Aperture) Draw(x0, y0 int, x1, y1 int,  context *Render) {

}
