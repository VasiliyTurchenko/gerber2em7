package render

import (
	"configurator"
	"errors"
	"fmt"
	"gerbparser"
	"github.com/spf13/viper"
	"image"
	"image/color"
	"math"
	"os"
	"plotter"
	"strconv"

	//	"regions"
)

const (
	MaxInt = int(^uint(0) >> 1)
	MinInt = int(-MaxInt - 1)
)

/*
 ************************** Rendering context ****************************
 */
type Render struct {
	// plotter properties
	// physical plotter single step size
	XRes float64
	YRes float64

	// pen width
	PenWidth float64

	// paper or pcb max dimensions
	CanvasWidth  int // paper property
	CanvasHeight int // paper property
	LimitsX0     int
	LimitsY0     int
	LimitsX1     int
	LimitsY1     int

	// magrin is a safety margin to draw all the border elements of the pcb
	margin float64

	YNeedsFlip bool

	// setPoint size in terms of real plotter pen points
	PointSize  float64
	PointSizeI int
	Plt        *plotter.Plotter
	// pcb properties
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
	// png image properties
	Img          *image.NRGBA
	ApColor      color.RGBA
	LineColor    color.RGBA
	RegionColor  color.RGBA
	ClearColor   color.RGBA
	ObRoundColor color.RGBA
	MovePenColor color.RGBA
	MissedColor  color.RGBA
	ContourColor color.RGBA

	// regions processor
	ProcessingRegion bool

	DrawContours        bool
	DrawMoves           bool
	DrawOnlyRegionsMode bool
	PrintRegionInfo     bool

	//statistic
	LineBresCounter   int
	MovePenCounters   int
	MovePenDistance   float64
	CircleBresCounter int
	LineBresLen       float64
	CircleLen         float64
	FilledRctCounter  int
	ObRoundCounter    int

	// polygon being processed
	polygonPtr *Polygon
}

func NewRender(plotter *plotter.Plotter, viper *viper.Viper, minX, minY, maxX, maxY float64) *Render {
	retVal := new(Render)
	retVal.Init(plotter, viper, minX, minY, maxX, maxY)
	return retVal
}

func (rc *Render) Init(plt *plotter.Plotter, viper *viper.Viper, minX, minY, maxX, maxY float64) {
	// physical plotter single step size
	rc.XRes = viper.GetFloat64(configurator.CfgPlotterXRes)
	rc.YRes = viper.GetFloat64(configurator.CfgPlotterYRes)

	arr := viper.Get(configurator.CfgPlotterPenSizes)
	b, ok := arr.([]interface{})
	if ok == false {
		panic("penSizes configuration error")
	}
	rc.PenWidth = b[0].(float64)

	// paper or pcb max dimensions
	rc.LimitsX0 = 0
	rc.LimitsY0 = 0
	rc.CanvasWidth = 297
	rc.CanvasHeight = 210
	rc.margin = 10.0
	rc.MinX = minX - rc.margin
	rc.MinY = minY - rc.margin
	rc.MaxX = maxX + rc.margin
	rc.MaxY = maxY + rc.margin

	rc.LimitsX1 = int((rc.MaxX - rc.MinX) / float64(rc.XRes))
	rc.LimitsY1 = int((rc.MaxY - rc.MinY) / float64(rc.YRes))

	maxLimX1 := int(float64(rc.CanvasWidth)/float64(rc.XRes))
	maxLimY1 := int(float64(rc.CanvasHeight)/float64(rc.YRes))

	if rc.LimitsX1 > maxLimX1 {
		fmt.Println("Warning: the PCB size X is bigger than plotter working area!")
		fmt.Println("the PCB will be truncated.")
		rc.LimitsX1 = maxLimX1
	}

	if rc.LimitsY1 > maxLimY1 {
		fmt.Println("Warning: the PCB size Y is bigger than plotter working area!")
		fmt.Println("the PCB will be truncated.")
		rc.LimitsY1 = maxLimY1
	}

	rc.Img = image.NewNRGBA(image.Rect(rc.LimitsX0, rc.LimitsY0, rc.LimitsX1, rc.LimitsY1))
	rc.YNeedsFlip = true

	// setPoint size in terms of real plotter pen points
	rc.PointSize = rc.PenWidth / rc.XRes
	rc.PointSizeI = int(math.Round(rc.PointSize))

	rc.ApColor = color.RGBA{255, 0, 0, 255}
	rc.LineColor = color.RGBA{0, 0, 255, 255}
	rc.RegionColor = color.RGBA{255, 0, 255, 255}
	rc.ClearColor = color.RGBA{255, 255, 0, 255}
	rc.ObRoundColor = color.RGBA{0, 127, 0, 255}
	rc.MovePenColor = color.RGBA{100, 100, 100, 255}
	rc.MissedColor = color.RGBA{255, 0, 255, 255}
	rc.ContourColor = color.RGBA{0, 255, 0, 255}

	rc.Plt = plt

	// drawing modes setting

	rc.DrawContours = viper.GetBool(configurator.CfgRenderDrawContours)
	rc.DrawMoves = viper.GetBool(configurator.CfgRenderDrawMoves)
	rc.DrawOnlyRegionsMode = viper.GetBool(configurator.CfgRenderDrawOnlyRegions)
	rc.PrintRegionInfo = viper.GetBool(configurator.CfgPrintRegionInfo)

	return
}

func (rc *Render) DrawFrame() {

	//if (rc.MaxY - rc.margin) <= 0 {
	//	rc.YNeedsFlip = true
	//}
	x2 := transformCoord(rc.MaxX-rc.MinX, rc.XRes)
	y2 := transformCoord(rc.MaxY-rc.MinY, rc.YRes)
	frameColor := color.RGBA{127, 127, 127, 255}
	rc.bresenhamWithPattern(0, 0, x2, 0, 1, frameColor, 10, 10)
	rc.bresenhamWithPattern(x2, 0, x2, y2, 1, frameColor, 10, 10)
	rc.bresenhamWithPattern(x2, y2, 0, y2, 1, frameColor, 10, 10)
	rc.bresenhamWithPattern(0, y2, 0, 0, 1, frameColor, 10, 10)

}

/*----------------------------------------------*/
// modified 07-Jun-2018
// draws a point
func (rc *Render) setPoint(x, y, pointSize int, col color.Color) {
	if pointSize < 0 {
		return
	}
	if rc.DrawContours == false {
		// Draw by bresenham algorithm
		x1, y1, err := -pointSize, 0, 2-2*pointSize
		for {
			rc.Img.Set(x-x1, y+y1, col)
			rc.Img.Set(x-y1, y-x1, col)
			rc.Img.Set(x+x1, y-y1, col)
			rc.Img.Set(x+y1, y+x1, col)
			pointSize = err
			if pointSize > x1 {
				x1++
				err += x1*2 + 1
			}
			if pointSize <= y1 {
				y1++
				err += y1*2 + 1
			}
			if x1 >= 0 {
				break
			}
		}
	} else {
		rc.Img.Set(x, y, col)
	}
}

// for D01 commands
func (rc *Render) drawByRectangleAperture(x0, y0, x1, y1, apSizeX, apSizeY int, col color.Color) {

	var w, h, xOrigin, yOrigin int

	if x0 != x1 && y0 != y1 {
		fmt.Println("Drawing by rectangular aperture with arbitrary angle is not supported!")
		rc.drawCircle(x0, y0, apSizeX/2, rc.PointSizeI, rc.MissedColor)
		rc.drawCircle(x1, y1, apSizeX/2, rc.PointSizeI, rc.MissedColor)
	}
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}

	if x0 == x1 { // vertical draw
		xOrigin = x0
		yOrigin = y0 + (y1-y0)/2
		h = y1 - y0 + apSizeY
		w = apSizeX
		// draw by pen from x0,y0 to rectangle's origin
		rc.drawByBrezenham(x0, y0, xOrigin, yOrigin, rc.PointSizeI, col)
		rc.drawFilledRectangle(xOrigin, yOrigin, w, h, col)
		// draw back by pen from rectangle's origin to x1, y1
		rc.drawByBrezenham(xOrigin, yOrigin, x1, y1, rc.PointSizeI, col)
		return
	}
	if y0 == y1 { // horizontal draw
		yOrigin = y0
		xOrigin = x0 + (x1-x0)/2
		w = x1 - x0 + apSizeX
		h = apSizeY
		// draw by pen from x0,y0 to rectangle's origin
		rc.drawByBrezenham(x0, y0, xOrigin, yOrigin, rc.PointSizeI, col)
		rc.drawFilledRectangle(xOrigin, yOrigin, w, h, col)
		rc.drawByBrezenham(xOrigin, yOrigin, x1, y1, rc.PointSizeI, col)
		return
	}
}

// for D01 commands
func (rc *Render) drawByCircleAperture(x0, y0, x1, y1, apDia int, col color.Color) {
	// save x0, y0, x1, y1
	savedx0 := x0
	savedy0 := y0
	savedx1 := x1
	savedy1 := y1

	var xPen, yPen int
	ptsz := rc.PointSizeI

	rc.drawDonut(x0, y0, apDia, 0, col)

	if y1 < y0 {
		x0, y0, x1, y1 = x1, y1, x0, y0
	} // now y0 definitely less than y1

	if y0 == y1 { // horizontal draw
		if x0 > x1 {
			// swap x0, x1
			x0, x1 = x1, x0
		}
		yOrigin := y0
		xOrigin := x0 + (x1-x0)/2
		// draw by pen to xOrigin, y Origin
		xPen, yPen = rc.drawByBrezenham(savedx0, savedy0, xOrigin, yOrigin, ptsz, col)
		w := x1 - x0
		h := apDia
		rc.drawFilledRectangle(xOrigin, yOrigin, w, h, col)
		// move pen back to original x1, y1 setPoint
		xPen, yPen = rc.drawByBrezenham(xOrigin, yOrigin, savedx1, savedy1, ptsz, col)
		rc.drawDonut(savedx1, savedy1, apDia, 0, col)
		_, _ = xPen, yPen
		return
	}
	if x0 == x1 {
		// y0 < y1 always here
		// vertical draw
		xOrigin := x0
		yOrigin := y0 + (y1-y0)/2
		h := y1 - y0
		w := apDia
		// draw by pen to xOrigin, y Origin
		xPen, yPen = rc.drawByBrezenham(savedx0, savedy0, xOrigin, yOrigin, ptsz, col)
		rc.drawFilledRectangle(xOrigin, yOrigin, w, h, col)
		// move pen back to original x1, y1 setPoint
		xPen, yPen = rc.drawByBrezenham(xOrigin, yOrigin, savedx1, savedy1, ptsz, col)
		rc.drawDonut(savedx1, savedy1, apDia, 0, col)
		_, _ = xPen, yPen
		return
	}
	// non-orthogonal draw
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	l := math.Hypot(dx, dy)
	var sdelta = 1.0
	if x1 > x0 {
		sdelta = -1.0
	}
	var lines = apDia / ptsz
	if lines < 1 {
		lines = 1
	}
	alpha := math.Asin(dy / l) //
	hypo := (float64(apDia) / 2) - (float64(ptsz) / 2)
	sin1 := math.Sin((math.Pi / 2) - alpha)
	cos1 := math.Cos((math.Pi / 2) - alpha)
	yv0 := float64(y0) - hypo*sin1
	xv0 := float64(x0) - (sdelta * hypo * cos1)
	dxv := math.Abs(float64(ptsz) * cos1)
	dyv := math.Abs(float64(ptsz) * sin1)
	var nx0, nx1, ny0, ny1 int
	for i := 0; i < lines; i++ {
		nx0 = int(math.Round(xv0))
		ny0 = int(math.Round(yv0))
		if i == 0 {
			// draw to start setPoint
			xPen, yPen = rc.drawByBrezenham(savedx0, savedy0, nx0, ny0, ptsz, col)
		}
		nx1 = int(math.Round(xv0 + dx))
		ny1 = int(math.Round(yv0 + dy))
		xPen, yPen = rc.drawByBrezenham(nx0, ny0, nx1, ny1, ptsz, col)
		xv0 = xv0 + sdelta*dxv
		yv0 = yv0 + dyv
	}
	// draw back to saved x1, y1
	xPen, yPen = rc.drawByBrezenham(nx1, ny1, savedx1, savedy1, ptsz, col)
	// and final drawDonut
	rc.drawDonut(savedx1, savedy1, apDia, 0, col)
	_, _ = xPen, yPen
}

// for aperture flash D03
//const strat int = 0  // closed rectangles inserted each into other
const strat int = 1 // zig-zag

// draws a filled rectangle
func (rc *Render) drawFilledRectangle(origX, origY, w, h int, col color.Color) {

	xPen := origX // real pen position
	yPen := origY // real pen position

	// performs rectangle aperture flash
	x0 := origX - (w / 2)
	y0 := origY - (h / 2)
	x1 := origX + (w / 2)
	y1 := origY + (h / 2)

	if rc.DrawContours == true {
		rc.drawByBrezenham(x0, y0, x1, y0, 1, rc.ContourColor)
		rc.drawByBrezenham(x1, y0, x1, y1, 1, rc.ContourColor)
		rc.drawByBrezenham(x1, y1, x0, y1, 1, rc.ContourColor)
		rc.drawByBrezenham(x0, y1, x0, y0, 1, rc.ContourColor)
	}
	x0 = x0 + (rc.PointSizeI / 2)
	y0 = y0 + (rc.PointSizeI / 2)

	x1 = x1 - (rc.PointSizeI / 2)
	y1 = y1 - (rc.PointSizeI / 2)

	// imitate pen moving to the start setPoint
	rc.drawByBrezenham(origX, origY, x0, y0, rc.PointSizeI, col)

	// draw contour
	xPen, yPen = rc.drawByBrezenham(x0, y0, x1, y0, rc.PointSizeI, col)
	xPen, yPen = rc.drawByBrezenham(x1, y0, x1, y1, rc.PointSizeI, col)
	xPen, yPen = rc.drawByBrezenham(x1, y1, x0, y1, rc.PointSizeI, col)
	xPen, yPen = rc.drawByBrezenham(x0, y1, x0, y0, rc.PointSizeI, col)

	xp := x0
	yp := y0

	x0 = x0 + rc.PointSizeI
	y0 = y0 + rc.PointSizeI
	x1 = x1 - rc.PointSizeI
	y1 = y1 - rc.PointSizeI

	rc.drawByBrezenham(xp, yp, x0, y0, rc.PointSizeI, col)

	if strat == 0 {
		for {
			xPen, yPen = rc.drawByBrezenham(x0, y0, x1, y0, rc.PointSizeI, col)
			xPen, yPen = rc.drawByBrezenham(x1, y0, x1, y1, rc.PointSizeI, col)
			xPen, yPen = rc.drawByBrezenham(x1, y1, x0, y1, rc.PointSizeI, col)
			xPen, yPen = rc.drawByBrezenham(x0, y1, x0, y0, rc.PointSizeI, col)

			x0 = x0 + rc.PointSizeI
			x1 = x1 - rc.PointSizeI
			y0 = y0 + rc.PointSizeI
			y1 = y1 - rc.PointSizeI

			if ((x1 - x0) < 0) || ((y1 - y0) < 0) {
				break
			}
		}
	}
	if strat == 1 {
		if w > h {
			var tmpY int
			var retX int
			for {
				xPen, yPen = rc.drawByBrezenham(x0, y0, x1, y0, rc.PointSizeI, col)
				tmpY = y0
				y0 = y0 + rc.PointSizeI
				if y0 > y1 {
					retX = x1
					break
				}
				xPen, yPen = rc.drawByBrezenham(x1, tmpY, x1, y0, rc.PointSizeI, col)
				xPen, yPen = rc.drawByBrezenham(x1, y0, x0, y0, rc.PointSizeI, col)
				tmpY = y0
				y0 = y0 + rc.PointSizeI
				if y0 > y1 {
					retX = x0
					break
				}
				xPen, yPen = rc.drawByBrezenham(x0, tmpY, x0, y0, rc.PointSizeI, col)
			}
			// imitate pen moving to the origin setPoint
			xPen, yPen = rc.drawByBrezenham(retX, tmpY, origX, origY, rc.PointSizeI, col)
		} else {
			var tmpX int
			var retY int
			for {
				xPen, yPen = rc.drawByBrezenham(x0, y0, x0, y1, rc.PointSizeI, col)
				tmpX = x0
				x0 = x0 + rc.PointSizeI
				if x0 > x1 {
					retY = y1
					break
				}
				xPen, yPen = rc.drawByBrezenham(tmpX, y1, x0, y1, rc.PointSizeI, col)
				xPen, yPen = rc.drawByBrezenham(x0, y1, x0, y0, rc.PointSizeI, col)
				tmpX = x0
				x0 = x0 + rc.PointSizeI
				if x0 > x1 {
					retY = y0
					break
				}
				xPen, yPen = rc.drawByBrezenham(tmpX, y0, x0, y0, rc.PointSizeI, col)
			}
			// imitate pen moving to the origin setPoint
			xPen, yPen = rc.drawByBrezenham(tmpX, retY, origX, origY, rc.PointSizeI, col)
		}
	}
	if xPen != origX || yPen != origY {
		fmt.Println("Error during filled rectangle drawing: pen did not returned to the origin setPoint!")
		os.Exit(700)
	}
	rc.FilledRctCounter++
}

func (rc *Render) drawDonut(origX, origY, dia, holeDia int, col color.Color) {
	// performs drawDonut (drawCircle) aperture flash
	radius := dia / 2
	holeRadius := holeDia / 2
	if rc.DrawContours == true {
		rc.drawCircle(origX, origY, radius, 1, rc.ContourColor)
		if holeDia > 0 {
			rc.drawCircle(origX, origY, holeRadius, 1, rc.ContourColor)
		}
	}
	radius = radius - (rc.PointSizeI / 2)
	for {
		rc.drawCircle(origX, origY, radius, rc.PointSizeI, col)
		radius = radius - rc.PointSizeI
		if radius < holeRadius+(rc.PointSizeI/2) {
			break
		}
	}
}

// drawCircle plots a circle with center x, y and radius r.
// Limiting behavior:
// r < 0 plots no pixels.
// r = 0 plots a single pixel at x, y.
// r = 1 plots four pixels in a diamond shape around the center pixel at x, y.
func (rc *Render) drawCircle(x, y, r, ptsz int, col color.Color) {
	if r < 0 {
		return
	}
	// statistics
	rc.CircleBresCounter++
	rc.CircleLen += 2 * math.Pi * float64(r)

	rc.Plt.Circle(x, y, r)

	// Draw By bresenham algorithm
	x1, y1, err := -r, 0, 2-2*r
	for {
		rc.setPoint(x-x1, y+y1, ptsz, col)
		rc.setPoint(x-y1, y-x1, ptsz, col)
		rc.setPoint(x+x1, y-y1, ptsz, col)
		rc.setPoint(x+y1, y+x1, ptsz, col)
		r = err
		if r > x1 {
			x1++
			err += x1*2 + 1
		}
		if r <= y1 {
			y1++
			err += y1*2 + 1
		}
		if x1 >= 0 {
			break
		}
	}
}

// Move pen function

func (rc *Render) movePen(x1, y1, x2, y2 int, col color.Color) (int, int) {
	rc.MovePenCounters++
	rc.MovePenDistance += math.Hypot(float64(x2-x1), float64(y2-y1))
	newX := x2
	newY := y2
	if rc.DrawMoves == true {
		newX, newY = rc.bresenham(x1, y1, x2, y2, 1, col)
	}
	rc.Plt.MoveTo(x2, y2)
	return newX, newY
}

func (rc *Render) drawByBrezenham(x1, y1, x2, y2, pointSize int, col color.Color) (int, int) {
	// statistics
	rc.LineBresCounter++
	rc.LineBresLen += math.Hypot(float64(x2-x1), float64(y2-y1))
	newX, newY := rc.bresenham(x1, y1, x2, y2, pointSize, col)
	rc.Plt.DrawLine(x1, y1, x2, y2)
	return newX, newY
}

// Generalized with integer
func (rc *Render) bresenham(x1, y1, x2, y2, pointSize int, col color.Color) (int, int) {
	var dx, dy, e, slope int
	newX := x2
	newY := y2
	// Because drawing p1 -> p2 is equivalent to draw p2 -> p1,
	// I sort points in x-axis order to handle only half of possible cases.
	if x1 > x2 {
		x1, y1, x2, y2 = x2, y2, x1, y1
	}

	dx, dy = x2-x1, y2-y1
	// Because setPoint is x-axis ordered, dx cannot be negative
	if dy < 0 {
		dy = -dy
	}

	switch {

	// Is line a setPoint ?
	case x1 == x2 && y1 == y2:
		rc.setPoint(x1, y1, pointSize, col)

		// Is line an horizontal ?
	case y1 == y2:
		for ; dx != 0; dx-- {
			rc.setPoint(x1, y1, pointSize, col)
			x1++
		}
		rc.setPoint(x1, y1, pointSize, col)

		// Is line a vertical ?
	case x1 == x2:
		if y1 > y2 {
			y1, y2 = y2, y1
		}
		for ; dy != 0; dy-- {
			rc.setPoint(x1, y1, pointSize, col)
			y1++
		}
		rc.setPoint(x1, y1, pointSize, col)

		// Is line a diagonal ?
	case dx == dy:
		if y1 < y2 {
			for ; dx != 0; dx-- {
				rc.setPoint(x1, y1, pointSize, col)
				x1++
				y1++
			}
		} else {
			for ; dx != 0; dx-- {
				rc.setPoint(x1, y1, pointSize, col)
				x1++
				y1--
			}
		}
		rc.setPoint(x1, y1, pointSize, col)

		// wider than high ?
	case dx > dy:
		if y1 < y2 {
			dy, e, slope = 2*dy, dx, 2*dx
			for ; dx != 0; dx-- {
				rc.setPoint(x1, y1, pointSize, col)
				x1++
				e -= dy
				if e < 0 {
					y1++
					e += slope
				}
			}
		} else {
			dy, e, slope = 2*dy, dx, 2*dx
			for ; dx != 0; dx-- {
				rc.setPoint(x1, y1, pointSize, col)
				x1++
				e -= dy
				if e < 0 {
					y1--
					e += slope
				}
			}
		}
		rc.setPoint(x2, y2, pointSize, col)

		// higher than wide.
	default:
		if y1 < y2 {
			dx, e, slope = 2*dx, dy, 2*dy
			for ; dy != 0; dy-- {
				rc.setPoint(x1, y1, pointSize, col)
				y1++
				e -= dx
				if e < 0 {
					x1++
					e += slope
				}
			}
		} else {
			dx, e, slope = 2*dx, dy, 2*dy
			for ; dy != 0; dy-- {
				rc.setPoint(x1, y1, pointSize, col)
				y1--
				e -= dx
				if e < 0 {
					x1++
					e += slope
				}
			}
		}
		rc.setPoint(x2, y2, pointSize, col)
	}
	return newX, newY
}

// draw line using pattern
// dash - length of the dash in pixels
// space - length of the space in pixels
func (rc *Render) bresenhamWithPattern(x1, y1, x2, y2, pointSize int, col color.Color, dash, space int) (int, int) {
	length := int(math.Hypot(float64(x2-x1), float64(y2-y1)))
	if length == 0 {
		return x1, y1
	}
	if x1 > x2 {
		x1, y1, x2, y2 = x2, y2, x1, y1
	}

	dx := x2 - x1
	dy := y2 - y1
	signdY := 1

	if dy < 0 {
		dy = -dy
		signdY = -1
	}

	phi := math.Acos(float64(dx) / float64(length))

	steps := length / (dash + space)
	x01 := x1
	y01 := y1
	dashX := int(float64(dash) * math.Cos(phi))
	dashY := int(float64(dash)*math.Sin(phi)) * signdY
	spaceX := int(float64(space) * math.Cos(phi))
	spaceY := int(float64(space)*math.Sin(phi)) * signdY

	for steps > 0 {
		x11 := x01 + dashX
		y11 := y01 + dashY
		x01, y01 = rc.bresenham(x01, y01, x11, y11, pointSize, col)
		x01 += spaceX
		y01 += spaceY
		steps--
	}
	// x01, y01 here are the coordinates of the last dash line
	if x01+dashX > x2 {
		dashX = x2 - x01
		dashY = y2 - y01
	}
	x11 := x01 + dashX
	y11 := y01 + dashY

	return rc.bresenham(x01, y01, x11, y11, pointSize, col)
}

// ARC functions
// TODO test G:\gerbv-2.6.2\example\cslk fails!
func (rc *Render) drawArc(x1, y1, x2, y2, i, j float64, apertureSize int, ipm gerbparser.IPmode, qm gerbparser.QuadMode, col color.Color) error {

	var xC, yC float64

	if qm == gerbparser.QuadModeSingle {
		// we have to find the sign of the I and J
		fmt.Println("G74 hook")
		return nil
	}
	if rc.DrawContours == true {
		rc.setPoint(int(x1), int(y1), 1, rc.ContourColor)
		rc.setPoint(int(x2), int(y2), 1, rc.ContourColor)
	}
	xC = x1 + i
	yC = y1 + j
	r := math.Hypot(i, j)
	rt := math.Hypot(x2-xC, y2-yC)

	dR := rt - r
	if math.Abs(dR) > float64(rc.PointSizeI) {

		errString := "G75 diff.=" + strconv.FormatFloat(rt-r, 'f', 5, 64) + "\n"
		errString = errString + "x1=" + strconv.FormatFloat(x1, 'f', 5, 64) +
			" y1=" + strconv.FormatFloat(y1, 'f', 5, 64) +
			" x2=" + strconv.FormatFloat(x2, 'f', 5, 64) +
			" y2=" + strconv.FormatFloat(y2, 'f', 5, 64) +
			" i=" + strconv.FormatFloat(i, 'f', 5, 64) +
			" j=" + strconv.FormatFloat(j, 'f', 5, 64)
		return errors.New(errString)
	}

	r = (r + rt) / 2

	cosPhi1 := (x1 - xC) / r
	if cosPhi1 > 1 {
		cosPhi1 = 1
	} else if cosPhi1 < -1 {
		cosPhi1 = -1
	}

	Phi1 := rad2Deg(math.Acos(cosPhi1))
	if float64(y1)-yC < 0 {
		Phi1 = 360.0 - Phi1
	}

	cosPhi2 := (x2 - xC) / r
	if cosPhi2 > 1 {
		cosPhi2 = 1
	} else if cosPhi2 < -1 {
		cosPhi2 = -1
	}
	Phi2 := rad2Deg(math.Acos(cosPhi2))
	if float64(y2)-yC < 0 {
		Phi2 = 360.0 - Phi2
	}

	numArcs := apertureSize / rc.PointSizeI // how many arcs to do..
	r = r + (float64(apertureSize) / 2) - (float64(rc.PointSizeI) / 2)
	for i := 0; i < numArcs; i++ {
		r := r - float64(i*rc.PointSizeI)

		if ipm == gerbparser.IPModeCCwC {
			var ppx = 0
			var ppy = 0
			if Phi1 > Phi2 {
				Phi1 = -(360.0 - Phi1)
			}
			plX1 := int(math.Round(x1))
			plX2 := int(math.Round(x2))
			plY1 := int(math.Round(y1))
			plY2 := int(math.Round(y2))
			plR := int(math.Round(r))
			plPhi1 := int(math.Round(Phi1))
			plPhi2 := int(math.Round(Phi2))

			rc.Plt.Arc(plX1, plY1, plX2, plY2, plR, plPhi1, plPhi2, ipm)

			angle := Phi1
			for {
				ax := int(math.Round(r*math.Cos(deg2Rad(angle)) + xC))
				ay := int(math.Round(r*math.Sin(deg2Rad(angle)) + yC))
				if ppx != ax || ppy != ay {
					rc.drawCircle(ax, ay, rc.PointSizeI, 1, col)
				}
				angle++
				if angle > Phi2 {
					break
				}
				ppx = ax
				ppy = ay
			}

		} else if ipm == gerbparser.IPModeCwC {
			var ppx = 0
			var ppy = 0
			if Phi1 < Phi2 {
				Phi2 = -(360.0 - Phi2)
			}

			plX1 := int(math.Round(x1))
			plX2 := int(math.Round(x2))
			plY1 := int(math.Round(y1))
			plY2 := int(math.Round(y2))
			plR := int(math.Round(r))
			plPhi1 := int(math.Round(Phi1))
			plPhi2 := int(math.Round(Phi2))

			rc.Plt.Arc(plX1, plY1, plX2, plY2, plR, plPhi1, plPhi2, ipm)

			angle := Phi1
			for {
				ax := int(math.Round(r*math.Cos(deg2Rad(angle)) + xC))
				ay := int(math.Round(r*math.Sin(deg2Rad(angle)) + yC))
				if ppx != ax || ppy != ay {
					rc.drawCircle(ax, ay, rc.PointSizeI, 1, col)
				}
				angle--
				if angle < Phi2 {
					break
				}
				ppx = ax
				ppy = ay
			}
		}
	}

	return nil
}

// obround aperture flash
func (rc *Render) drawObRound(centerX, centerY, width, height, holeDia int, color color.Color) {
	var sideDia int
	if width > height {
		sideDia = height
		rc.drawFilledRectangle(centerX, centerY, width-sideDia, height, color)
		xd1 := centerX - (width / 2) + (sideDia / 2)
		xd2 := centerX + (width / 2) - (sideDia / 2)
		rc.drawDonut(xd1, centerY, sideDia, holeDia, color)
		rc.drawDonut(xd2, centerY, sideDia, holeDia, color)
	} else {
		sideDia = width
		rc.drawFilledRectangle(centerX, centerY, width, height-sideDia, color)
		yd1 := centerY - (height / 2) + (sideDia / 2)
		yd2 := centerY + (height / 2) - (sideDia / 2)
		rc.drawDonut(centerX, yd1, sideDia, holeDia, color)
		rc.drawDonut(centerX, yd2, sideDia, holeDia, color)
	}
	rc.ObRoundCounter++
}

//
func rad2Deg(a float64) float64 {
	return 360 * a / (2 * math.Pi)
}

//
func deg2Rad(a float64) float64 {
	return (a / 360) * (2 * math.Pi)
}

//
func abs(x int) int {
	switch {
	case x >= 0:
		return x
	case x >= MinInt:
		return -x
	}
	panic("math/int.Abs: invalid argument")
}

/*
**************************** step processor *******************************
 */
func (rc *Render) ProcessStep(stepData *gerbparser.State) {

//	stepData.Print()

	var Xp int
	var Yp int
	Xc := transformCoord(stepData.Coord.GetX()-rc.MinX, rc.XRes)
	Yc := transformCoord(stepData.Coord.GetY()-rc.MinY, rc.YRes)
	if stepData.PrevCoord == nil {
		Xp = transformCoord(0-rc.MinX, rc.XRes)
		Yp = transformCoord(0-rc.MinY, rc.YRes)
	} else {
		Xp = transformCoord(stepData.PrevCoord.GetX()-rc.MinX, rc.XRes)
		Yp = transformCoord(stepData.PrevCoord.GetY()-rc.MinY, rc.YRes)
	}

	if stepData.Region != nil {
		// process region
		if rc.polygonPtr == nil {
			rc.polygonPtr = newPolygon()
		}
		if rc.addStepToPolygon(stepData) == stepData.Region.GetNumXY() {
			// we can process region
			rc.renderPolygon()
			rc.polygonPtr = nil
		}
	} else {
		var stepColor color.RGBA
		switch stepData.Action {
		case gerbparser.OpcodeD01_DRAW: // draw
			if stepData.Polarity == gerbparser.PolTypeDark {
				stepColor = rc.LineColor
			} else {
				stepColor = rc.ClearColor
			}

			var apertureSize int
			if abs(Xc-Xp) < (4*rc.PointSizeI) && abs(Yc-Yp) < (4*rc.PointSizeI) {
				stepData.IpMode = gerbparser.IPModeLinear
			}
			if stepData.IpMode == gerbparser.IPModeLinear {
				// linear interpolation
				if rc.DrawOnlyRegionsMode != true {
					if stepData.CurrentAp.Type == gerbparser.AptypeCircle {
						apertureSize = transformCoord(stepData.CurrentAp.Diameter, rc.XRes)
						rc.drawByCircleAperture(Xp, Yp, Xc, Yc, apertureSize, stepColor)
					} else if stepData.CurrentAp.Type == gerbparser.AptypeRectangle {
						// draw with rectangle aperture
						w := transformCoord(stepData.CurrentAp.XSize, rc.XRes)
						h := transformCoord(stepData.CurrentAp.YSize, rc.YRes)
						rc.drawByRectangleAperture(Xp, Yp, Xc, Yc, w, h, stepColor)
					} else {
						fmt.Println("Error. Only solid drawCircle and solid rectangle may be used to draw.")
						break
					}
				}
			} else {
				// non-linear interpolation
				if rc.DrawOnlyRegionsMode != true {
					if stepData.CurrentAp.Type == gerbparser.AptypeCircle {
						apertureSize = transformCoord(stepData.CurrentAp.Diameter, rc.XRes)
						var (
							fXp, fYp float64
						)
						if stepData.PrevCoord == nil {
							fXp = transformFloatCoord(0-rc.MinX, rc.XRes)
							fYp = transformFloatCoord(0-rc.MinY, rc.YRes)
						} else {
							fXp = transformFloatCoord(stepData.PrevCoord.GetX()-rc.MinX, rc.XRes)
							fYp = transformFloatCoord(stepData.PrevCoord.GetY()-rc.MinY, rc.YRes)
						}

						fXc := transformFloatCoord(stepData.Coord.GetX()-rc.MinX, rc.XRes)
						fYc := transformFloatCoord(stepData.Coord.GetY()-rc.MinY, rc.YRes)
						fI := transformFloatCoord(stepData.Coord.GetI(), rc.XRes)
						fJ := transformFloatCoord(stepData.Coord.GetJ(), rc.YRes)

						// Arcs require floats!
						err:= rc.drawArc(fXp,
							fYp,
							fXc,
							fYc,
							fI,
							fJ,
							apertureSize,
							stepData.IpMode,
							stepData.QMode,
							// TODO
							rc.RegionColor)
						if err != nil {
							stepData.Print()
							checkError(err, 998)
						}
						rc.drawDonut(Xp, Yp, apertureSize, 0, stepColor)
						rc.drawDonut(Xc, Yc, apertureSize, 0, stepColor)
					} else if stepData.CurrentAp.Type == gerbparser.AptypeRectangle {
						fmt.Println("Arc drawing by rectangle aperture is not supported now.")
					} else {
						fmt.Println("Error. Only solid drawCircle and solid rectangle may be used to draw.")
						break
					}
				}
			}
			//
		case gerbparser.OpcodeD02_MOVE: // move
			rc.movePen(Xp, Yp, Xc, Yc, rc.MovePenColor)
		//
		case gerbparser.OpcodeD03_FLASH: // flash
			if rc.DrawOnlyRegionsMode != true {
				rc.movePen(Xp, Yp, Xc, Yc, rc.MovePenColor)
				if stepData.Polarity == gerbparser.PolTypeDark {
					stepColor = rc.ApColor
				} else {
					stepColor = rc.ClearColor
				}
				w := transformCoord(stepData.CurrentAp.XSize, rc.XRes)
				h := transformCoord(stepData.CurrentAp.YSize, rc.YRes)
				d := transformCoord(stepData.CurrentAp.Diameter, rc.XRes)
				hd := transformCoord(stepData.CurrentAp.HoleDiameter, rc.XRes)

				switch stepData.CurrentAp.Type {
				case gerbparser.AptypeRectangle:
					rc.drawFilledRectangle(Xc, Yc, w, h, stepColor)
				case gerbparser.AptypeCircle:
					rc.drawDonut(Xc, Yc, d, hd, stepColor)
				case gerbparser.AptypeObround:
					if w == h {
						rc.drawDonut(Xc, Yc, w, hd, stepColor)
					} else {
						rc.drawObRound(Xc, Yc, w, h, 0, rc.ObRoundColor)
					}
				case gerbparser.AptypePoly:
					rc.drawDonut(Xc, Yc, d, hd, rc.MissedColor)
					fmt.Println("Polygonal apertures ain't supported.")
				default:
					checkError(errors.New("bad aperture type found"), 501)
					break
				}
			}
		default:
			checkError(errors.New("(rc *Render) ProcessStep(stepData *gerbparser.State) internal error. Bad opcode"), 666)
			fmt.Println("")
			break
		}
	}
}

/* some draw helpers */

func transformCoord(inc float64, res float64) int {
	return int(inc / res)
}

func transformFloatCoord(inc float64, res float64) float64 {
	return inc / res
}

func checkError(err error, exitCode int) {
	if err != nil {
		fmt.Println(err)
		os.Exit(exitCode)
	}
}

/*
*********************** region (polygon) processor ***********************************
 */

type Polygon struct {
	steps       *[]*gerbparser.State
	polX        *[]float64
	polY        *[]float64
	numVertices int
}

func newPolygon() *Polygon {
	retVal := new(Polygon)
	steps := make([]*gerbparser.State, 0)
	retVal.steps = &steps
	polX := make([]float64, 0)
	polY := make([]float64, 0)
	retVal.polX = &polX
	retVal.polY = &polY
	return retVal
}

func (rc *Render) addStepToPolygon(step *gerbparser.State) int {
	*rc.polygonPtr.steps = append(*rc.polygonPtr.steps, step)
	return len(*rc.polygonPtr.steps)
}

func (rc *Render) renderPolygon() {

	if (*rc.polygonPtr.steps)[0].Action == gerbparser.OpcodeD02_MOVE {
		*rc.polygonPtr.steps = (*rc.polygonPtr.steps)[1:]
	}
	prev := (*rc.polygonPtr.steps)[0].PrevCoord
	// check if the region contains self-intersections or is not closed
	for i := 0; i < len(*rc.polygonPtr.steps); i++ {
		if (*rc.polygonPtr.steps)[i].Coord.Equals(prev, 0.001) {
			if rc.PrintRegionInfo == true {
				fmt.Println("Closed segment found with  ", i, "vertexes")
			}
			if i < len(*rc.polygonPtr.steps)-2 {
				fmt.Println("More than one segment in the region!")
				fmt.Println("There is", (len(*rc.polygonPtr.steps) - 2 - i), "points are left out of the region")
			}
			break
		}
		if i == len(*rc.polygonPtr.steps)-1 {
			// the segment is not closed!
			fmt.Println("The segment is not closed!")
			fmt.Println(prev.String())
			fmt.Println((*rc.polygonPtr.steps)[0].Coord.String())
			fmt.Println((*rc.polygonPtr.steps)[len(*rc.polygonPtr.steps)-2].Coord.String())
			fmt.Println((*rc.polygonPtr.steps)[len(*rc.polygonPtr.steps)-1].Coord.String())
			os.Exit(1000)
		}
	}

	// let's create a array of nodes (vertices)
	rc.polygonPtr.numVertices = len(*rc.polygonPtr.steps)
	minYInPolygon := 100000000.0
	maxYInPolygon := 0.0
	for j := 0; j < rc.polygonPtr.numVertices; j++ {
		if (*rc.polygonPtr.steps)[j].IpMode != gerbparser.IPModeLinear {
			//			rc.interpolate(&minYInPolygon, &maxYInPolygon, &steps[j])
			rc.interpolate(&minYInPolygon, &maxYInPolygon, (*rc.polygonPtr.steps)[j])
		} else {
			xj := ((*rc.polygonPtr.steps)[j].Coord.GetX() - rc.MinX) / rc.XRes
			yj := ((*rc.polygonPtr.steps)[j].Coord.GetY() - rc.MinY) / rc.YRes
			if yj < minYInPolygon {
				minYInPolygon = yj
			}
			if yj > maxYInPolygon {
				maxYInPolygon = yj
			}
			*rc.polygonPtr.polX = append(*rc.polygonPtr.polX, xj)
			*rc.polygonPtr.polY = append(*rc.polygonPtr.polY, yj)
		}
	}
	rc.polygonPtr.numVertices = len(*rc.polygonPtr.polX)

	var nodes = 0
	var nodeX []int
	nodeX = make([]int, rc.polygonPtr.numVertices)
	var pixelY int

	// take into account real plotter pen setPoint size
	startY := int(math.Round(minYInPolygon + rc.PointSize/2))
	stopY := int(math.Round(maxYInPolygon - rc.PointSize/2))
	marginX := int(math.Round(rc.PointSize / 2))

	// fill the inner points of the polygon
	var i int = 0
	for pixelY = startY; pixelY < stopY; pixelY += rc.PointSizeI {
		fPixelY := float64(pixelY)
		nodes = 0
		j := rc.polygonPtr.numVertices - 1
		for i = 0; i < rc.polygonPtr.numVertices; i++ {
			if ((*rc.polygonPtr.polY)[i] < fPixelY && (*rc.polygonPtr.polY)[j] >= fPixelY) ||
				((*rc.polygonPtr.polY)[j] < fPixelY && (*rc.polygonPtr.polY)[i] >= fPixelY) {
				nodeX[nodes] = int((*rc.polygonPtr.polX)[i] + ((fPixelY)-(*rc.polygonPtr.polY)[i])/
					((*rc.polygonPtr.polY)[j]-(*rc.polygonPtr.polY)[i])*((*rc.polygonPtr.polX)[j]-(*rc.polygonPtr.polX)[i]))
				nodes++
			}
			j = i
		}
		i = 0
		for {
			if i < nodes-1 {
				if nodeX[i] > nodeX[i+1] {
					nodeX[i], nodeX[i+1] = nodeX[i+1], nodeX[i]
					if i != 0 {
						i--
					}
				} else {
					i++
				}
			} else {
				break
			}
		}
		//  Fill the pixels between node pairs.
		for i = 0; i < nodes; i += 2 {
			rc.drawByBrezenham(nodeX[i]+marginX, pixelY, nodeX[i+1]-marginX, pixelY, rc.PointSizeI, rc.RegionColor)
			rc.Plt.DrawLine(nodeX[i]+marginX, pixelY, nodeX[i+1]-marginX, pixelY)
		}
	}
	return
}

/*
 interpolate circle by straight lines
*/
func (rc *Render) interpolate(minpoly *float64, maxpoly *float64, st *gerbparser.State) {
	var xc, yc float64 // drawArc center coordinates in mm
	if st.QMode == gerbparser.QuadModeSingle {
		// we have to find the sign of the I and J
		fmt.Println("G74 hook")
		os.Exit(800)
	}
	xc = st.PrevCoord.GetX() + st.Coord.GetI()
	yc = st.PrevCoord.GetY() + st.Coord.GetJ()
	r := math.Hypot(st.Coord.GetI(), st.Coord.GetJ())
	rt := math.Hypot(st.Coord.GetX()-xc, st.Coord.GetY()-yc)
	dr := rt - r

	if math.Abs(dr) > rc.PointSize {
		fmt.Println("G75 diff.=", rt-r)
		panic("(rc *Render) interpolate(): Deviation more than pointSize.")
	}
	r = (r + rt) / 2

	cosFi1 := (st.PrevCoord.GetX() - xc) / r
	if cosFi1 > 1 {
		cosFi1 = 1
	} else if cosFi1 < -1 {
		cosFi1 = -1
	}

	fi1 := rad2Deg(math.Acos(cosFi1))
	if st.PrevCoord.GetY()-yc < 0 {
		fi1 = 360.0 - fi1
	}

	cosFi2 := (st.Coord.GetX() - xc) / r
	if cosFi2 > 1 {
		cosFi2 = 1
	} else if cosFi2 < -1 {
		cosFi2 = -1
	}
	fi2 := rad2Deg(math.Acos(cosFi2))
	if st.Coord.GetY()-yc < 0 {
		fi2 = 360.0 - fi2
	}

	if st.IpMode == gerbparser.IPModeCCwC {
		if fi1 > fi2 {
			fi1 = -(360.0 - fi1)
		}

		angle := fi1
		for {
			ax := r*math.Cos(deg2Rad(angle)) + xc
			ay := r*math.Sin(deg2Rad(angle)) + yc
			ay, res := rc.addToCorners(ax, ay)
			if res == true {
				if ay > *maxpoly {
					*maxpoly = ay
				}
				if ay < *minpoly {
					*minpoly = ay
				}
			}
			angle++
			if angle > fi2 {
				break
			}
		}
	} else if st.IpMode == gerbparser.IPModeCwC {
		if fi1 < fi2 {
			fi2 = -(360.0 - fi2)
		}
		angle := fi1
		for {
			ax := r*math.Cos(deg2Rad(angle)) + xc
			ay := r*math.Sin(deg2Rad(angle)) + yc
			ay, res := rc.addToCorners(ax, ay)
			if res == true {
				if ay > *maxpoly {
					*maxpoly = ay
				}
				if ay < *minpoly {
					*minpoly = ay
				}
			}
			angle--
			if angle < fi2 {
				break
			}
		}
	} else {
		panic("(rc *Render) interpolate(): Bad IpMode.")
	}

}

func (rc *Render) addToCorners(ax, ay float64) (float64, bool) {
	ax = (ax - rc.MinX) / rc.XRes
	ay = (ay - rc.MinY) / rc.YRes
	if len(*rc.polygonPtr.polX) == 0 {
		*rc.polygonPtr.polX = append(*rc.polygonPtr.polX, ax)
		*rc.polygonPtr.polY = append(*rc.polygonPtr.polY, ay)
		return ay, true
	} else {
		lastElement := len(*rc.polygonPtr.polX) - 1 // last element
		if (math.Abs(ax-(*rc.polygonPtr.polX)[lastElement]) > rc.PointSize) ||
			(math.Abs(ay-(*rc.polygonPtr.polY)[lastElement]) > rc.PointSize) {
			*rc.polygonPtr.polX = append(*rc.polygonPtr.polX, ax)
			*rc.polygonPtr.polY = append(*rc.polygonPtr.polY, ay)
			return ay, true
		}
	}
	return 0, false
}

// ################################### EOF ###############################################
