//Aperture Macros support
package render

import (
	"errors"
	"fmt"
	. "gerberbasetypes"
	"math"
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

	Draw(int, int, int, int, *Render)

	// returns a string representation of thr primitive
	String() string

	// instantiates an macro primitive using parameters, scale factor and macro variables
	Init(float64)
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
		return AMPrimitiveThermal{amp, modifStrings}
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

func (amp AMPrimitiveComment) Draw(x0, y0 int, x1, y1 int, context *Render) {
	return
}

func (amp AMPrimitiveComment) Init(scale float64) {

}

// ********************************************* CIRCLE *********************************************************
// ++ add donut functionality
type AMPrimitiveCircle struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []interface{}
}

func (amp AMPrimitiveCircle) String() string {
	retVal := "Aperture macro primitive:\t"
	retVal = retVal + amp.PrimitiveType.String() + "\n"
	retVal = retVal + ArrayInfo(amp.AMModifiers, []string{"Exposure", "Diameter", "Center X", "Center Y", "Rotation", "Hole diameter"})
	return retVal
}

func (amp AMPrimitiveCircle) Render(x0, y0 int, context *Render) {
	// coordinates of the circle center after rotation
	xd, yd, _ := RotatePoint(amp.AMModifiers[2].(float64), amp.AMModifiers[3].(float64), amp.AMModifiers[4].(float64))
	xC := x0 + transformCoord(xd, context.XRes)
	yC := y0 + transformCoord(yd, context.YRes)
	d := transformCoord(amp.AMModifiers[1].(float64), context.XRes)
	hd := transformCoord(amp.AMModifiers[5].(float64), context.XRes)
	context.MovePen(x0,y0,xC,yC, context.MovePenColor)
	context.DrawDonut(xC, yC, d, hd, context.ApColor)
	// go back
	context.MovePen(xC,yC,x0,y0, context.MovePenColor)
	return
}

func (amp AMPrimitiveCircle) Draw(x0, y0 int, x1, y1 int, context *Render) {

	return
}
// Instantiates circle primitive
/*
0	Exposure off/on (0/1)
1	Diameter. A decimal ≥ 0
2	Center X coordinate. A decimal.
3	Center Y coordinate. A decimal.
4	Rotation angle of the center, in degrees counterclockwise. A decimal.
	The primitive is rotated around the origin of the macro definition, i.e. the
	(0, 0) point of macro coordinates.
	The rotation modifier is optional. The default is no rotation.
5	Hole diameter (optional)
 */
func (amp AMPrimitiveCircle) Init(scale float64) {
	if len(amp.AMModifiers) < 4 {
		panic("unable to create aperture macro primitive circle - not enough parameters, have " +
			strconv.Itoa(len(amp.AMModifiers)) + ", need 4 or 5")
	}
	if len(amp.AMModifiers) == 4 {
		amp.AMModifiers = append(amp.AMModifiers, 0.0)
	}
// add hole diameter = 0 if otherwise not specified
	if len(amp.AMModifiers) == 0 {
		amp.AMModifiers = append(amp.AMModifiers, 0.0)
	}

	for i := range amp.AMModifiers {
		amp.AMModifiers[i] = convertToFloat(amp.AMModifiers[i])
		if (i > 0 && i < 4) || i == 5 {
			switch amp.AMModifiers[i].(type) {
			case float64:
				amp.AMModifiers[i] = scale * (amp.AMModifiers[i]).(float64)
			default:
			}
		}
	}
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
	// if rotation = 0 use rectangle algorithm, polygon otherwise
	rot := amp.AMModifiers[6].(float64)
	width := transformCoord(amp.AMModifiers[1].(float64), context.XRes)

	xsr, ysr, _ := RotatePoint(amp.AMModifiers[2].(float64), amp.AMModifiers[3].(float64), rot)
	xStart := transformCoord(xsr, context.XRes)
	yStart := transformCoord(ysr, context.YRes)

	xer, yer, _ := RotatePoint(amp.AMModifiers[4].(float64), amp.AMModifiers[5].(float64), rot)
	xEnd := transformCoord(xer,  context.XRes)
	yEnd := transformCoord(yer, context.YRes)

	var dx, dy, height int
	if xStart == xEnd || yStart == yEnd {
		//render.DrawFilledRectangle(xC, yC, w, h, render.ApColor)
		if xStart == xEnd {
			dx = xStart
			height = yEnd - yStart
			dy = yStart + height/ 2
		}
		if yStart == yEnd {
			width , height = xEnd - xStart, width
			dx = xStart + (width) / 2
			dy = yStart
		}
		if height < 0 {
			height = -height
		}
		if width < 0 {
			width = -width
		}
		context.MovePen(x0, y0, x0 + dx, y0 + dy, context.MovePenColor)
		context.DrawFilledRectangle(x0 + dx, y0 + dy, width, height, context.ApColor)
		context.MovePen(x0 + dx, y0 + dy, x0, y0, context.MovePenColor)
	} else {
		phi, _ := GetAngle(xer - xsr, yer-ysr)
		widthF := amp.AMModifiers[1].(float64)
		verticesX := make([]float64, 0)
		verticesY := make([]float64, 0)
		xd := (widthF / 2) * math.Sin(phi)
		yd := (widthF / 2) * math.Cos(phi)
		verticesX = append(verticesX, xsr + xd)
		verticesY = append(verticesY, ysr - yd)
		verticesX = append(verticesX, xsr - xd)
		verticesY = append(verticesY, ysr + yd)
		verticesX = append(verticesX, xer - xd)
		verticesY = append(verticesY, yer + yd)
		verticesX = append(verticesX, xer + xd)
		verticesY = append(verticesY, yer - yd)
		verticesX = append(verticesX, verticesX[0])
		verticesY = append(verticesY, verticesY[0])

		for i := range verticesX {
			verticesX[i] = float64(x0) + transformFloatCoord(verticesX[i], context.XRes)
			verticesY[i] = float64(y0) + transformFloatCoord(verticesY[i], context.YRes)
		}
		context.MovePen(x0, y0, int(verticesX[0]), int(verticesY[0]), context.MovePenColor)
		context.RenderOutline(&verticesX, &verticesY)
		context.MovePen(int(verticesX[0]), int(verticesY[0]), x0, y0, context.MovePenColor)
	}
	return
}

func (amp AMPrimitiveVectLine) Draw(x0, y0 int, x1, y1 int, context *Render) {

}

func (amp AMPrimitiveVectLine) Init(scale float64) {
	if len(amp.AMModifiers) < 7 {
		panic("unable to create aperture macro primitive vector line - not enough parameters, have " +
			strconv.Itoa(len(amp.AMModifiers)) + ", need 7")
	}
	for i := range amp.AMModifiers {
		amp.AMModifiers[i] = convertToFloat(amp.AMModifiers[i])
		if i > 0 && i < 6 {
			switch amp.AMModifiers[i].(type) {
			case float64:
				amp.AMModifiers[i] = scale * (amp.AMModifiers[i]).(float64)
			default:
			}
		}
	}
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
	// if rotation = 0 use rectangle algorithm, polygon otherwise
	// make VectorLine and use it

	var vLineModifs = []interface{}{
		amp.AMModifiers[0],	// exposure
		amp.AMModifiers[2], // width of centerline goes as height of vectorline
		amp.AMModifiers[3].(float64) - (amp.AMModifiers[1].(float64)) / 2.0 , // start X
		amp.AMModifiers[4].(float64), // start Y
		amp.AMModifiers[3].(float64) + (amp.AMModifiers[1].(float64)) / 2.0 , // // end X
		amp.AMModifiers[4].(float64), // start Y
		amp.AMModifiers[5], // rot
	}

	var vLine = AMPrimitiveVectLine {AMPrimitive_VectLine, vLineModifs}
	vLine.Render(x0, y0, context)
	return
}
func (amp AMPrimitiveCenterLine) Draw(x0, y0 int, x1, y1 int, context *Render) {

}

func (amp AMPrimitiveCenterLine) Init(scale float64) {
	if len(amp.AMModifiers) < 6 {
		panic("unable to create aperture macro primitive center line - not enough parameters, have " +
			strconv.Itoa(len(amp.AMModifiers)) + ", need 6")
	}
	for i := range amp.AMModifiers {
		amp.AMModifiers[i] = convertToFloat(amp.AMModifiers[i])
		if i > 0 && i < 5 {
			switch amp.AMModifiers[i].(type) {
			case float64:
				amp.AMModifiers[i] = scale * (amp.AMModifiers[i]).(float64)
			default:
			}
		}
	}
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

func (amp AMPrimitiveOutLine) Draw(x0, y0 int, x1, y1 int, context *Render) {

}

func (amp AMPrimitiveOutLine) Init(scale float64) {

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
	rot := amp.AMModifiers[5].(float64)
	numVertices := amp.AMModifiers[1].(float64)
	centerX := amp.AMModifiers[2].(float64)
	centerY := amp.AMModifiers[3].(float64)
	dia := amp.AMModifiers[4].(float64)

	verticesX := make([]float64, 0)
	verticesY := make([]float64, 0)
	deltaPhi := rad2Deg((2 * math.Pi) / numVertices)
	for i := 0; i < int(numVertices); i++ {
		phi := float64(i) * deltaPhi
		verticesX = append(verticesX, centerX + (dia / 2) * math.Cos(deg2Rad(phi)))
		verticesY = append(verticesY, centerY + (dia / 2) * math.Sin(deg2Rad(phi)))
	}
	for i := range verticesX {
		verticesX[i], verticesY[i], _ = RotatePoint(verticesX[i], verticesY[i], rot)
		verticesX[i] = float64(x0) + transformFloatCoord(verticesX[i], context.XRes)
		verticesY[i] = float64(y0) + transformFloatCoord(verticesY[i], context.YRes)
	}
	context.MovePen(x0, y0, int(verticesX[0]), int(verticesY[0]), context.MovePenColor)
	context.RenderOutline(&verticesX, &verticesY)
	context.MovePen(int(verticesX[0]), int(verticesY[0]), x0, y0, context.MovePenColor)
}

func (amp AMPrimitivePolygon) Draw(x0, y0 int, x1, y1 int, context *Render) {
}

func (amp AMPrimitivePolygon) Init(scale float64) {
	if len(amp.AMModifiers) < 6 {
		panic("unable to create aperture macro primitive polygon - not enough parameters, have " +
			strconv.Itoa(len(amp.AMModifiers)) + ", need 6")
	}
	for i := range amp.AMModifiers {
		amp.AMModifiers[i] = convertToFloat(amp.AMModifiers[i])
		if i > 1 && i < 5 {
			switch amp.AMModifiers[i].(type) {
			case float64:
				amp.AMModifiers[i] = scale * (amp.AMModifiers[i]).(float64)
			default:
			}
		}
	}

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

func (amp AMPrimitiveMoire) Draw(x0, y0 int, x1, y1 int, context *Render) {

}

func (amp AMPrimitiveMoire) Init(scale float64) {

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

func (amp AMPrimitiveThermal) Draw(x0, y0 int, x1, y1 int, context *Render) {

}

func (amp AMPrimitiveThermal) Init(scale float64) {

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

func (am ApertureMacro) Draw(x0, y0 int, x1, y1 int,  context *Render) {

	for i := range am.Primitives {
		am.Primitives[i].Draw(x0, y0, x1, y1,  context)
	}

	return
}

func (am ApertureMacro) Init(scale float64) {

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

				for i := range (*instance).Primitives {
					(*instance).Primitives[i].Init(scale)
				}

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


func convertToFloat(arg interface{}) float64 {
	panicString1 := "convertToFloat(arg interface{}) float64 - variables not implemented"
	panicString2 := "convertToFloat(arg interface{}) float64 - not supported interface{}"
	switch arg.(type) {
	case float64:
		return arg.(float64)
	case string:
		retVal, err := strconv.ParseFloat(arg.(string), 64)
		if err != nil {
			panic(panicString1)
		}
		return retVal
	case int:
		return float64(arg.(int))
	default:
		panic(panicString2)
	}
	return 0
}

func band(arg, bandVal float64) float64 {
	if arg > bandVal {
		arg = bandVal
	} else if arg < -bandVal {
		arg = -bandVal
	}
	return arg
}

// rotates the point(x0,y0) counter-clockwise by phi degrees
func RotatePoint(x0, y0, phi float64) (x1, y1 float64, phi1 float64) {

	phiRad := deg2Rad(phi)

	phi0, r := GetAngle(x0, y0)
	phi1 = phi0 + phiRad
	x1 = r * math.Cos(phi1)
	y1 = r * math.Sin(phi1)

	return x1, y1, phi1
}

// returns angle between hyp and x side
func GetAngle(x, y float64) (angle float64, hyp float64) {
	hyp = math.Hypot(x, y)
	cosPhi := band(x / hyp, 1.0)
	angle = math.Acos(cosPhi)
	if y < 0 {
		angle = math.Pi *2  - angle
	}
	return angle, hyp
}
