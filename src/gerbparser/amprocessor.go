//Aperture Macros support
package gerbparser

import (
	"errors"
	"strconv"
	"strings"
)

type AMPrimitive interface {
	// takes the state with the FLASH opcode, where aperture code is macro
	// returns the sequence of steps which allow to draw this aperture
	Render (*State) *[]*State

	// returns a string representation of thr primitive
	String() string
}

// creates and returns new object
func NewAMPrimitive(amp AMPrimitiveType, modifStrings []string) AMPrimitive {
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

//type AMPrimitive struct {
//	PrimitiveType AMPrimitiveType
//	AMModifiers   []string
//}

type AMPrimitiveComment struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []string
}

func (amp AMPrimitiveComment) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	subStr1 := "\t\tmodifier("
	subStr2 := ")="
	subStr3 := "\n"
	for i, s := range amp.AMModifiers {
		SubStr := subStr1 + strconv.Itoa(i) + subStr2 + s + subStr3
		retVal = retVal + SubStr
	}
	return retVal
}

func (amp AMPrimitiveComment) Render (step *State) *[]*State {
	return &[]*State{step}
}


type AMPrimitiveCircle struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []string
}

func (amp AMPrimitiveCircle) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	subStr1 := "\t\tmodifier("
	subStr2 := ")="
	subStr3 := "\n"
	for i, s := range amp.AMModifiers {
		SubStr := subStr1 + strconv.Itoa(i) + subStr2 + s + subStr3
		retVal = retVal + SubStr
	}
	return retVal
}

func (amp AMPrimitiveCircle) Render (step *State) *[]*State {
	return &[]*State{step}
}


type AMPrimitiveVectLine struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []string
}

func (amp AMPrimitiveVectLine) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	subStr1 := "\t\tmodifier("
	subStr2 := ")="
	subStr3 := "\n"
	for i, s := range amp.AMModifiers {
		SubStr := subStr1 + strconv.Itoa(i) + subStr2 + s + subStr3
		retVal = retVal + SubStr
	}
	return retVal
}

func (amp AMPrimitiveVectLine) Render (step *State) *[]*State {
	return &[]*State{step}
}


type AMPrimitiveCenterLine struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []string
}

func (amp AMPrimitiveCenterLine) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	subStr1 := "\t\tmodifier("
	subStr2 := ")="
	subStr3 := "\n"
	for i, s := range amp.AMModifiers {
		SubStr := subStr1 + strconv.Itoa(i) + subStr2 + s + subStr3
		retVal = retVal + SubStr
	}
	return retVal
}

func (amp AMPrimitiveCenterLine) Render (step *State) *[]*State {
	return &[]*State{step}
}

type AMPrimitiveOutLine struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []string
}

func (amp AMPrimitiveOutLine) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	subStr1 := "\t\tmodifier("
	subStr2 := ")="
	subStr3 := "\n"
	for i, s := range amp.AMModifiers {
		SubStr := subStr1 + strconv.Itoa(i) + subStr2 + s + subStr3
		retVal = retVal + SubStr
	}
	return retVal
}

func (amp AMPrimitiveOutLine) Render (step *State) *[]*State {
	return &[]*State{step}
}


type AMPrimitivePolygon struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []string
}

func (amp AMPrimitivePolygon) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	subStr1 := "\t\tmodifier("
	subStr2 := ")="
	subStr3 := "\n"
	for i, s := range amp.AMModifiers {
		SubStr := subStr1 + strconv.Itoa(i) + subStr2 + s + subStr3
		retVal = retVal + SubStr
	}
	return retVal
}

func (amp AMPrimitivePolygon) Render (step *State) *[]*State {
	return &[]*State{step}
}


type AMPrimitiveMoire struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []string
}

func (amp AMPrimitiveMoire) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	subStr1 := "\t\tmodifier("
	subStr2 := ")="
	subStr3 := "\n"
	for i, s := range amp.AMModifiers {
		SubStr := subStr1 + strconv.Itoa(i) + subStr2 + s + subStr3
		retVal = retVal + SubStr
	}
	return retVal
}

func (amp AMPrimitiveMoire) Render (step *State) *[]*State {
	return &[]*State{step}
}


type AMPrimitiveThermal struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []string
}

func (amp AMPrimitiveThermal) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	subStr1 := "\t\tmodifier("
	subStr2 := ")="
	subStr3 := "\n"
	for i, s := range amp.AMModifiers {
		SubStr := subStr1 + strconv.Itoa(i) + subStr2 + s + subStr3
		retVal = retVal + SubStr
	}
	return retVal
}

func (amp AMPrimitiveThermal) Render (step *State) *[]*State {
	return &[]*State{step}
}


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
		retVal = retVal + "\t\t" + am.Primitives[i].String() + "\n"
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
			retVal.Comments = append(retVal.Comments, s[3:])
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
			for i := range modifiersArr {
				modifiersArr[i] = strings.TrimSpace(modifiersArr[i])
			}
			retVal.Primitives = append(retVal.Primitives, NewAMPrimitive(primType, modifiersArr))
		}
	}

	return retVal, nil
}
