//Aperture Macros support
package render

import (
	"calculator"
	"errors"
	"fmt"
	. "gerberbasetypes"
	"math"
	"strconv"
	"strings"
	stor "strings_storage"
	glog "glog_t"
)

// aperture macro dictionary
var AMacroDict []*ApertureMacro

type AMPrimitive interface {
	// takes the state with the FLASH opcode, where aperture code is macro
	// returns the sequence of steps which allow to draw this aperture
	//	Render(int, int, color.RGBA)
	Render(int, int, *Render)

	// draws a line or an arc using aperture as "brush"
	Draw(int, int, int, int, *Render)

	// returns a string representation of thr primitive
	String() string

	// instantiates an macro primitive using parameters, scale factor and macro variables
	Init(float64, []float64) AMPrimitive

	// returns a copy of object
	Copy() AMPrimitive
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
//		panic("unknown aperture macro primitive type")
	glog.Fatalln("unknown aperture macro primitive type")
		return nil
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
	retVal = retVal + ArrayInfo(amp.AMModifiers, []string{})
	return retVal
}

func (amp AMPrimitiveComment) Render(x0, y0 int, context *Render) {
	return
}

func (amp AMPrimitiveComment) Draw(x0, y0 int, x1, y1 int, context *Render) {
	return
}

func (amp AMPrimitiveComment) Init(scale float64, params []float64) AMPrimitive {

	return NewAMPrimitive(AMPrimitive_Comment, []interface{}{})
}

func (amp AMPrimitiveComment) Copy() AMPrimitive {
	var modifStrings []interface{}
	for i := range amp.AMModifiers {
		modifStrings = append(modifStrings, amp.AMModifiers[i])
	}
	return NewAMPrimitive(AMPrimitive_Comment, modifStrings)
}

// ********************************************* CIRCLE *********************************************************
// ++ add donut functionality
type AMPrimitiveCircle struct {
	PrimitiveType AMPrimitiveType
	AMModifiers   []interface{}
	//cirCX	int
	//cirCY	int
	//cirD	int
	//cirHD	int
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
	//xC := amp.cirCX + x0
	//yC := amp.cirCY + y0
	context.MovePen(x0, y0, xC, yC, context.MovePenColor)
	//	context.DrawDonut(xC, yC, amp.cirD, amp.cirHD, context.ApColor)
	colr := context.ApColor
	if 0.0 == amp.AMModifiers[0].(float64) {
		colr = context.ClearColor
	}
	context.DrawDonut(xC, yC, d, hd, colr)
	// go back
	context.MovePen(xC, yC, x0, y0, context.MovePenColor)
	return
}

func (amp AMPrimitiveCircle) Draw(x0, y0 int, x1, y1 int, context *Render) {
	BadMethod()
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
func (amp AMPrimitiveCircle) Init(scale float64, params []float64) AMPrimitive {
	if len(amp.AMModifiers) < 4 {
		//panic("unable to create aperture macro primitive circle - not enough parameters, have " +
		//	strconv.Itoa(len(amp.AMModifiers)) + ", need 4 or 5")
		glog.Fatalln("unable to create aperture macro primitive circle - not enough parameters, " +
			strconv.Itoa(len(amp.AMModifiers)) + " given, need 4 or 5")
	}
	if len(amp.AMModifiers) == 4 {
		amp.AMModifiers = append(amp.AMModifiers, 0.0)
	}
	// add hole diameter = 0 if otherwise not specified
	if len(amp.AMModifiers) == 5 {
		amp.AMModifiers = append(amp.AMModifiers, 0.0)
	}

	for i := range amp.AMModifiers {
		amp.AMModifiers[i] = convertToFloat(amp.AMModifiers[i], params)
		if (i > 0 && i < 4) || i == 5 {
			//switch amp.AMModifiers[i].(type) {
			//case float64:
			//	amp.AMModifiers[i] = scale * (amp.AMModifiers[i]).(float64)
			//default:
			amp.AMModifiers[i] = scale * (amp.AMModifiers[i]).(float64)
			//}
		}
	}
	//xd, yd, _ := RotatePoint(amp.AMModifiers[2].(float64), amp.AMModifiers[3].(float64), amp.AMModifiers[4].(float64))
	//amp.cirCX = transformCoord(xd, context.XRes)
	//amp.cirCY = transformCoord(yd, context.YRes)
	//amp.cirD = transformCoord(amp.AMModifiers[1].(float64), context.XRes)
	//amp.cirHD = transformCoord(amp.AMModifiers[5].(float64), context.XRes)
	//
	return amp
}

func (amp AMPrimitiveCircle) Copy() AMPrimitive {
	var modifStrings []interface{}
	for i := range amp.AMModifiers {
		modifStrings = append(modifStrings, amp.AMModifiers[i])
	}
	return NewAMPrimitive(AMPrimitive_Circle, modifStrings)
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
	xEnd := transformCoord(xer, context.XRes)
	yEnd := transformCoord(yer, context.YRes)

	var dx, dy, height int
	if xStart == xEnd || yStart == yEnd {
		//render.DrawFilledRectangle(xC, yC, w, h, render.ApColor)
		if xStart == xEnd {
			dx = xStart
			height = yEnd - yStart
			dy = yStart + height/2
		}
		if yStart == yEnd {
			width, height = xEnd-xStart, width
			dx = xStart + (width)/2
			dy = yStart
		}
		if height < 0 {
			height = -height
		}
		if width < 0 {
			width = -width
		}
		context.MovePen(x0, y0, x0+dx, y0+dy, context.MovePenColor)
		colr := context.ApColor
		if amp.AMModifiers[0].(float64) == 0.0 {
			colr = context.ClearColor
		}
		context.DrawFilledRectangle(x0+dx, y0+dy, width, height, colr)
		context.MovePen(x0+dx, y0+dy, x0, y0, context.MovePenColor)
	} else {
		phi, _ := GetAngle(xer-xsr, yer-ysr)
		widthF := amp.AMModifiers[1].(float64)
		verticesX := make([]float64, 0)
		verticesY := make([]float64, 0)
		xd := (widthF / 2) * math.Sin(phi)
		yd := (widthF / 2) * math.Cos(phi)
		verticesX = append(verticesX, xsr+xd)
		verticesY = append(verticesY, ysr-yd)
		verticesX = append(verticesX, xsr-xd)
		verticesY = append(verticesY, ysr+yd)
		verticesX = append(verticesX, xer-xd)
		verticesY = append(verticesY, yer+yd)
		verticesX = append(verticesX, xer+xd)
		verticesY = append(verticesY, yer-yd)
		verticesX = append(verticesX, verticesX[0])
		verticesY = append(verticesY, verticesY[0])

		for i := range verticesX {
			verticesX[i] = float64(x0) + transformFloatCoord(verticesX[i], context.XRes)
			verticesY[i] = float64(y0) + transformFloatCoord(verticesY[i], context.YRes)
		}
		context.MovePen(x0, y0, int(verticesX[0]), int(verticesY[0]), context.MovePenColor)

		colr := context.RegionColor
		if amp.AMModifiers[0].(float64) == 0.0 {
			colr = context.ClearColor
		}

		context.RenderOutline(&verticesX, &verticesY, colr)
		context.MovePen(int(verticesX[0]), int(verticesY[0]), x0, y0, context.MovePenColor)
	}
	return
}

func (amp AMPrimitiveVectLine) Draw(x0, y0 int, x1, y1 int, context *Render) {
	BadMethod()
}

func (amp AMPrimitiveVectLine) Init(scale float64, params []float64) AMPrimitive {
	if len(amp.AMModifiers) < 7 {
		glog.Fatalln("unable to create aperture macro primitive vector line - not enough parameters, " +
			strconv.Itoa(len(amp.AMModifiers)) + " given, need 7")
	}
	for i := range amp.AMModifiers {
		amp.AMModifiers[i] = convertToFloat(amp.AMModifiers[i], params)
		if i > 0 && i < 6 {
			switch amp.AMModifiers[i].(type) {
			case float64:
				amp.AMModifiers[i] = scale * (amp.AMModifiers[i]).(float64)
			default:
			}
		}
	}
	return amp
}

func (amp AMPrimitiveVectLine) Copy() AMPrimitive {
	var modifStrings []interface{}
	for i := range amp.AMModifiers {
		modifStrings = append(modifStrings, amp.AMModifiers[i])
	}
	return NewAMPrimitive(AMPrimitive_VectLine, modifStrings)
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
		amp.AMModifiers[0],                                                // exposure
		amp.AMModifiers[2],                                                // width of centerline goes as height of vectorline
		amp.AMModifiers[3].(float64) - (amp.AMModifiers[1].(float64))/2.0, // start X
		amp.AMModifiers[4].(float64),                                      // start Y
		amp.AMModifiers[3].(float64) + (amp.AMModifiers[1].(float64))/2.0, // // end X
		amp.AMModifiers[4].(float64),                                      // start Y
		amp.AMModifiers[5],                                                // rot
	}

	var vLine = AMPrimitiveVectLine{AMPrimitive_VectLine, vLineModifs}
	vLine.Render(x0, y0, context)
	return
}
func (amp AMPrimitiveCenterLine) Draw(x0, y0 int, x1, y1 int, context *Render) {
	BadMethod()
}

func (amp AMPrimitiveCenterLine) Init(scale float64, params []float64) AMPrimitive {
	if len(amp.AMModifiers) < 6 {
		glog.Fatalln("unable to create aperture macro primitive center line - not enough parameters, " +
			strconv.Itoa(len(amp.AMModifiers)) + " given, need 6")
	}
	for i := range amp.AMModifiers {
		amp.AMModifiers[i] = convertToFloat(amp.AMModifiers[i], params)
		if i > 0 && i < 5 {
			switch amp.AMModifiers[i].(type) {
			case float64:
				amp.AMModifiers[i] = scale * (amp.AMModifiers[i]).(float64)
			default:
			}
		}
	}
	return amp
}

func (amp AMPrimitiveCenterLine) Copy() AMPrimitive {
	var modifStrings []interface{}
	for i := range amp.AMModifiers {
		modifStrings = append(modifStrings, amp.AMModifiers[i])
	}
	return NewAMPrimitive(AMPrimitive_CenterLine, modifStrings)
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
	//	numCoordPairs := int(convertToFloat(amp.AMModifiers[1])) + 1
	numCoordPairs := int(amp.AMModifiers[1].(float64)) + 1
	rot := amp.AMModifiers[len(amp.AMModifiers)-1].(float64)
	verticesX := make([]float64, 0)
	verticesY := make([]float64, 0)
	i := 2
	for i < 2+numCoordPairs*2 {
		verticesX = append(verticesX, amp.AMModifiers[i].(float64))
		verticesY = append(verticesY, amp.AMModifiers[i+1].(float64))
		i += 2
	}
	for i := range verticesX {
		verticesX[i], verticesY[i], _ = RotatePoint(verticesX[i], verticesY[i], rot)
		verticesX[i] = float64(x0) + transformFloatCoord(verticesX[i], context.XRes)
		verticesY[i] = float64(y0) + transformFloatCoord(verticesY[i], context.YRes)
	}
	context.MovePen(x0, y0, int(verticesX[0]), int(verticesY[0]), context.MovePenColor)

	colr := context.RegionColor
	if amp.AMModifiers[0].(float64) == 0.0 {
		colr = context.ClearColor
	}

	context.RenderOutline(&verticesX, &verticesY, colr)
	context.MovePen(int(verticesX[0]), int(verticesY[0]), x0, y0, context.MovePenColor)
	return
}

func (amp AMPrimitiveOutLine) Draw(x0, y0 int, x1, y1 int, context *Render) {
	BadMethod()

}

/*
0		Exposure off/on (0/1)
1		The number of vertices of the outline = the number of coordinate
		pairs minus one.
		An integer ≥3.
2, 3	Start point X and Y coordinates. Decimals.
4, 5	First subsequent X and Y coordinates. Decimals.
		... Further subsequent X and Y coordinates. Decimals.
		The X and Y coordinates are not modal: both X and Y must be
		specified for all points.
4+2n, 5+2n Last subsequent X and Y coordinates. Decimals.
		Must be equal to the start coordinates.
6+2n	Rotation angle, in degrees counterclockwise, a decimal.
		The primitive is rotated around the origin of the macro definition, i.e. the
		(0, 0) point of macro coordinates.
*/
func (amp AMPrimitiveOutLine) Init(scale float64, params []float64) AMPrimitive {
	numCoordPairs := int(convertToFloat(amp.AMModifiers[1], params))
	if numCoordPairs < 3 {
		glog.Fatalln("unable to create aperture macro primitive outline - not enough coordinate pairs, " +
			strconv.Itoa(numCoordPairs) + " given, need at least 3")
	}
	correctLength := 2 + numCoordPairs*2 + 1
	if len(amp.AMModifiers) < correctLength {
		glog.Fatalln("unable to create aperture macro primitive outline - not enough parameters, " +
			strconv.Itoa(len(amp.AMModifiers)) + " given, need " + strconv.Itoa(correctLength))
	}
	numCoordPairs++

	for i := range amp.AMModifiers {
		amp.AMModifiers[i] = convertToFloat(amp.AMModifiers[i], params)
		if i > 2 && i < len(amp.AMModifiers)-2 {
			switch amp.AMModifiers[i].(type) {
			case float64:
				amp.AMModifiers[i] = scale * (amp.AMModifiers[i]).(float64)
			default:
			}
		}
	}
	return amp
}

func (amp AMPrimitiveOutLine) Copy() AMPrimitive {
	var modifStrings []interface{}
	for i := range amp.AMModifiers {
		modifStrings = append(modifStrings, amp.AMModifiers[i])
	}
	return NewAMPrimitive(AMPRimitive_OutLine, modifStrings)
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
		verticesX = append(verticesX, centerX+(dia/2)*math.Cos(deg2Rad(phi)))
		verticesY = append(verticesY, centerY+(dia/2)*math.Sin(deg2Rad(phi)))
	}
	for i := range verticesX {
		verticesX[i], verticesY[i], _ = RotatePoint(verticesX[i], verticesY[i], rot)
		verticesX[i] = float64(x0) + transformFloatCoord(verticesX[i], context.XRes)
		verticesY[i] = float64(y0) + transformFloatCoord(verticesY[i], context.YRes)
	}
	context.MovePen(x0, y0, int(verticesX[0]), int(verticesY[0]), context.MovePenColor)

	colr := context.RegionColor
	if amp.AMModifiers[0].(float64) == 0.0 {
		colr = context.ClearColor
	}

	context.RenderOutline(&verticesX, &verticesY, context.RegionColor)
	context.MovePen(int(verticesX[0]), int(verticesY[0]), x0, y0, colr)
}

func (amp AMPrimitivePolygon) Draw(x0, y0 int, x1, y1 int, context *Render) {
	BadMethod()
}

func (amp AMPrimitivePolygon) Init(scale float64, params []float64) AMPrimitive {
	if len(amp.AMModifiers) < 6 {
		glog.Fatalln("unable to create aperture macro primitive polygon - not enough parameters, " +
			strconv.Itoa(len(amp.AMModifiers)) + " given, need 6")
	}
	for i := range amp.AMModifiers {
		amp.AMModifiers[i] = convertToFloat(amp.AMModifiers[i], params)
		if i > 1 && i < 5 {
			switch amp.AMModifiers[i].(type) {
			case float64:
				amp.AMModifiers[i] = scale * (amp.AMModifiers[i]).(float64)
			default:
			}
		}
	}
	return amp
}

func (amp AMPrimitivePolygon) Copy() AMPrimitive {
	var modifStrings []interface{}
	for i := range amp.AMModifiers {
		modifStrings = append(modifStrings, amp.AMModifiers[i])
	}
	return NewAMPrimitive(AMPrimitive_Polygon, modifStrings)
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
	outerDia := amp.AMModifiers[2].(float64)
	rThickness := amp.AMModifiers[3].(float64)
	gap := amp.AMModifiers[4].(float64)
	maxNumRings := int(amp.AMModifiers[5].(float64))
	ringsCount := 0
	ring := AMPrimitiveCircle{AMPrimitive_Circle, []interface{}{}}
	for ringsCount < maxNumRings {
		ring.AMModifiers = append(ring.AMModifiers, 1.0)                            // polarity
		ring.AMModifiers = append(ring.AMModifiers, outerDia)                     // outer diameter
		ring.AMModifiers = append(ring.AMModifiers, amp.AMModifiers[0].(float64)) // center X
		ring.AMModifiers = append(ring.AMModifiers, amp.AMModifiers[1].(float64)) // center Y
		ring.AMModifiers = append(ring.AMModifiers, amp.AMModifiers[8].(float64)) // rot
		ring.AMModifiers = append(ring.AMModifiers, outerDia-2*rThickness)        // thickness of the donut
		ring.Render(x0, y0, context)
		ringsCount++
		ring.AMModifiers = []interface{}{}
		outerDia = outerDia - 2*(rThickness+gap)
		if outerDia <= 0 {
			break
		}
	}
	xHairThickness := amp.AMModifiers[6].(float64)
	xHairLen := amp.AMModifiers[7].(float64)
	if (xHairThickness != 0) && (xHairLen != 0) {
		vectLine := AMPrimitiveVectLine{AMPrimitive_VectLine, []interface{}{}}
		//	"Exposure", "Width", "Start X", "Start Y", "End X", "End Y", "Rotation"
		vectLine.AMModifiers = append(vectLine.AMModifiers, 1.0)                                       // polarity
		vectLine.AMModifiers = append(vectLine.AMModifiers, xHairThickness)                          // width
		vectLine.AMModifiers = append(vectLine.AMModifiers, amp.AMModifiers[0].(float64)-xHairLen/2) // start X
		vectLine.AMModifiers = append(vectLine.AMModifiers, amp.AMModifiers[1].(float64))            // start Y
		vectLine.AMModifiers = append(vectLine.AMModifiers, amp.AMModifiers[0].(float64)+xHairLen/2) // end x
		vectLine.AMModifiers = append(vectLine.AMModifiers, amp.AMModifiers[1].(float64))            // end Y
		vectLine.AMModifiers = append(vectLine.AMModifiers, amp.AMModifiers[8].(float64))            // rot
		vectLine.Render(x0, y0, context)
		vectLine.AMModifiers[2] = amp.AMModifiers[0].(float64)
		vectLine.AMModifiers[3] = amp.AMModifiers[1].(float64) - xHairLen/2
		vectLine.AMModifiers[4] = amp.AMModifiers[0].(float64)
		vectLine.AMModifiers[5] = amp.AMModifiers[1].(float64) + xHairLen/2
		vectLine.Render(x0, y0, context)
	}
	return
}

func (amp AMPrimitiveMoire) Draw(x0, y0 int, x1, y1 int, context *Render) {
	BadMethod()
}

/*
0	Center point X coordinate. A decimal.
1	Center point Y coordinate. A decimal.
2	Outer diameter of outer concentric ring. A decimal ≥ 0.
3	Ring thickness. A decimal ≥ 0.
4	Gap between rings. A decimal ≥ 0.
5	Maximum number of rings. An integer ≥ 0.
	The effective number of rings can be less if the center is reached. If
	there is not enough space for the inner ring it becomes a full disc.
6	Crosshair thickness. A decimal ≥ 0. If the thickness is zero there are
	no rings.
7	Crosshair length. A decimal ≥ 0. If the length is 0 there are no
	crosshairs.
8	Rotation angle, in degrees counterclockwise. A decimal.
	The primitive is rotated around the origin of the macro definition, i.e. the
	(0, 0) point of macro coordinates.
*/

func (amp AMPrimitiveMoire) Init(scale float64, params []float64) AMPrimitive {
	if len(amp.AMModifiers) < 7 {
		glog.Fatalln("unable to create aperture macro primitive moire - not enough parameters, " +
			strconv.Itoa(len(amp.AMModifiers)) + " given, need 7")
	}
	for i := range amp.AMModifiers {
		amp.AMModifiers[i] = convertToFloat(amp.AMModifiers[i], params)
		if i < 5 || (i > 5 && i < 8) {
			switch amp.AMModifiers[i].(type) {
			case float64:
				amp.AMModifiers[i] = scale * (amp.AMModifiers[i]).(float64)
			default:
			}
		}
	}
	return amp
}

func (amp AMPrimitiveMoire) Copy() AMPrimitive {
	var modifStrings []interface{}
	for i := range amp.AMModifiers {
		modifStrings = append(modifStrings, amp.AMModifiers[i])
	}
	return NewAMPrimitive(AMPrimitive_Moire, modifStrings)
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

	rot := amp.AMModifiers[5].(float64)
	innerRadius := amp.AMModifiers[3].(float64) / 2
	gap := amp.AMModifiers[4].(float64)
	phi0 := math.Asin((gap / 2) / innerRadius)
	phi1 := (math.Pi / 2) - phi0
	innerVerticesX, innerVerticesY := GetFirstQuadrantArc(innerRadius, phi0, phi1, context.PenWidth)
	outerRadius := amp.AMModifiers[2].(float64) / 2
	phi0 = math.Asin((gap / 2) / outerRadius)
	phi1 = (math.Pi / 2) - phi0
	outerVerticesX, outerVerticesY := GetFirstQuadrantArc(outerRadius, phi1, phi0, context.PenWidth)

	verticesXI := make([]float64, 0)
	verticesXI = append(verticesXI, *innerVerticesX...)
	verticesXI = append(verticesXI, *outerVerticesX...)
	verticesYI := make([]float64, 0)
	verticesYI = append(verticesYI, *innerVerticesY...)
	verticesYI = append(verticesYI, *outerVerticesY...)

	// 2nd quadrant
	verticesXII := []float64{}
	verticesYII := []float64{}
	for i := range verticesXI {
		verticesXII = append(verticesXII, -verticesXI[i])
		verticesYII = append(verticesYII, verticesYI[i])
	}
	// 3rd quadrant
	verticesXIII := []float64{}
	verticesYIII := []float64{}
	for i := range verticesXII {
		verticesXIII = append(verticesXIII, verticesXII[i])
		verticesYIII = append(verticesYIII, -verticesYI[i])
	}
	// 4rd quadrant
	verticesXIV := []float64{}
	verticesYIV := []float64{}
	for i := range verticesXII {
		verticesXIV = append(verticesXIV, verticesXI[i])
		verticesYIV = append(verticesYIV, verticesYIII[i])
	}

	cx := amp.AMModifiers[0].(float64)
	cy := amp.AMModifiers[1].(float64)

	for i := range verticesXI {
		verticesXI[i] = verticesXI[i] + cx
		verticesYI[i] = verticesYI[i] + cy
		verticesXI[i], verticesYI[i], _ = RotatePoint(verticesXI[i], verticesYI[i], rot)
		verticesXI[i] = float64(x0) + transformFloatCoord(verticesXI[i], context.XRes)
		verticesYI[i] = float64(y0) + transformFloatCoord(verticesYI[i], context.YRes)
	}
	context.MovePen(x0, y0, int(verticesXI[0]), int(verticesYI[0]), context.MovePenColor)
	context.RenderOutline(&verticesXI, &verticesYI, context.RegionColor)
	context.MovePen(int(verticesXI[0]), int(verticesYI[0]), x0, y0, context.MovePenColor)

	for i := range verticesXII {
		verticesXII[i] = verticesXII[i] + cx
		verticesYII[i] = verticesYII[i] + cy
		verticesXII[i], verticesYII[i], _ = RotatePoint(verticesXII[i], verticesYII[i], rot)
		verticesXII[i] = float64(x0) + transformFloatCoord(verticesXII[i], context.XRes)
		verticesYII[i] = float64(y0) + transformFloatCoord(verticesYII[i], context.YRes)
	}
	context.MovePen(x0, y0, int(verticesXII[0]), int(verticesYII[0]), context.MovePenColor)
	context.RenderOutline(&verticesXII, &verticesYII, context.RegionColor)
	context.MovePen(int(verticesXII[0]), int(verticesYII[0]), x0, y0, context.MovePenColor)

	for i := range verticesXIII {
		verticesXIII[i] = verticesXIII[i] + cx
		verticesYIII[i] = verticesYIII[i] + cy
		verticesXIII[i], verticesYIII[i], _ = RotatePoint(verticesXIII[i], verticesYIII[i], rot)
		verticesXIII[i] = float64(x0) + transformFloatCoord(verticesXIII[i], context.XRes)
		verticesYIII[i] = float64(y0) + transformFloatCoord(verticesYIII[i], context.YRes)
	}
	context.MovePen(x0, y0, int(verticesXIII[0]), int(verticesYIII[0]), context.MovePenColor)
	context.RenderOutline(&verticesXIII, &verticesYIII, context.RegionColor)
	context.MovePen(int(verticesXIII[0]), int(verticesYIII[0]), x0, y0, context.MovePenColor)

	for i := range verticesXIV {
		verticesXIV[i] = verticesXIV[i] + cx
		verticesYIV[i] = verticesYIV[i] + cy
		verticesXIV[i], verticesYIV[i], _ = RotatePoint(verticesXIV[i], verticesYIV[i], rot)
		verticesXIV[i] = float64(x0) + transformFloatCoord(verticesXIV[i], context.XRes)
		verticesYIV[i] = float64(y0) + transformFloatCoord(verticesYIV[i], context.YRes)
	}
	context.MovePen(x0, y0, int(verticesXIV[0]), int(verticesYIV[0]), context.MovePenColor)
	context.RenderOutline(&verticesXIV, &verticesYIV, context.RegionColor)
	context.MovePen(int(verticesXIV[0]), int(verticesYIV[0]), x0, y0, context.MovePenColor)

	return
}

func (amp AMPrimitiveThermal) Draw(x0, y0 int, x1, y1 int, context *Render) {
	BadMethod()

}

/*
0	Center point X coordinate. A decimal.
1	Center point Y coordinate. A decimal.
2	Outer diameter. A decimal > inner diameter
3	Inner diameter. A decimal ≥ 0
4	Gap thickness. A decimal < (outer diameter)/√2.
	The gaps are on the X and Y axes through the center without
	rotation. They rotate with the primitive.
	Note that if the (gap thickness)*√2 ≥ (inner diameter) the inner circle
	disappears. This is not invalid.
5	Rotation angle, in degrees counterclockwise. A decimal.
	The primitive is rotated around the origin of the macro definition, i.e.
	(0, 0) point of macro coordinates.
*/
func (amp AMPrimitiveThermal) Init(scale float64, params []float64) AMPrimitive {
	if len(amp.AMModifiers) < 6 {
		glog.Fatalln("unable to create aperture macro primitive thermal - not enough parameters, " +
			strconv.Itoa(len(amp.AMModifiers)) + " given, need 6")
	}
	for i := range amp.AMModifiers {
		amp.AMModifiers[i] = convertToFloat(amp.AMModifiers[i], params)
		if i < 5 {
			switch amp.AMModifiers[i].(type) {
			case float64:
				amp.AMModifiers[i] = scale * (amp.AMModifiers[i]).(float64)
			default:
			}
		}
	}
	return amp
}

func (amp AMPrimitiveThermal) Copy() AMPrimitive {
	var modifStrings []interface{}
	for i := range amp.AMModifiers {
		modifStrings = append(modifStrings, amp.AMModifiers[i])
	}
	return NewAMPrimitive(AMPrimitive_Thermal, modifStrings)
}

// ********************************************* AM container *************************************************
type AMVariable struct {
	Name           string
	Value          string
	PrimitiveIndex int
}

func (amv AMVariable) Copy() AMVariable {
	var retVal AMVariable
	retVal.Name = "" + amv.Name
	retVal.Value = "" + amv.Value
	retVal.PrimitiveIndex = amv.PrimitiveIndex
	return retVal
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

func (am ApertureMacro) Copy() ApertureMacro {
	var retVal ApertureMacro
	retVal.Name = "" + am.Name
	for i := range am.Comments {
		retVal.Comments = append(retVal.Comments, am.Comments[i])
	}
	for i := range am.Variables {
		retVal.Variables = append(retVal.Variables, am.Variables[i].Copy())
	}
	for i := range am.Primitives {
		retVal.Primitives = append(retVal.Primitives, am.Primitives[i].Copy())
	}
	return retVal
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
			retVal.Variables = append(retVal.Variables, AMVariable{s[:eqSignPos], s[eqSignPos+1:], prIndex})

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

func (am ApertureMacro) Draw(x0, y0 int, x1, y1 int, context *Render) {

	for i := range am.Primitives {
		am.Primitives[i].Draw(x0, y0, x1, y1, context)
	}
	return
}

func (am ApertureMacro) Init(scale float64, params []float64) ApertureMacro {
	return am
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
				checkError(err)
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
	code, _ := strconv.Atoi(apString[:i])
	var def = ""
	if commaPos == -1 {
		commaPos = len(apString[i:])
	} else {
		def = apString[commaPos+i+1:]
	}
	name := apString[i : commaPos+i]

	retVal := new(Aperture)
	if len(name) == 0 {
		glog.Fatalln("bad aperture " + strconv.Itoa(code) + " name")
	}
	if len(name) == 1 && (name[0] == 'C' || name[0] == 'R' || name[0] == 'O' || name[0] == 'P') {
		// it's ordinary aperture
		err := retVal.Init2(code, name, def, scale)
		if err != nil {
			glog.Errorln(name + def)
			glog.Fatalln(err)
		}

	} else { // it's macro aperture
		retVal.SourceString = def
		retVal.Type = AptypeMacro
		retVal.Code = code
		// find in macro definitions dictionary for the name
		// TODO Parse def
		var instance ApertureMacro
		params := make([]string, 0)
		if len(def) != 0 {
			params = strings.Split(def, "X")
		}
		ParamsF := make([]float64, 0)
		for i := range params {
			flP, err := strconv.ParseFloat(params[i], 64)
			if err != nil {
				glog.Fatalln("non-number value found in macro parameters")
			}
			ParamsF = append(ParamsF, flP)
		}

		for j := range AMacroDict {
			if strings.Compare(AMacroDict[j].Name, name) == 0 {

				instance = AMacroDict[j].Copy()

				for k := 0; k < len(instance.Primitives); k++ {
					for n := range instance.Variables {
						// have to recalc variables before initializing next primitive
						if instance.Variables[n].PrimitiveIndex == k {
							varStorage := make (map[string]float64)
							for i, pf := range ParamsF {
								varStorage["$"+strconv.Itoa(i+1)] = pf
							}
							varIndex, err := strconv.Atoi(instance.Variables[n].Name[1:])
							if err != nil {
								glog.Fatalln("bad variable name: " + instance.Variables[n].Name)
							}
							addParamsF := varIndex - len(ParamsF)
							for addParamsF > 0 {
								ParamsF = append(ParamsF, 0.0)
								addParamsF--
							}
							ParamsF[varIndex-1] = calculator.CalcExpression(instance.Variables[n].Value,&varStorage)
						}
					}
					instance.Primitives[k] = instance.Primitives[k].Init(scale, ParamsF)
				}
				break
			}
		}
		if len(instance.Name) == 0 {
			glog.Fatalln("unable to instantiate aperture macro " + strconv.Itoa(code) + name)
		}
		retVal.MacroPtr = &instance
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
	var tmpSplitted []string
	if strings.Contains(def[1:], "X") == true {
		tmpSplitted = strings.Split(def, "X")
		for j := range tmpSplitted {
			tmpSplitted[j] = strings.TrimSpace(tmpSplitted[j])
		}
	} else {
		tmpSplitted = append(tmpSplitted, strings.TrimSpace(def))
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

func convertToFloat(arg interface{}, params []float64) float64 {
	panicString1 := "convertToFloat(arg interface{}) float64 - variables not implemented"
	panicString2 := "convertToFloat(arg interface{}) float64 - not supported interface{}"
	panicString3 := "convertToFloat(arg interface{}) float64 - variable has bad name: "
	switch arg.(type) {
	case float64:
		return arg.(float64)
	case string:
		if strings.Contains(arg.(string), "$") == true {
			// detect expression
			if strings.IndexAny(arg.(string), "+-xX/") != -1 {
				// there is an expression
				varStorage := make (map[string]float64)
				for i, f := range params {
					varStorage["$"+ strconv.Itoa(i+1)]  = f
				}
				// calculate
				retVal := calculator.CalcExpression(arg.(string), &varStorage)
				//
				return retVal
			}

			varNum, err := strconv.Atoi(arg.(string)[1:])
			if err != nil {
//				panic(panicString3 + arg.(string))
				glog.Fatal(panicString3 + arg.(string))
			}
			if len(params) >= varNum {
				return params[varNum-1]
			} else {
				return 0
			}
		} else {
			retVal, err := strconv.ParseFloat(arg.(string), 64)
			if err != nil {
				glog.Fatal(panicString1)
			}
			return retVal
		}
	case int:
		return float64(arg.(int))
	default:
		glog.Fatal(panicString2)
	}
	return 0
}

// limits arg by bandVal with respect of sign arg
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

	if x0 != 0 || y0 != 0 {
		phiRad := deg2Rad(phi)

		phi0, r := GetAngle(x0, y0)
		phi1 = phi0 + phiRad
		x1 = r * math.Cos(phi1)
		y1 = r * math.Sin(phi1)
	} else {
		x1 = x0
		y1 = y0
		phi1 = phi
	}
	return x1, y1, phi1
}

// returns angle between hyp and x side
func GetAngle(x, y float64) (angle float64, hyp float64) {
	hyp = math.Hypot(x, y)
	cosPhi := band(x/hyp, 1.0)
	angle = math.Acos(cosPhi)
	if y < 0 {
		angle = math.Pi*2 - angle
	}
	return angle, hyp
}

// returns an 1st quadrant arc
// arcStep - length of a segment of interpolated arc
func GetFirstQuadrantArc(r, phi0, phi1, arcStep float64) (vertX *[]float64, vertY *[]float64) {

	pointsX := make([]float64, 0)
	pointsY := make([]float64, 0)
	phiStep := arcStep / r
	if phi0 > phi1 {
		phiStep = -phiStep
	}
	nSteps := int(math.Round(math.Abs(phi0-phi1) / math.Abs(phiStep)))
	for nSteps > 0 {
		pointsX = append(pointsX, r*math.Cos(phi0))
		pointsY = append(pointsY, r*math.Sin(phi0))
		phi0 += phiStep
		nSteps--
	}
	return &pointsX, &pointsY

}

func BadMethod() {
	glog.Fatal("the macro aperture can not be used to DRAW (D01*)")
}
