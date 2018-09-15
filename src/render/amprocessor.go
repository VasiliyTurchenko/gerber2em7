//Aperture Macros support
package render

import (
	"errors"
	"fmt"
	. "gerberbasetypes"
	"strconv"
	"strings"
	stor "strings_storage"
)

// aperture macro dictionary
var AMacroDict []*ApertureMacro

type AMPrimitive interface {
	// takes the state with the FLASH opcode, where aperture code is macro
	// returns the sequence of steps which allow to draw this aperture
	//	Render(int, int, color.RGBA)
	Render(int, int, *Render)

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
	return retVal
}

func (amp AMPrimitiveComment) Render(x0, y0 int, context *Render) {
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

func (amp AMPrimitiveCircle) Render(x0, y0 int, context *Render) {

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

func (amp AMPrimitiveVectLine) Render(x0, y0 int, context *Render) {
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

func (amp AMPrimitiveCenterLine) Render(x0, y0 int, context *Render) {
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

func (amp AMPrimitiveOutLine) Render(x0, y0 int, context *Render) {
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

func (amp AMPrimitivePolygon) Render(x0, y0 int, context *Render) {
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

func (amp AMPrimitiveMoire) Render(x0, y0 int, context *Render) {
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

func (amp AMPrimitiveThermal) Render(x0, y0 int, context *Render) {
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
		return retVal, errors.New("aperture macro name not found")
	}

	if strings.HasPrefix(splittedStr[len(splittedStr)-1], "%") == false {
		return retVal, errors.New("aperture macro trailing % not found")
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
				return retVal, errors.New("problem with variable: " + s)
			}
			prIndex := len(retVal.Primitives)
			retVal.Variables = append(retVal.Variables, AMVariable{s[1:eqSignPos], s[eqSignPos+1:], prIndex})

			continue
		}

		if len(s) > 2 {
			commaPos := strings.Index(s, ",")
			if commaPos > 2 {
				return retVal, errors.New("bad aperture macro primitive: " + s)
			}
			primTypeI, err := strconv.Atoi(s[:commaPos])
			if err != nil {
				return retVal, err
			}
			var primType AMPrimitiveType
			// an odd primitive type fix:
			if primTypeI == 2 {
				primTypeI = 20
			}
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

func (am ApertureMacro) Render(x0, y0 int, context *Render) {

	for i := range am.Primitives {
		am.Primitives[i].Render(x0, y0, context)
	}

	return
}

/*
	auxiliary functions
*/

func ArrayInfo(inArray []interface{}, itemNames []string) string {

	// each step constructs the sub-string
	// \t%itemName% = %itemValue%\n
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

/* Aperture macro definitions extractor */
func ExtractAMDefinitions(inStrings *stor.Storage) ([]*ApertureMacro, *stor.Storage) {
	aMacroDict := make([]*ApertureMacro, 0)
	retStorage := stor.NewStorage()
	apMacroString := ""
	inStrings.ResetPos()
	for {
		gerberString := inStrings.String()
		if len(gerberString) == 0 {
			break
		}
		/*------------------- aperture macro processing start ---------------- */
		if strings.HasPrefix(gerberString, GerberApertureMacroDef) &&
			strings.HasSuffix(gerberString, "%") {
			apMacroString = gerberString
			var err error
			apMacroPtr, err := NewApertureMacro(apMacroString)
			if err != nil {
				checkError(err, 587)
			}
			aMacroDict = append(aMacroDict, apMacroPtr) // store correct aperture
			apMacroString = ""
			continue
		}
		// all unprocessed above goes here
		retStorage.Accept(gerberString)
	}
	return aMacroDict, retStorage
}

// Instantiates an aperture using definition and parameters
//
// %ADD11CIRCLE,.5*%
//     ^---------^
//func NewApertureInstance(code int, name string, def string, scale float64) *Aperture {
func NewApertureInstance(gerberString string, scale float64) *Aperture {

	apString := gerberString[4 : len(gerberString)-2]
	var i int
	for i = 0; i < len(apString); i++ {
		if apString[i] < '0' || apString[i] > '9' {
			break
		}
	}
	commaPos := strings.Index(apString[i:], ",")
	if commaPos == -1 {
		commaPos = len(apString[i:])
	}
	code, _ := strconv.Atoi(apString[:i])
	name := apString[i : commaPos+i]
	def := apString[commaPos+i:]
	retVal := new(Aperture)
	if len(name) == 0 {
		panic("bad aperture " + strconv.Itoa(code) + " name")
	}
	if len(name) == 1 && (name[0] == 'C' || name[0] == 'R' || name[0] == 'O' || name[0] == 'P') {
		// it's ordinary aperture
		err := retVal.Init2(code, name, def, scale)
		if err != nil {
			panic(err)
		}

	} else { // it's macro aperture
		retVal.SourceString = def
		retVal.Type = AptypeMacro
		retVal.Code = code
		// find in macro definitions dictionary for the name

		var instance *ApertureMacro
		for i := range AMacroDict {
			if strings.Compare(AMacroDict[i].Name, name) == 0 {
				instance = new(ApertureMacro)
				*instance = *AMacroDict[i]
				break
			}
		}
		if instance == nil {
			panic("unable to instantiate aperture macro " + strconv.Itoa(code) + name)
		}

		retVal.MacroPtr = instance
	}
	return retVal
}

// %ADD10C,0.0650*%
//     ^--------^

func (apert *Aperture) Init2(code int, name string, def string, scale float64) error {

	var err error = nil
	// for backward compatibility
	apert.SourceString = strconv.Itoa(code) + name + def

	apert.Code = code

	var tmpVal float64
	tmpSplitted := strings.Split(def[1:], "X")
	for j := range tmpSplitted {
		tmpSplitted[j] = strings.TrimSpace(tmpSplitted[j])
	}
	switch name[0] {
	case 'C':
		apert.Type = AptypeCircle
		if len(tmpSplitted) == 1 || len(tmpSplitted) == 2 {
			for i, s := range tmpSplitted {
				tmpVal, err = strconv.ParseFloat(s, 64)
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
				tmpVal, err = strconv.ParseFloat(s, 64)
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
				tmpVal, err = strconv.ParseFloat(s, 64)
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
				tmpVal, err = strconv.ParseFloat(s, 64)
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
		err = errors.New("bad aperture name " + name)
	}

	apert.HoleDiameter *= scale
	apert.Diameter *= scale
	apert.YSize *= scale
	apert.XSize *= scale

	return err
}
