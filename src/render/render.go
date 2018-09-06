package render

import (
	"errors"
	"fmt"
	"gerbparser"
	"image"
	"image/color"
	"math"
	"os"
	"plotter"
	//	"regions"
)

const (
	MaxInt   = int(^uint(0) >> 1)
	MinInt   = int(-MaxInt - 1)
	MaxInt32 = int32(math.MaxInt32)
	MinInt32 = int32(math.MinInt32)
	MaxInt64 = int64(math.MaxInt64)
	MinInt64 = int64(math.MinInt64)
)

/*
 ************************** Rendering context ****************************
 */
type Render struct {
	// plotter properties
	// physical plotter single step size
	XRes         float64
	YRes         float64

	// pen width
	PenWidth     [1]float64

	// paper or pcb max dimensions
	CanvasWidth  int // paper property
	CanvasHeight int // paper property
	LimitsX0     int
	LimitsY0     int
	LimitsX1     int
	LimitsY1     int

	// point size in terms of real plotter pen points
	PointSize    float64
	PointSizeI   int
	Plt          *plotter.Plotter
	// pcb properties
	MinX float64
	MinY float64
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
}

//var rc *Render

func (rc *Render) Init(plt *plotter.Plotter /*, rpr *regions.RegionProcessor*/) {

	// physical plotter single step size
	rc.XRes = 0.025 // mm per step
	rc.YRes = 0.025 // mm per step

	//	rc.XRes = 0.25       // mm per step
	//	rc.YRes = 0.25       // mm per step

	rc.PenWidth[0] = 0.07 // mm

	// paper or pcb max dimensions
	rc.LimitsX0 = 0
	rc.LimitsY0 = 0
	rc.CanvasWidth = 297
	rc.CanvasHeight = 210
	rc.LimitsX1 = int(float64(rc.CanvasWidth) / float64(rc.XRes))
	rc.LimitsY1 = int(float64(rc.CanvasHeight) / float64(rc.YRes))

	// point size in terms of real plotter pen points
	rc.PointSize = rc.PenWidth[0] / rc.XRes
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
	return
}

func (rc *Render) SetMinXY(x, y float64) {
	rc.MinX = x
	rc.MinY = y
}

/*
func (rc *Render) PointSizeF() float64 {
	return rc.PointSize
}

func (rc *Render) PointSizeI() int {
	return int(math.Round(rc.PointSize))
}
*/

type Counters struct {
	LineBresCounter   int
	MovePenCounters   int
	MovePenDistance   float64
	CircleBresCounter int
	LineBresLen       float64
	CircleLen         float64
	FilledRctCounter  int
	ObRoundCounter    int
}

var Stat Counters

type PlotterConfig struct {
	DrawContours        bool
	DrawMoves           bool
	DrawOnlyRegionsMode bool
	//
}

var PlCfg PlotterConfig

func (plc *PlotterConfig) SetDrawContoursMode() {
	plc.DrawContours = true
}
func (plc *PlotterConfig) SetDrawSolidsMode() {
	plc.DrawContours = false
}
func (plc *PlotterConfig) SetDrawMovesMode() {
	plc.DrawMoves = true
}
func (plc *PlotterConfig) SetNotDrawMovesMode() {
	plc.DrawMoves = false
}
func (plc *PlotterConfig) SetDrawOnlyRegionsMode() {
	plc.DrawOnlyRegionsMode = true
}
func (plc *PlotterConfig) SetDrawAllMode() {
	plc.DrawOnlyRegionsMode = false
}

/*----------------------------------------------*/
// modified 07-Jun-2018
func (rc *Render) Point(x, y, ptsz int, col color.Color) {
	if ptsz < 0 {
		return
	}
	if PlCfg.DrawContours == false {
		// Draw by bresenham algorithm
		x1, y1, err := -ptsz, 0, 2-2*ptsz
		for {
			rc.Img.Set(x-x1, y+y1, col)
			rc.Img.Set(x-y1, y-x1, col)
			rc.Img.Set(x+x1, y-y1, col)
			rc.Img.Set(x+y1, y+x1, col)
			ptsz = err
			if ptsz > x1 {
				x1++
				err += x1*2 + 1
			}
			if ptsz <= y1 {
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
func (rc *Render) DrawByRectAp(x0, y0, x1, y1, apSizeX, apSizeY int, col color.Color) {

	var w, h, xOrigin, yOrigin int
//	ptsz := int(math.Round(rc.PointSize))
	ptsz := rc.PointSizeI

	if x0 != x1 && y0 != y1 {
		fmt.Println("Drawing by rectangular aperture with arbitrary angle is not supported!")
		rc.circle(x0, y0, apSizeX/2, ptsz, rc.MissedColor)
		rc.circle(x1, y1, apSizeX/2, ptsz, rc.MissedColor)
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
		rc.DrawByBrezenham(x0, y0, xOrigin, yOrigin, ptsz, col)
		rc.FilledRectangle(xOrigin, yOrigin, w, h, col)
		// draw back by pen from rectangle's origin to x1, y1
		rc.DrawByBrezenham(xOrigin, yOrigin, x1, y1, ptsz, col)
		return
	}
	if y0 == y1 { // horizontal draw
		yOrigin = y0
		xOrigin = x0 + (x1-x0)/2
		w = x1 - x0 + apSizeX
		h = apSizeY
		// draw by pen from x0,y0 to rectangle's origin
		rc.DrawByBrezenham(x0, y0, xOrigin, yOrigin, ptsz, col)
		rc.FilledRectangle(xOrigin, yOrigin, w, h, col)
		rc.DrawByBrezenham(xOrigin, yOrigin, x1, y1, ptsz, col)
		return
	}
}

// for D01 commands
func (rc *Render) DrawByCircleAp(x0, y0, x1, y1, apDia int, col color.Color) {
	// save x0, y0, x1, y1
	savedx0 := x0
	savedy0 := y0
	savedx1 := x1
	savedy1 := y1

	var xpen, ypen int
	ptsz := rc.PointSizeI

	rc.Donut(x0, y0, apDia, 0, col)

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
		xpen, ypen = rc.DrawByBrezenham(savedx0, savedy0, xOrigin, yOrigin, ptsz, col)
		w := x1 - x0
		h := apDia
		rc.FilledRectangle(xOrigin, yOrigin, w, h, col)
		// move pen back to original x1, y1 point
		xpen, ypen = rc.DrawByBrezenham(xOrigin, yOrigin, savedx1, savedy1, ptsz, col)
		rc.Donut(savedx1, savedy1, apDia, 0, col)
		_, _ = xpen, ypen
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
		xpen, ypen = rc.DrawByBrezenham(savedx0, savedy0, xOrigin, yOrigin, ptsz, col)
		rc.FilledRectangle(xOrigin, yOrigin, w, h, col)
		// move pen back to original x1, y1 point
		xpen, ypen = rc.DrawByBrezenham(xOrigin, yOrigin, savedx1, savedy1, ptsz, col)
		rc.Donut(savedx1, savedy1, apDia, 0, col)
		_, _ = xpen, ypen
		return
	}
	// non-orthogonal draw
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	l := hyp(dx, dy)
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
			// draw to start point
			xpen, ypen = rc.DrawByBrezenham(savedx0, savedy0, nx0, ny0, ptsz, col)
		}
		nx1 = int(math.Round(xv0 + dx))
		ny1 = int(math.Round(yv0 + dy))
		xpen, ypen = rc.DrawByBrezenham(nx0, ny0, nx1, ny1, ptsz, col)
		xv0 = xv0 + sdelta*dxv
		yv0 = yv0 + dyv
	}
	// draw back to saved x1, y1
	xpen, ypen = rc.DrawByBrezenham(nx1, ny1, savedx1, savedy1, ptsz, col)
	// and final donut
	rc.Donut(savedx1, savedy1, apDia, 0, col)
	_, _ = xpen, ypen
}

// for aperture flash D03
//const strat int = 0  // closed rectangles inserted each into other
const strat int = 1 // zig-zag

/*


 */
func (rc *Render) FilledRectangle(origX, origY, w, h /*, ptsz */ int, col color.Color) {

	xpen := origX // real pen position
	ypen := origY // real pen position
	ptsz := rc.PointSizeI

	// performs rectangle aperture flash
	x0 := origX - (w / 2)
	y0 := origY - (h / 2)
	x1 := origX + (w / 2)
	y1 := origY + (h / 2)

	if PlCfg.DrawContours == true {
		rc.DrawByBrezenham(x0, y0, x1, y0, 1, rc.ContourColor)
		rc.DrawByBrezenham(x1, y0, x1, y1, 1, rc.ContourColor)
		rc.DrawByBrezenham(x1, y1, x0, y1, 1, rc.ContourColor)
		rc.DrawByBrezenham(x0, y1, x0, y0, 1, rc.ContourColor)
	}
	x0 = x0 + (ptsz / 2)
	y0 = y0 + (ptsz / 2)

	x1 = x1 - (ptsz / 2)
	y1 = y1 - (ptsz / 2)

	// imitate pen moving to the start point
	rc.DrawByBrezenham(origX, origY, x0, y0, ptsz, col)

	// draw contour
	xpen, ypen = rc.DrawByBrezenham(x0, y0, x1, y0, ptsz, col)
	xpen, ypen = rc.DrawByBrezenham(x1, y0, x1, y1, ptsz, col)
	xpen, ypen = rc.DrawByBrezenham(x1, y1, x0, y1, ptsz, col)
	xpen, ypen = rc.DrawByBrezenham(x0, y1, x0, y0, ptsz, col)

	xp := x0
	yp := y0

	x0 = x0 + ptsz
	y0 = y0 + ptsz
	x1 = x1 - ptsz
	y1 = y1 - ptsz

	rc.DrawByBrezenham(xp, yp, x0, y0, ptsz, col)

	if strat == 0 {
		for {
			xpen, ypen = rc.DrawByBrezenham(x0, y0, x1, y0, ptsz, col)
			xpen, ypen = rc.DrawByBrezenham(x1, y0, x1, y1, ptsz, col)
			xpen, ypen = rc.DrawByBrezenham(x1, y1, x0, y1, ptsz, col)
			xpen, ypen = rc.DrawByBrezenham(x0, y1, x0, y0, ptsz, col)

			x0 = x0 + ptsz
			x1 = x1 - ptsz
			y0 = y0 + ptsz
			y1 = y1 - ptsz

			if ((x1 - x0) < 0) || ((y1 - y0) < 0) {
				break
			}
		}
	}
	if strat == 1 {
		if w > h {
			var tmpy int
			var retx int
			for {
				xpen, ypen = rc.DrawByBrezenham(x0, y0, x1, y0, ptsz, col)
				tmpy = y0
				y0 = y0 + ptsz
				if y0 > y1 {
					retx = x1
					break
				}
				xpen, ypen = rc.DrawByBrezenham(x1, tmpy, x1, y0, ptsz, col)
				xpen, ypen = rc.DrawByBrezenham(x1, y0, x0, y0, ptsz, col)
				tmpy = y0
				y0 = y0 + ptsz
				if y0 > y1 {
					retx = x0
					break
				}
				xpen, ypen = rc.DrawByBrezenham(x0, tmpy, x0, y0, ptsz, col)
			}
			// imitate pen moving to the origin point
			xpen, ypen = rc.DrawByBrezenham(retx, tmpy, origX, origY, ptsz, col)
		} else {
			var tmpx int
			var rety int
			for {
				xpen, ypen = rc.DrawByBrezenham(x0, y0, x0, y1, ptsz, col)
				tmpx = x0
				x0 = x0 + ptsz
				if x0 > x1 {
					rety = y1
					break
				}
				xpen, ypen = rc.DrawByBrezenham(tmpx, y1, x0, y1, ptsz, col)
				xpen, ypen = rc.DrawByBrezenham(x0, y1, x0, y0, ptsz, col)
				tmpx = x0
				x0 = x0 + ptsz
				if x0 > x1 {
					rety = y0
					break
				}
				xpen, ypen = rc.DrawByBrezenham(tmpx, y0, x0, y0, ptsz, col)
			}
			// imitate pen moving to the origin point
			xpen, ypen = rc.DrawByBrezenham(tmpx, rety, origX, origY, ptsz, col)
		}
	}
	if xpen != origX || ypen != origY {
		fmt.Println("Error during filled rectangle drawing: pen does not return to the origin point!")
		os.Exit(700)
	}
	Stat.FilledRctCounter++
}

func (rc *Render) Donut(origX, origY, dia, holeDia int, col color.Color) {
	// performs donut (circle) aperture flash
	ptsz := rc.PointSizeI
	radius := dia / 2
	holeRadius := holeDia / 2
	if PlCfg.DrawContours == true {
		rc.circle(origX, origY, radius, 1, rc.ContourColor)
		if holeDia > 0 {
			rc.circle(origX, origY, holeRadius, 1, rc.ContourColor)
		}
	}
	radius = radius - (ptsz / 2)
	for {
		rc.circle(origX, origY, radius, ptsz, col)
		radius = radius - ptsz
		if radius < holeRadius+(ptsz/2) {
			break
		}
	}
}

// Circle plots a circle with center x, y and radius r.
// Limiting behavior:
// r < 0 plots no pixels.
// r = 0 plots a single pixel at x, y.
// r = 1 plots four pixels in a diamond shape around the center pixel at x, y.
func (rc *Render) circle(x, y, r, ptsz int, col color.Color) {
	if r < 0 {
		return
	}
	// statistics
	Stat.CircleBresCounter++
	Stat.CircleLen += 2 * math.Pi * float64(r)

	rc.Plt.Circle(x, y, r)

	// Draw By bresenham algorithm
	x1, y1, err := -r, 0, 2-2*r
	for {
		rc.Point(x-x1, y+y1, ptsz, col)
		rc.Point(x-y1, y-x1, ptsz, col)
		rc.Point(x+x1, y-y1, ptsz, col)
		rc.Point(x+y1, y+x1, ptsz, col)
		/*
			img.Set(x-x1, y+y1, col)
			img.Set(x-y1, y-x1, col)
			img.Set(x+x1, y-y1, col)
			img.Set(x+y1, y+x1, col)
		*/
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

func (rc *Render) MovePen(x1, y1, x2, y2, ptsz int, col color.Color) (int, int) {
	Stat.MovePenCounters++
	Stat.MovePenDistance += hyp(float64(x2-x1), float64(y2-y1))
	newX := x2
	newY := y2
	if PlCfg.DrawMoves == true {
		newX, newY = rc.bresenham(x1, y1, x2, y2, ptsz, col)
	}
	rc.Plt.MoveTo(x2, y2)
	return newX, newY
}

func (rc *Render) DrawByBrezenham(x1, y1, x2, y2, ptsz int, col color.Color) (int, int) {
	// stastistics
	Stat.LineBresCounter++
	Stat.LineBresLen += hyp(float64(x2-x1), float64(y2-y1))
	newx, newy := rc.bresenham(x1, y1, x2, y2, ptsz, col)
	rc.Plt.DrawLine(x1, y1, x2, y2)
	return newx, newy
}

// Generalized with integer
func (rc *Render) bresenham(x1, y1, x2, y2, ptsz int, col color.Color) (int, int) {
	var dx, dy, e, slope int
	newx := x2
	newy := y2
	// Because drawing p1 -> p2 is equivalent to draw p2 -> p1,
	// I sort points in x-axis order to handle only half of possible cases.
	if x1 > x2 {
		x1, y1, x2, y2 = x2, y2, x1, y1
	}

	dx, dy = x2-x1, y2-y1
	// Because point is x-axis ordered, dx cannot be negative
	if dy < 0 {
		dy = -dy
	}

	switch {

	// Is line a point ?
	case x1 == x2 && y1 == y2:
		rc.Point(x1, y1, ptsz, col)

		// Is line an horizontal ?
	case y1 == y2:
		for ; dx != 0; dx-- {
			rc.Point(x1, y1, ptsz, col)
			x1++
		}
		rc.Point(x1, y1, ptsz, col)

		// Is line a vertical ?
	case x1 == x2:
		if y1 > y2 {
			y1, y2 = y2, y1
		}
		for ; dy != 0; dy-- {
			rc.Point(x1, y1, ptsz, col)
			y1++
		}
		rc.Point(x1, y1, ptsz, col)

		// Is line a diagonal ?
	case dx == dy:
		if y1 < y2 {
			for ; dx != 0; dx-- {
				rc.Point(x1, y1, ptsz, col)
				x1++
				y1++
			}
		} else {
			for ; dx != 0; dx-- {
				rc.Point(x1, y1, ptsz, col)
				x1++
				y1--
			}
		}
		rc.Point(x1, y1, ptsz, col)

		// wider than high ?
	case dx > dy:
		if y1 < y2 {
			dy, e, slope = 2*dy, dx, 2*dx
			for ; dx != 0; dx-- {
				rc.Point(x1, y1, ptsz, col)
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
				rc.Point(x1, y1, ptsz, col)
				x1++
				e -= dy
				if e < 0 {
					y1--
					e += slope
				}
			}
		}
		rc.Point(x2, y2, ptsz, col)

		// higher than wide.
	default:
		if y1 < y2 {
			dx, e, slope = 2*dx, dy, 2*dy
			for ; dy != 0; dy-- {
				rc.Point(x1, y1, ptsz, col)
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
				rc.Point(x1, y1, ptsz, col)
				y1--
				e -= dx
				if e < 0 {
					x1++
					e += slope
				}
			}
		}
		rc.Point(x2, y2, ptsz, col)
	}
	return newx, newy
}

// ARC functions
func (rc *Render) Arc(x1, y1, x2, y2, i, j float64, aps int /* ptsz int, */ , ipm gerbparser.IPmode, qm gerbparser.Quadmode, col color.Color) bool {
	//	var signI, signJ int = 0, 0
	var xc, yc float64
	ptsz := rc.PointSizeI

	if qm == gerbparser.QuadmodeSingle {
		// we have to find the sign of the I and J
		fmt.Println("G74 hook")
		return false
	}
	if PlCfg.DrawContours == true {
		rc.Point(int(x1), int(y1), 1, rc.ContourColor)
		rc.Point(int(x2), int(y2), 1, rc.ContourColor)
	}
	xc = x1 + i
	yc = y1 + j
	r := hyp(i, j)
	rt := hyp(x2-xc, y2-yc)

	dr := rt - r
	if math.Abs(dr) > float64(ptsz) {
		fmt.Println("G75 diff.=", rt-r)
		fmt.Println(x1, y1, x2, y2, i, j)
		return true
	}

	r = (r + rt) / 2

	cosfi1 := (x1 - xc) / r
	if cosfi1 > 1 {
		cosfi1 = 1
	} else if cosfi1 < -1 {
		cosfi1 = -1
	}

	fi1 := Rad2Deg(math.Acos(cosfi1))
	if float64(y1)-yc < 0 {
		fi1 = 360.0 - fi1
	}

	cosfi2 := (x2 - xc) / r
	if cosfi2 > 1 {
		cosfi2 = 1
	} else if cosfi2 < -1 {
		cosfi2 = -1
	}
	fi2 := Rad2Deg(math.Acos(cosfi2))
	if float64(y2)-yc < 0 {
		fi2 = 360.0 - fi2
	}

	nr := aps / ptsz // how many arcs to do..
	r = r + (float64(aps) / 2) - (float64(ptsz) / 2)
	for i := 0; i < nr; i++ {
		r := r - float64(i*ptsz)
		if ipm == gerbparser.IPModeCCwC {
			var ppx = 0
			var ppy = 0
			if fi1 > fi2 {
				fi1 = -(360.0 - fi1)
			}
			plx1 := int(math.Round(x1))
			plx2 := int(math.Round(x2))
			ply1 := int(math.Round(y1))
			ply2 := int(math.Round(y2))
			plr := int(math.Round(r))
			plfi1 := int(math.Round(fi1))
			plfi2 := int(math.Round(fi2))

			rc.Plt.Arc(plx1, ply1, plx2, ply2, plr, plfi1, plfi2, ipm)

			angle := fi1
			for {
				ax := int(math.Round(r*math.Cos(Deg2Rad(angle)) + xc))
				ay := int(math.Round(r*math.Sin(Deg2Rad(angle)) + yc))
				if ppx != ax || ppy != ay {
					//					img.Set(ax, ay, col)
					rc.circle(ax, ay, ptsz, 1, col)
				}
				//			fmt.Println(angle, ax, ay)
				angle++
				if angle > fi2 {
					break
				}
				ppx = ax
				ppy = ay
			}
			//		fmt.Println( fi1, "ccw ", fi2)

		} else if ipm == gerbparser.IPModeCwC {
			var ppx = 0
			var ppy = 0
			if fi1 < fi2 {
				fi2 = -(360.0 - fi2)
			}

			plx1 := int(math.Round(x1))
			plx2 := int(math.Round(x2))
			ply1 := int(math.Round(y1))
			ply2 := int(math.Round(y2))
			plr := int(math.Round(r))
			plfi1 := int(math.Round(fi1))
			plfi2 := int(math.Round(fi2))

			rc.Plt.Arc(plx1, ply1, plx2, ply2, plr, plfi1, plfi2, ipm)

			angle := fi1
			for {
				ax := int(math.Round(r*math.Cos(Deg2Rad(angle)) + xc))
				ay := int(math.Round(r*math.Sin(Deg2Rad(angle)) + yc))
				if ppx != ax || ppy != ay {
					//					img.Set(ax, ay, col)
					rc.circle(ax, ay, ptsz, 1, col)
				}
				//			fmt.Println(angle, ax, ay)
				angle--
				if angle < fi2 {
					break
				}
				ppx = ax
				ppy = ay
			}
			//		fmt.Println(fi1, "cw ", fi2)
		}
	}

	return false
}

// obround aperture flash
func (rc *Render) ObRound(centerX, centerY, width, height, holeDia int, color color.Color) {
	var sideDia int
	if width > height {
		sideDia = height
		rc.FilledRectangle(centerX, centerY, width-sideDia, height, color)
		xd1 := centerX - (width / 2) + (sideDia / 2)
		xd2 := centerX + (width / 2) - (sideDia / 2)
		rc.Donut(xd1, centerY, sideDia, holeDia, color)
		rc.Donut(xd2, centerY, sideDia, holeDia, color)
	} else {
		sideDia = width
		rc.FilledRectangle(centerX, centerY, width, height-sideDia, color)
		yd1 := centerY - (height / 2) + (sideDia / 2)
		yd2 := centerY + (height / 2) - (sideDia / 2)
		rc.Donut(centerX, yd1, sideDia, holeDia, color)
		rc.Donut(centerX, yd2, sideDia, holeDia, color)

		/*
			fmt.Println(">>>>>>>>>>>>>>>>>")
			fmt.Println(centerX, centerY)
			fmt.Println(centerX, yd1, sideDia, holeDia)
			fmt.Println(centerX, yd2, sideDia, holeDia)
		*/
	}
	Stat.ObRoundCounter++
}

//
func quad(a float64) float64 {
	return a * a
}

//
func hyp(a, b float64) float64 {
	return math.Sqrt(quad(a) + quad(b))
}

//
func Rad2Deg(a float64) float64 {
	return 360 * a / (2 * math.Pi)
}

//
func Deg2Rad(a float64) float64 {
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
func (rc *Render) StepProcessor(stepData *gerbparser.State) {
	Xc := transformCoord(stepData.Coord.GetX()-rc.MinX, rc.XRes)
	Yc := transformCoord(stepData.Coord.GetY()-rc.MinY, rc.YRes)
	var (
		Xp, Yp   int
		fXp, fYp float64
	)
	if stepData.PrevCoord == nil {
		Xp = transformCoord(0-rc.MinX, rc.XRes)
		Yp = transformCoord(0-rc.MinY, rc.YRes)
		fXp = transformFloatCoord(0-rc.MinX, rc.XRes)
		fYp = transformFloatCoord(0-rc.MinY, rc.YRes)
	} else {
		Xp = transformCoord(stepData.PrevCoord.GetX()-rc.MinX, rc.XRes)
		Yp = transformCoord(stepData.PrevCoord.GetY()-rc.MinY, rc.YRes)
		fXp = transformFloatCoord(stepData.PrevCoord.GetX()-rc.MinX, rc.XRes)
		fYp = transformFloatCoord(stepData.PrevCoord.GetY()-rc.MinY, rc.YRes)

	}

	fXc := transformFloatCoord(stepData.Coord.GetX()-rc.MinX, rc.XRes)
	fYc := transformFloatCoord(stepData.Coord.GetY()-rc.MinY, rc.YRes)

	fI := transformFloatCoord(stepData.Coord.GetI(), rc.XRes)
	fJ := transformFloatCoord(stepData.Coord.GetJ(), rc.YRes)

	if stepData.Region != nil {
		// process region
		if rc.ProcessingRegion == false {
			rc.BeginRegion()
		}
		if rc.AddStepToRegion(stepData) == stepData.Region.GetNumXY() {
			// we can process region
			rc.RenderPoly()
			rc.EndRegion()
		}
	} else {
		var stepcolor color.RGBA
		switch stepData.Action {
		case gerbparser.OpcodeD01: // draw
			if stepData.Polarity == gerbparser.PoltypeDark {
				stepcolor = rc.LineColor
			} else {
				stepcolor = rc.ClearColor
			}
			var aps int
			_ = aps
			if abs(Xc-Xp) < (4*rc.PointSizeI) && abs(Yc-Yp) < (4*rc.PointSizeI) {
				stepData.IpMode = gerbparser.IPModeLinear
			}
			if stepData.IpMode == gerbparser.IPModeLinear {
				// linear interpolation
				if PlCfg.DrawOnlyRegionsMode != true {
					if stepData.CurrentAp.Type == gerbparser.AptypeCircle {
						aps = transformCoord(stepData.CurrentAp.Diameter, rc.XRes)
						//_ = aps
						rc.DrawByCircleAp(Xp, Yp, Xc, Yc, aps, stepcolor)
					} else if stepData.CurrentAp.Type == gerbparser.AptypeRectangle {
						// draw with rectangle aperture
						w := transformCoord(stepData.CurrentAp.XSize, rc.XRes)
						h := transformCoord(stepData.CurrentAp.YSize, rc.YRes)
						rc.DrawByRectAp(Xp, Yp, Xc, Yc, w, h, stepcolor)
					} else {
						fmt.Println("Error. Only solid circle and solid rectangle may be used to draw.")
						break
					}
				}
			} else {
				// non-linear interpolation
				if PlCfg.DrawOnlyRegionsMode != true {
					if stepData.CurrentAp.Type == gerbparser.AptypeCircle {
						aps = transformCoord(stepData.CurrentAp.Diameter, rc.XRes)
						// Arcs require floats!
						if rc.Arc(fXp, fYp, fXc, fYc, fI, fJ, aps, stepData.IpMode, stepData.QMode, rc.RegionColor) == true {
							fmt.Println(stepData)
							fmt.Println(stepData.Coord)
							os.Exit(998)
						}
						rc.Donut(Xp, Yp, aps, 0, stepcolor)
						rc.Donut(Xc, Yc, aps, 0, stepcolor)
					} else if stepData.CurrentAp.Type == gerbparser.AptypeRectangle {
						fmt.Println("Arc drawing by rectangle aperture is not supported now.")
					} else {
						fmt.Println("Error. Only solid circle and solid rectangle may be used to draw.")
						break
					}
				}
			}
			//
		case gerbparser.OpcodeD02: // move
			rc.MovePen(Xp, Yp, Xc, Yc, 1, rc.MovePenColor)
		case gerbparser.OpcodeD03: // flash
			if PlCfg.DrawOnlyRegionsMode != true {
				rc.MovePen(Xp, Yp, Xc, Yc, 1, rc.MovePenColor)
				if stepData.Polarity == gerbparser.PoltypeDark {
					stepcolor = rc.ApColor
				} else {
					stepcolor = rc.ClearColor
				}
				w := transformCoord(stepData.CurrentAp.XSize, rc.XRes)
				h := transformCoord(stepData.CurrentAp.YSize, rc.YRes)
				d := transformCoord(stepData.CurrentAp.Diameter, rc.XRes)
				hd := transformCoord(stepData.CurrentAp.HoleDiameter, rc.XRes)

				switch stepData.CurrentAp.Type {
				case gerbparser.AptypeRectangle:
					rc.FilledRectangle(Xc, Yc, w, h, stepcolor)
				case gerbparser.AptypeCircle:
					rc.Donut(Xc, Yc, d, hd, stepcolor)
				case gerbparser.AptypeObround:
					if w == h {
						rc.Donut(Xc, Yc, w, hd, stepcolor)
					} else {
						rc.ObRound(Xc, Yc, w, h, 0, rc.ObRoundColor)
					}
				case gerbparser.AptypePoly:
					rc.Donut(Xc, Yc, d, hd, rc.MissedColor)
					fmt.Println("Polygonal apertures ain't supported.")
				default:
					checkError(errors.New("bad aperture type found"), 501)
					break
				}
			}
		default:
			checkError(errors.New("internal error. Bad opcode"), 666)
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

func checkError(err error, exitcode int) {
	if err != nil {
		fmt.Println(err)
		os.Exit(exitcode)
	}
}

/*
*********************** region processor ***********************************
 */

var curreg []gerbparser.State
var polX []float64
var polY []float64
var polyCorners int

func (rc *Render) BeginRegion() {
	rc.ProcessingRegion = true
	curreg = make([]gerbparser.State, 0)
}

func (rc *Render) EndRegion() {
	rc.ProcessingRegion = false
}

func (rc *Render) AddStepToRegion(st *gerbparser.State) int {
	curreg = append(curreg, *st)
	return len(curreg)
}

func (rc *Render) RenderPoly() {

	if curreg[0].Action == gerbparser.OpcodeD02 {
		curreg = curreg[1:]
	}
	xxp := curreg[0].PrevCoord.GetX()
	yyp := curreg[0].PrevCoord.GetY()
	var i int
	for i = 0; i < len(curreg); i++ {
		if float64eq1mil(curreg[i].Coord.GetX(), xxp) && float64eq1mil(curreg[i].Coord.GetY(), yyp) {
			fmt.Println("Closed segment found with  ", i, "vertexes")
			if i < len(curreg)-2 {
				fmt.Println("more than one segment in the region!")
			}
			break
		}
		if i == len(curreg)-1 {
			// the segment is not closed!
			fmt.Println(xxp, yyp)
			curreg[0].Coord.Print()
			curreg[len(curreg)-2].Coord.Print()
			curreg[len(curreg)-1].Coord.Print()
			os.Exit(1000)
		}
	}
	// let's create a array of nodes (vertices)
	polX = make([]float64, 0)
	polY = make([]float64, 0)
	polyCorners = len(curreg)
	minpolY := 100000000.0
	maxpolY := 0.0
	for j := 0; j < polyCorners; j++ {
		if curreg[j].IpMode != gerbparser.IPModeLinear {
			rc.interpolate(&minpolY, &maxpolY, &curreg[j])
		} else {
			xj := (curreg[j].Coord.GetX() - rc.MinX) / rc.XRes
			yj := (curreg[j].Coord.GetY() - rc.MinY) / rc.YRes
			if yj < minpolY {
				minpolY = yj
			}
			if yj > maxpolY {
				maxpolY = yj
			}
			polX = append(polX, xj)
			polY = append(polY, yj)
		}
	}
	polyCorners = len(polX)
	var nodes = 0
	var nodeX []int
	nodeX = make([]int, polyCorners)
	var pixelY int

	// take into account real plotter pen point size
	ystart := int(math.Round(minpolY + rc.PointSize/2))
	ystop := int(math.Round(maxpolY - rc.PointSize/2))
	xmargin := int(math.Round(rc.PointSize / 2))
	psz := rc.PointSizeI
	//	psz := int(math.Round(rp.rc.PointSize))

	for pixelY = ystart; pixelY < ystop; /*pixelY++*/ pixelY += psz {
		fPixelY := float64(pixelY)
		nodes = 0
		j := polyCorners - 1
		for i = 0; i < polyCorners; i++ {
			if (polY[i] < fPixelY && polY[j] >= fPixelY) ||
				(polY[j] < fPixelY && polY[i] >= fPixelY) {
				nodeX[nodes] = int(polX[i] + ((fPixelY)-polY[i])/(polY[j]-polY[i])*(polX[j]-polX[i]))
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
			rc.DrawByBrezenham(nodeX[i]+xmargin, pixelY, nodeX[i+1]-xmargin, pixelY, psz, rc.RegionColor)
			rc.Plt.DrawLine(nodeX[i]+xmargin, pixelY, nodeX[i+1]-xmargin, pixelY)
		}
	}
	return
}

func float64eq1mil(a, b float64) bool {
	if math.Abs(a-b) < 0.001 {
		return true
	} else {
		return false
	}
}

/*
 interpolate circle by straight lines
*/
func (rc *Render) interpolate(minpoly *float64, maxpoly *float64, st *gerbparser.State) {
	var xc, yc float64 // arc center coordinates in mm
	if st.QMode == gerbparser.QuadmodeSingle {
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
		panic("func interpolate in the package regions. Deviation more than pointsize.")
	}
	r = (r + rt) / 2

	cosfi1 := (st.PrevCoord.GetX() - xc) / r
	if cosfi1 > 1 {
		cosfi1 = 1
	} else if cosfi1 < -1 {
		cosfi1 = -1
	}

	fi1 := Rad2Deg(math.Acos(cosfi1))
	if st.PrevCoord.GetY()-yc < 0 {
		fi1 = 360.0 - fi1
	}

	cosfi2 := (st.Coord.GetX() - xc) / r
	if cosfi2 > 1 {
		cosfi2 = 1
	} else if cosfi2 < -1 {
		cosfi2 = -1
	}
	fi2 := Rad2Deg(math.Acos(cosfi2))
	if st.Coord.GetY()-yc < 0 {
		fi2 = 360.0 - fi2
	}

	if st.IpMode == gerbparser.IPModeCCwC {
		if fi1 > fi2 {
			fi1 = -(360.0 - fi1)
		}

		angle := fi1
		for {
			ax := r*math.Cos(Deg2Rad(angle)) + xc
			ay := r*math.Sin(Deg2Rad(angle)) + yc
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
			ax := r*math.Cos(Deg2Rad(angle)) + xc
			ay := r*math.Sin(Deg2Rad(angle)) + yc
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
		panic("func interpolate in the package regions. Bad IpMode.")
	}

}

func (rc *Render) addToCorners(ax, ay float64) (float64, bool) {
	ax = (ax - rc.MinX) / rc.XRes
	ay = (ay - rc.MinY) / rc.YRes
	if len(polX) == 0 {
		polX = append(polX, ax)
		polY = append(polY, ay)
		return ay, true
	} else {
		le := len(polX) - 1 // last element
		if (math.Abs(ax-polX[le]) > rc.PointSize) || (math.Abs(ay-polY[le]) > rc.PointSize) {
			polX = append(polX, ax)
			polY = append(polY, ay)
			return ay, true
		}
	}
	return 0, false
}

// ################################### EOF ###############################################
