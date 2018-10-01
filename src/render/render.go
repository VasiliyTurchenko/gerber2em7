package render

import (
	"configurator"
	"errors"
	"github.com/spf13/viper"
	glog "glog_t"
	"image"
	"image/color"
	"math"
	"strconv"
)

import (
	. "gerberbasetypes"
	"plotter"
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
	Plt        *plotter.PlotterParams
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
	PolygonPtr *Polygon
}

func NewRender(plotter *plotter.PlotterParams, viper *viper.Viper, minX, minY, maxX, maxY float64) *Render {
	retVal := new(Render)
	retVal.Init(plotter, viper, minX, minY, maxX, maxY)
	return retVal
}

func (rc *Render) Init(plt *plotter.PlotterParams, viper *viper.Viper, minX, minY, maxX, maxY float64) {
	// physical plotter single step size
	rc.XRes = viper.GetFloat64(configurator.CfgPlotterXRes)
	rc.YRes = viper.GetFloat64(configurator.CfgPlotterYRes)

	arr := viper.Get(configurator.CfgPlotterPenSizes)
	b, ok := arr.([]interface{})
	if ok == false {
		glog.Fatalln("penSizes configuration error")
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

	maxLimX1 := int(float64(rc.CanvasWidth) / float64(rc.XRes))
	maxLimY1 := int(float64(rc.CanvasHeight) / float64(rc.YRes))

	if rc.LimitsX1 > maxLimX1 {
		glog.Warningln("PCB size X is bigger than plotter working area! The PCB will be truncated.")
		rc.LimitsX1 = maxLimX1
	}

	if rc.LimitsY1 > maxLimY1 {
		glog.Warningln("PCB size Y is bigger than plotter working area! The PCB will be truncated.")
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
	rc.MissedColor = color.RGBA{127, 127, 255, 255}
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
	pointSize = pointSize/2 + pointSize%2
	// actually does not draw filled circle but only cirle line
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
func (rc *Render) DrawByRectangleAperture(x0, y0, x1, y1, apSizeX, apSizeY int, col color.Color) {

	var w, h, xOrigin, yOrigin int

	if x0 != x1 && y0 != y1 {
		glog.Errorln("Drawing by rectangular aperture with arbitrary angle is not supported!")
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
		rc.DrawFilledRectangle(xOrigin, yOrigin, w, h, col)
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
		rc.DrawFilledRectangle(xOrigin, yOrigin, w, h, col)
		rc.drawByBrezenham(xOrigin, yOrigin, x1, y1, rc.PointSizeI, col)
		return
	}
}

// for D01 commands
func (rc *Render) DrawByCircleAperture(x0, y0, x1, y1, apDia int, col color.Color) {
	// save x0, y0, x1, y1
	savedx0 := x0
	savedy0 := y0
	savedx1 := x1
	savedy1 := y1

	var xPen, yPen int
	ptsz := rc.PointSizeI

	rc.DrawDonut(x0, y0, apDia, 0, col)

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
		rc.DrawFilledRectangle(xOrigin, yOrigin, w, h, col)
		// move pen back to original x1, y1 setPoint
		xPen, yPen = rc.drawByBrezenham(xOrigin, yOrigin, savedx1, savedy1, ptsz, col)
		rc.DrawDonut(savedx1, savedy1, apDia, 0, col)
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
		rc.DrawFilledRectangle(xOrigin, yOrigin, w, h, col)
		// move pen back to original x1, y1 setPoint
		xPen, yPen = rc.drawByBrezenham(xOrigin, yOrigin, savedx1, savedy1, ptsz, col)
		rc.DrawDonut(savedx1, savedy1, apDia, 0, col)
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
	// and final DrawDonut
	rc.DrawDonut(savedx1, savedy1, apDia, 0, col)
	_, _ = xPen, yPen
}

// for aperture flash D03
//const strat int = 0  // closed rectangles inserted each into other
const strat int = 1 // zig-zag

// draws a filled rectangle
func (rc *Render) DrawFilledRectangle(origX, origY, w, h int, col color.Color) {

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
		glog.Fatalln("Error during filled rectangle drawing: pen did not returned to the origin setPoint!")
	}
	rc.FilledRctCounter++
}

func (rc *Render) DrawDonut(origX, origY, dia, holeDia int, col color.Color) {
	// performs DrawDonut (drawCircle) aperture flash
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

func (rc *Render) MovePen(x1, y1, x2, y2 int, col color.Color) (int, int) {
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
func (rc *Render) DrawArc(x1, y1, x2, y2, i, j float64, apertureSize int, ipm IPmode, qm QuadMode, col color.Color) error {

	var xC, yC float64

	if qm == QuadModeSingle {
		// we have to find the sign of the I and J
		glog.Fatalln("G74 hook")

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

	cosPhi1 := band((x1 - xC) / r, 1.0)
	Phi1 := rad2Deg(math.Acos(cosPhi1))
	if float64(y1)-yC < 0 {
		Phi1 = 360.0 - Phi1
	}

	cosPhi2 := band((x2 - xC) / r, 1.0)
	Phi2 := rad2Deg(math.Acos(cosPhi2))
	if float64(y2)-yC < 0 {
		Phi2 = 360.0 - Phi2
	}

	numArcs := apertureSize / rc.PointSizeI // how many arcs to do..
	r = r + (float64(apertureSize) / 2) - (float64(rc.PointSizeI) / 2)
	for i := 0; i < numArcs; i++ {
		r := r - float64(i*rc.PointSizeI)

		if ipm == IPModeCCwC {
			var ppx = 0
			var ppy = 0
			if Phi1 >= Phi2 {
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

		} else if ipm == IPModeCwC {
			var ppx = 0
			var ppy = 0
			if Phi1 <= Phi2 {
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
func (rc *Render) DrawObRound(centerX, centerY, width, height, holeDia int, color color.Color) {
	var sideDia int
	if width > height {
		sideDia = height
		rc.DrawFilledRectangle(centerX, centerY, width-sideDia, height, color)
		xd1 := centerX - (width / 2) + (sideDia / 2)
		xd2 := centerX + (width / 2) - (sideDia / 2)
		rc.DrawDonut(xd1, centerY, sideDia, holeDia, color)
		rc.DrawDonut(xd2, centerY, sideDia, holeDia, color)
	} else {
		sideDia = width
		rc.DrawFilledRectangle(centerX, centerY, width, height-sideDia, color)
		yd1 := centerY - (height / 2) + (sideDia / 2)
		yd2 := centerY + (height / 2) - (sideDia / 2)
		rc.DrawDonut(centerX, yd1, sideDia, holeDia, color)
		rc.DrawDonut(centerX, yd2, sideDia, holeDia, color)
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

/*
*********************** region (polygon) processor ***********************************
 */

type Polygon struct {
	steps       *[]*State
	polX        *[]float64
	polY        *[]float64
//	numVertices int
	id          int
}

func NewPolygon(id int) *Polygon {
	retVal := new(Polygon)
	steps := make([]*State, 0)
	retVal.steps = &steps
	polX := make([]float64, 0)
	polY := make([]float64, 0)
	retVal.polX = &polX
	retVal.polY = &polY
	retVal.id = id
	return retVal
}

func (rc *Render) AddStepToPolygon(step *State) int {
	*rc.PolygonPtr.steps = append(*rc.PolygonPtr.steps, step)
	return len(*rc.PolygonPtr.steps)
}

func (rc *Render) RenderPolygon() {
	j := 0
	for j < len(*rc.PolygonPtr.steps) {
		*rc.PolygonPtr.polX = (*rc.PolygonPtr.polX)[:0]
		*rc.PolygonPtr.polY = (*rc.PolygonPtr.polY)[:0]
		if (*rc.PolygonPtr.steps)[0].Action == OpcodeD02_MOVE {
			j++
		}
		for j < len(*rc.PolygonPtr.steps) && (*rc.PolygonPtr.steps)[j].Action != OpcodeD02_MOVE  {
			if (*rc.PolygonPtr.steps)[j].IpMode != IPModeLinear {
				rc.interpolate((*rc.PolygonPtr.steps)[j])
			} else {
				xj := ((*rc.PolygonPtr.steps)[j].Coord.GetX() - rc.MinX) / rc.XRes
				yj := ((*rc.PolygonPtr.steps)[j].Coord.GetY() - rc.MinY) / rc.YRes
				*rc.PolygonPtr.polX = append(*rc.PolygonPtr.polX, xj)
				*rc.PolygonPtr.polY = append(*rc.PolygonPtr.polY, yj)
			}
			j++
		}
		colr := rc.RegionColor
		if (*rc.PolygonPtr.steps)[0].ApTransParams.Polarity == PolTypeClear {
			glog.Errorln("Clear polarity is not supported yet.")
			colr = rc.ClearColor
		}
		rc.RenderOutline(rc.PolygonPtr.polX, rc.PolygonPtr.polY, colr)
	}
	return
}

/*
interpolate circle by straight lines
*/
func (rc *Render) interpolate(st *State) {
	var xc, yc float64 // DrawArc center coordinates in mm
	if st.QMode == QuadModeSingle {
		// we have to find the sign of the I and J
		glog.Fatalln("G74 hook")
	}
	xc = st.PrevCoord.GetX() + st.Coord.GetI()
	yc = st.PrevCoord.GetY() + st.Coord.GetJ()
	r := math.Hypot(st.Coord.GetI(), st.Coord.GetJ())
	rt := math.Hypot(st.Coord.GetX()-xc, st.Coord.GetY()-yc)
	dr := rt - r

	if math.Abs(dr) > rc.PointSize {
		glog.Fatalln("(rc *Render) interpolate(): Deviation more than pointSize.", "G75 diff.=", rt-r)
	}
	r = (r + rt) / 2

	cosFi1 := band((st.PrevCoord.GetX() - xc) / r, 1.0)

	fi1 := rad2Deg(math.Acos(cosFi1))
	if st.PrevCoord.GetY()-yc < 0 {
		fi1 = 360.0 - fi1
	}

	cosFi2 := band((st.Coord.GetX() - xc) / r, 1.0)
	fi2 := rad2Deg(math.Acos(cosFi2))
	if st.Coord.GetY()-yc < 0 {
		fi2 = 360.0 - fi2
	}

	if st.IpMode == IPModeCCwC {
		if fi1 >= fi2 {
			fi1 = -(360.0 - fi1)
		}

		angle := fi1
		for {
			ax := r*math.Cos(deg2Rad(angle)) + xc
			ay := r*math.Sin(deg2Rad(angle)) + yc
			ay, _ = rc.addToCorners(ax, ay)
			angle++
			if angle > fi2 {
				break
			}
		}
	} else if st.IpMode == IPModeCwC {
		if fi1 <= fi2 {
			fi2 = -(360.0 - fi2)
		}
		angle := fi1
		for {
			ax := r*math.Cos(deg2Rad(angle)) + xc
			ay := r*math.Sin(deg2Rad(angle)) + yc
			ay, _ = rc.addToCorners(ax, ay)
			angle--
			if angle < fi2 {
				break
			}
		}
	} else {
		glog.Fatalln("(rc *Render) interpolate(): Bad IpMode.")
	}

}

func (rc *Render) addToCorners(ax, ay float64) (float64, bool) {
	ax = (ax - rc.MinX) / rc.XRes
	ay = (ay - rc.MinY) / rc.YRes
	if len(*rc.PolygonPtr.polX) == 0 {
		*rc.PolygonPtr.polX = append(*rc.PolygonPtr.polX, ax)
		*rc.PolygonPtr.polY = append(*rc.PolygonPtr.polY, ay)
		return ay, true
	} else {
		lastElement := len(*rc.PolygonPtr.polX) - 1 // last element
		if (math.Abs(ax-(*rc.PolygonPtr.polX)[lastElement]) > rc.PointSize) ||
			(math.Abs(ay-(*rc.PolygonPtr.polY)[lastElement]) > rc.PointSize) {
			*rc.PolygonPtr.polX = append(*rc.PolygonPtr.polX, ax)
			*rc.PolygonPtr.polY = append(*rc.PolygonPtr.polY, ay)
			return ay, true
		}
	}
	return 0, false
}

/* some draw helpers */

func transformCoord(inc float64, res float64) int {
	return int(math.Round(inc / res))
}

func transformFloatCoord(inc float64, res float64) float64 {
	return inc / res
}

// renders an outline (a.k.a. polygon).
// edges are straight lines
// coordinates are pixels of rc.Img but in float64
func (rc *Render) RenderOutline(verticesX *[]float64, verticesY *[]float64, colr color.RGBA) {

	if len(*verticesX) != len(*verticesY) {
		glog.Fatalln("(rc *Render) RenderOutline() : vertices arrays lengths are different")
	}
	numVertices := len(*verticesX)
	minY := (*verticesY)[0]
	maxY := (*verticesY)[0]
	for _, y := range *verticesY {
		if y > maxY {
			maxY = y
		}
		if y < minY {
			minY = y
		}
	}

	var nodes = 0
	var nodeX []int
	nodeX = make([]int, numVertices)
	var pixelY int

	// take into account real plotter pen setPoint size
	marginX := rc.PointSize / 2
	marginXI := int(math.Round(marginX))
	startY := int(math.Round(minY + marginX))
	stopY := int(math.Round(maxY - marginX))

	// fill the inner points of the polygon
	var i = 0

	for pixelY = startY; pixelY < stopY; pixelY += rc.PointSizeI {
		fPixelY := float64(pixelY)
		nodes = 0
		j := numVertices - 1
		for i = 0; i < numVertices; i++ {
			if ((*verticesY)[i] < fPixelY && (*verticesY)[j] >= fPixelY) ||
				((*verticesY)[j] < fPixelY && (*verticesY)[i] >= fPixelY) {

				nodeX[nodes] = int(math.Round((*verticesX)[i] + (fPixelY-(*verticesY)[i])/
					((*verticesY)[j]-(*verticesY)[i])*((*verticesX)[j]-(*verticesX)[i])))

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
			rc.drawByBrezenham(nodeX[i]+marginXI, pixelY, nodeX[i+1]-marginXI, pixelY, rc.PointSizeI, colr)
			rc.Plt.DrawLine(nodeX[i]+marginXI, pixelY, nodeX[i+1]-marginXI, pixelY)
		}
	}
	return
}

// ################################### EOF ###############################################
