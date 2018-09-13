//Aperture Macros support
package amprocessor

import (
	"errors"
	"fmt"
	"render"
	"strconv"
	"strings"
)

type AMPrimitive interface {
	// takes the state with the FLASH opcode, where aperture code is macro
	// returns the sequence of steps which allow to draw this aperture
	//	Render(int, int, color.RGBA)
	Render(int, int /* *render.Render*/)

	// returns a string representation of thr primitive
	String() string
}

// creates and returns new object
func NewAMPrimitive(amp AMPrimitiveType, modifStrings []interface{}) AMPrimitive {
	switch amp {
	case AMPrimitive_Comment:
		return AMPrimitiveComment{AMPrimitive_Comment, modifStrings}
	case AMPrimitive_Circle:
		return AMPrimitiveCircle{AMPrimitive_Circle, modifStrings}
	case AMPrimitive_VectLine:
		return AMPrimitiveVectLine{AMPrimitive_VectLine, modifStrings}
	case AMPrimitive_CenterLine:
		return AMPrimitiveCenterLine{AMPrimitive_CenterLine, modifStrings}
	case AMPRimitive_OutLine:
		return AMPrimitiveOutLine{AMPRimitive_OutLine, modifStrings}
	case AMPrimitive_Polygon:
		return AMPrimitivePolygon{AMPrimitive_Polygon, modifStrings}
	case AMPrimitive_Moire:
		return AMPrimitiveMoire{AMPrimitive_Moire, modifStrings}
	case AMPrimitive_Thermal:
		return AMPrimitiveThermal{AMPrimitive_Thermal, modifStrings}
	default:
		panic("unknown aperture macro primitive type")
	}

}

type AMPrimitiveType int

func (amp AMPrimitiveType) String() string {
	var retVal string
	switch amp {
	case AMPrimitive_Comment:
		retVal = "comment"
	case AMPrimitive_Circle:
		retVal = "circle"
	case AMPrimitive_VectLine:
		retVal = "vector line"
	case AMPrimitive_CenterLine:
		retVal = "center line"
	case AMPRimitive_OutLine:
		retVal = "outline"
	case AMPrimitive_Polygon:
		retVal = "polygon"
	case AMPrimitive_Moire:
		retVal = "moire"
	case AMPrimitive_Thermal:
		retVal = "thermal"
	default:
		retVal = "unknown"
	}
	return retVal
}

const (
	AMPrimitive_Comment    AMPrimitiveType = 0
	AMPrimitive_Circle     AMPrimitiveType = 1
	AMPrimitive_VectLine   AMPrimitiveType = 20
	AMPrimitive_CenterLine AMPrimitiveType = 21
	AMPRimitive_OutLine    AMPrimitiveType = 4
	AMPrimitive_Polygon    AMPrimitiveType = 5
	AMPrimitive_Moire      AMPrimitiveType = 6
	AMPrimitive_Thermal    AMPrimitiveType = 7
)

type AMPrimitiveComment struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []interface{}
}

func (amp AMPrimitiveComment) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	//subStr1 := "\t\tmodifier("
	//subStr2 := ")="
	//subStr3 := "\n"
	//for i, s := range amp.AMModifiers {
	//	SubStr := subStr1 + strconv.Itoa(i) + subStr2 + s + subStr3
	//	retVal = retVal + SubStr
	//}
	return retVal
}

func (amp AMPrimitiveComment) Render(x0, y0 int /*, context *render.Render*/) {
	return
}

// ********************************************* CIRCLE *********************************************************
type AMPrimitiveCircle struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []interface{}
}

func (amp AMPrimitiveCircle) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	retVal = retVal + ArrayInfo(amp.AMModifiers, []string{"Exposure", "Diameter", "Center X", "Center Y", "Rotation"})
	return retVal
}

func (amp AMPrimitiveCircle) Render(x0, y0 int /*, context *render.Render*/) {

	return
}

// ***************************************** VECTOR LINE *****************************************************
type AMPrimitiveVectLine struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []interface{}
}

func (amp AMPrimitiveVectLine) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	// Exposure, Width, Start X, Start Y, End X, End Y, Rotation
	retVal = retVal + ArrayInfo(amp.AMModifiers, []string{"Exposure", "Width", "Start X", "Start Y", "End X", "End Y", "Rotation"})
	return retVal
}

func (amp AMPrimitiveVectLine) Render(x0, y0 int /*, context *render.Render*/) {
	return
}

// ***************************************** CENTER LINE *****************************************************
type AMPrimitiveCenterLine struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []interface{}
}

func (amp AMPrimitiveCenterLine) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	// Exposure, Width, Hight, Center X, Center Y, Rotation
	retVal = retVal + ArrayInfo(amp.AMModifiers, []string{"Exposure", "Width", "Hight", "Center X", "Center Y", "Rotation"})
	return retVal
}

func (amp AMPrimitiveCenterLine) Render(x0, y0 int /*, context *render.Render*/) {
	return
}

// ***************************************** OUTLINE *****************************************************
type AMPrimitiveOutLine struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []interface{}
}

func (amp AMPrimitiveOutLine) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	//Exposure, # vertices, Start X, Start Y, Subsequent points..., Rotation
	retVal = retVal + ArrayInfo(amp.AMModifiers[:4], []string{"Exposure", "# vertices", "Start X", "Start Y"})
	numPairs := (len(amp.AMModifiers) - 5) / 2
	var i int
	for i = 0; i < numPairs; i++ {
		retVal = retVal + ArrayInfo(amp.AMModifiers[4+i*2:6+i*2], []string{"Vertice " + strconv.Itoa(i) + " X", "Vertice " + strconv.Itoa(i) + " Y"})
	}
	retVal = retVal + ArrayInfo(amp.AMModifiers[len(amp.AMModifiers)-1:], []string{"Rotation"})
	return retVal
}

func (amp AMPrimitiveOutLine) Render(x0, y0 int /*, context *render.Render*/) {
	return
}

// ***************************************** POLYGON *****************************************************
type AMPrimitivePolygon struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []interface{}
}

func (amp AMPrimitivePolygon) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	//Exposure, # vertices, Center X, Center Y, Diameter, Rotation
	retVal = retVal + ArrayInfo(amp.AMModifiers, []string{"Exposure", "# vertices", "Center X", "Center Y", "Diameter", "Rotation"})
	return retVal
}

func (amp AMPrimitivePolygon) Render(x0, y0 int /*, context *render.Render*/) {
	return
}

// ***************************************** MOIRE *****************************************************
type AMPrimitiveMoire struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []interface{}
}

func (amp AMPrimitiveMoire) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	//Center X, Center Y, Outer diameter rings, Ring thickness, Gap, Max # rings, Crosshair thickness, Crosshair length, Rotation
	retVal = retVal + ArrayInfo(amp.AMModifiers,
		[]string{"Center X", "Center Y", "Outer diameter rings", "Ring thickness", "Gap", "Max # rings", "Crosshair thickness", "Crosshair length", "Rotation"})
	return retVal
}

func (amp AMPrimitiveMoire) Render(x0, y0 int /*, context *render.Render*/) {
	return
}

// ***************************************** THERMAL *****************************************************
type AMPrimitiveThermal struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []interface{}
}

func (amp AMPrimitiveThermal) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	//Center X, Center Y, Outer diameter, Inner diameter, Gap, Rotation
	retVal = retVal + ArrayInfo(amp.AMModifiers, []string{"Center X", "Center Y", "Outer diameter", "Inner diameter", "Gap", "Rotation"})
	return retVal
}

func (amp AMPrimitiveThermal) Render(x0, y0 int /*, context *render.Render*/) {
	return
}

// ********************************************* AM container *************************************************
type AMVariable struct {
	Name           string
	Value          string
	PrimitiveIndex int
}

func (amv AMVariable) String() string {
	return amv.Name + "=" + amv.Value + " (primitive index=" + strconv.Itoa(amv.PrimitiveIndex) + ")"
}

type ApertureMacro struct {
	Name       string // name from source string
	Comments   []string
	Variables  []AMVariable
	Primitives []AMPrimitive
}

func (am ApertureMacro) String() string {
	retVal := "\nAperture macro name:\t" + am.Name + "\nComments:\n"
	for i := range am.Comments {
		retVal = retVal + "\t\t" + am.Comments[i] + "\n"
	}
	retVal = retVal + "Variables:\n"
	for i := range am.Variables {
		retVal = retVal + "\t\t" + am.Variables[i].String() + "\n"
	}
	retVal = retVal + "Primitives:\n"
	for i := range am.Primitives {
		retVal = retVal + "\t" + am.Primitives[i].String() + "\n"
	}
	return retVal
}

func NewApertureMacro(src string) (*ApertureMacro, error) {
	retVal := new(ApertureMacro)
	splittedStr := strings.Split(src, "*")
	if strings.HasPrefix(splittedStr[0], "%AM") == true {
		retVal.Name = splittedStr[0][3:]
	} else {
		return retVal, errors.New("Aperture macro name not found")
	}

	if strings.HasPrefix(splittedStr[len(splittedStr)-1], "%") == false {
		return retVal, errors.New("Aperture macro trailing % not found")
	}

	for _, s := range splittedStr[1:] {
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, "0 ") {
			retVal.Comments = append(retVal.Comments, s[2:])
			continue
		}

		if strings.HasPrefix(s, "$") {
			eqSignPos := strings.Index(s, "=")
			if eqSignPos == -1 {
				return retVal, errors.New("Problem with variable: " + s)
			}
			prIndex := len(retVal.Primitives)
			retVal.Variables = append(retVal.Variables, AMVariable{s[1:eqSignPos], s[eqSignPos+1:], prIndex})

			continue
		}

		if len(s) > 2 {
			commaPos := strings.Index(s, ",")
			if commaPos > 2 {
				return retVal, errors.New("Bad aperture macro primitive: " + s)
			}
			primTypeI, err := strconv.Atoi(s[:commaPos])
			if err != nil {
				return retVal, err
			}
			var primType AMPrimitiveType
			primType = AMPrimitiveType(primTypeI)
			modifiersArr := strings.Split(s[commaPos+1:], ",")
			modifInterfaceArr := make([]interface{}, len(modifiersArr))
			for i := range modifiersArr {
				modifInterfaceArr[i] = strings.TrimSpace(modifiersArr[i])
			}
			retVal.Primitives = append(retVal.Primitives, NewAMPrimitive(primType, modifInterfaceArr))
		}
	}
	return retVal, nil
}

func (am ApertureMacro) Render(x0, y0 int, context *render.Render) {

	for i := range am.Primitives {
		am.Primitives[i].Render(x0, y0 /*, context*/)
	}

	return
}

/*
	auxiliary functions
*/

func ArrayInfo(inArray []interface{}, itemNames []string) string {

	// each step constructs the sub-string
	// \t%itemname% = %itemValue%\n
	retVal := ""

	var limIn int = len(inArray)
	var limIt int = len(itemNames)
	var i int = 0

	for i < limIn || i < limIt {
		subStr1 := "\t"
		if i < limIt {
			subStr1 = subStr1 + itemNames[i]
		} else {
			subStr1 = subStr1 + "<unnamed>"
		}

		subStr2 := " = "
		if i < limIn {
			switch v := inArray[i].(type) {
			case string:
				subStr2 = subStr2 + v + "\n"
			case fmt.Stringer:
				subStr2 = subStr2 + v.String() + "\n"
			default:
				subStr2 = subStr2 + fmt.Sprintf("%v", v) + "\n"
			}
		} else {
			subStr2 = subStr2 + "<empty>\n"
		}
		retVal = retVal + subStr1 + subStr2
		i++
	}
	return retVal
}
