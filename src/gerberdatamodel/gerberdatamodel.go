package gerberdatamodel

import (
	"errors"
	"fmt"
	"github.com/akavel/polyclip-go"
	"github.com/go-gl/mathgl/mgl64"
	"glog_t"
	"image/color"
	"math"
	"strconv"
)

type Renderer interface {
	Init() error
	Render()
	String() string
}

type Plotter struct {
	name string
}

func (gc *Plotter) Render() {
	glog_t.Warningln("Plotter.Render() " + gc.name + " invoked!")
}

func (gc *Plotter) String() string {
	return "Plotter.String() " + gc.name + " stub"
}

func (gc *Plotter) Init() error {
	return errors.New("Plotter.Init() " + gc.name + "stub")
}

type GObject interface {
	Clone() GObject
	Render(renderers *[]Renderer)
	Union(*GObject) GObject
	Subtract(*GObject) GObject
	And(*GObject) GObject
	Xor(*GObject) GObject
	Delete()
	String() string
}

//type polyclip.Point struct {
//	X float64
//	Y float64
//}

// gerber data model primary unit
type GerberCircle struct {
	centerPoint polyclip.Point
	radius      float64
	c           color.Color
}

func (cir *GerberCircle) String() string {
	return "GerberCircle: center " + cir.centerPoint.String() +
		", radius " + strconv.FormatFloat(cir.radius, 'f', 5, 64) +
		", color (R G B A) = " + fmt.Sprint(cir.c.RGBA())
}

func NewGerberCircle(center polyclip.Point, radius float64, color color.Color) *GerberCircle {
	retval := GerberCircle{center, radius, color}
	return &retval
}

func (cir *GerberCircle) Clone() GObject {
	retVal := new(GerberCircle)
	retVal = cir
	return retVal
}

func (self *GerberCircle) Render(renderers *[]Renderer) {
	for i := range *renderers {
		(*renderers)[i].Render()
	}
}

func (self *GerberCircle) Union(another *GObject) GObject {
	return &GerberCircle{}
}


func (cir *GerberCircle) Union1(another *GerberCircle) *GerberPoly {

	subj := make(polyclip.Polygon, 0)
	subj = append(subj, cir.ToPoly(cir.radius / 20.0).PolyCorners)

	subj2 := make(polyclip.Polygon, 0)
	subj2 = append(subj2, another.ToPoly(another.radius / 20.0).PolyCorners)

	resPoly := subj.Construct(polyclip.INTERSECTION, subj2)

	retval := GerberPoly{cir.centerPoint, resPoly[0], resPoly[0].BoundingBox(), cir.c}

	return &retval
}

func (self *GerberCircle) Subtract(another *GObject) GObject {
	return &GerberCircle{}
}

func (self *GerberCircle) And(another *GObject) GObject {
	return &GerberCircle{}
}

func (self *GerberCircle) Xor(another *GObject) GObject {
	return &GerberCircle{}
}

func (self *GerberCircle) Delete() {

}

// the polygon mus be set in clock-wise order of vertices
func (cir *GerberCircle) ToPoly(chord float64) GerberPoly {
	retVal := new(GerberPoly)
	retVal.originPoint = cir.centerPoint
	retVal.c = cir.c
	arcStep := chord / cir.radius

	// interpolate 1/8 of the circle
	angStop := mgl64.DegToRad(180.0 - 45.0)
	for angle := mgl64.DegToRad(180.0); angle > angStop; angle = angle - arcStep {
		xa := cir.radius * math.Cos(angle)
		ya := cir.radius * math.Sin(angle)
		retVal.PolyCorners = append(retVal.PolyCorners, polyclip.Point{xa, ya})
	}
	l := len(retVal.PolyCorners) - 1
	for i := l; i > 0; i-- {
		retVal.PolyCorners = append(retVal.PolyCorners, polyclip.Point{-retVal.PolyCorners[i].Y, -retVal.PolyCorners[i].X})
	}
	l = len(retVal.PolyCorners) - 1
	for i := l; i > 0; i-- {
		retVal.PolyCorners = append(retVal.PolyCorners, polyclip.Point{-retVal.PolyCorners[i].X, retVal.PolyCorners[i].Y})
	}
	l = len(retVal.PolyCorners) - 1
	for i := l; i > 0; i-- {
		retVal.PolyCorners = append(retVal.PolyCorners, polyclip.Point{retVal.PolyCorners[i].X, -retVal.PolyCorners[i].Y})
	}

	for i := range retVal.PolyCorners {
		retVal.PolyCorners[i].X = retVal.PolyCorners[i].X + retVal.originPoint.X
		retVal.PolyCorners[i].Y = retVal.PolyCorners[i].Y + retVal.originPoint.Y
	}

	retVal.boundBox = retVal.PolyCorners.BoundingBox()

	glog_t.Warningln(retVal.PolyCorners[0].String())
	glog_t.Warningln(retVal.PolyCorners[len(retVal.PolyCorners) - 1].String())

	return *retVal
}

// VERY SLOW!!!
func (cir *GerberCircle) ToPoly2(chord float64) GerberPoly {
	retVal := new(GerberPoly)
	retVal.originPoint = cir.centerPoint
	retVal.c = cir.c
	arcStep := chord / cir.radius

	// interpolate 1/8 of the circle
	angStop := mgl64.DegToRad(360.0) - arcStep/2
	for angle := 0.0; angle <= angStop; angle = angle + arcStep {
		xa := cir.radius * math.Cos(angle)
		ya := cir.radius * math.Sin(angle)
		retVal.PolyCorners = append(retVal.PolyCorners, polyclip.Point{xa, ya})
	}
	retVal.boundBox = retVal.PolyCorners.BoundingBox()
	return *retVal
}

// gerber data model primary unit
type GerberPoly struct {
	originPoint polyclip.Point
	PolyCorners polyclip.Contour
	boundBox    polyclip.Rectangle
	c           color.Color
}

func NewGerberPoly() *GerberPoly {
	return &GerberPoly{}
}

func (self *GerberPoly) Clone() GObject {
	return &GerberPoly{}
}

func (self *GerberPoly) Render(renderers *[]Renderer) {

}

func (self *GerberPoly) Union(another *GObject) GObject {

	retVal := GerberPoly{}
	retVal.originPoint = self.originPoint
	retVal.c = self.c

	return &GerberPoly{}
}

func (self *GerberPoly) Subtract(another *GObject) GObject {
	return &GerberPoly{}
}

func (self *GerberPoly) And(another *GObject) GObject {
	return &GerberPoly{}
}

func (self *GerberPoly) Xor(another *GObject) GObject {
	return &GerberPoly{}
}

func (self *GerberPoly) Delete() {

}

func (self *GerberPoly) String() string {
	return ""
}

type Edge struct {
	// the Edge is directional!
	point0 polyclip.Point
	point1 polyclip.Point
}

const float64EqualityThreshold = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

func Intersection(e0 Edge, e1 Edge) (Edge, bool) {

	Dx1 := e0.point1.X - e0.point0.X
	Dy1 := e0.point1.Y - e0.point0.Y

	Dx2 := e1.point1.X - e1.point0.X
	Dy2 := e1.point1.Y - e1.point0.Y

	var k1, k2 float64
	var y01, y02 float64

	if almostEqual(Dx1, 0.0) != true {
		k1 = Dy1 / Dx1
		y01 = e0.point0.Y - k1*e0.point0.X
	}

	if almostEqual(Dx2, 0.0) != true {
		k2 = Dy2 / Dx2
		y02 = e1.point0.Y - k2*e1.point0.X
	}

	if almostEqual(Dx1, 0.0) == true && almostEqual(Dx2, 0.0) != true {
		// line2: y2(x) = K2*x + y02
		yp := k2*e0.point1.X + y02
		if yp >= MinY(e1.point0, e1.point1).Y && yp <= MaxY(e1.point0, e1.point1).Y {
			return Edge{polyclip.Point{e0.point1.X, yp}, polyclip.Point{e0.point1.X, yp}}, true
		} else {
			return Edge{polyclip.Point{}, polyclip.Point{}}, false
		}
	}

	if almostEqual(Dx2, 0.0) == true && almostEqual(Dx1, 0.0) != true {
		// line2: y1(x) = K1*x + y01
		yp := k1*e1.point1.X + y01
		if yp >= MinY(e0.point0, e0.point1).Y && yp <= MaxY(e0.point0, e0.point1).Y {
			return Edge{polyclip.Point{e1.point1.X, yp}, polyclip.Point{e1.point1.X, yp}}, true
		} else {
			return Edge{polyclip.Point{}, polyclip.Point{}}, false
		}
	}

	if almostEqual(Dx1, 0.0) == true && almostEqual(Dx2, 0.0) == true {
		if almostEqual(e0.point0.X, e1.point0.X) == false {
			// two vertical lines with different x
			return Edge{}, false
		}
		if almostEqual(e0.point0.X, e1.point0.X) == true && almostEqual(e0.point0.Y, e1.point0.Y) == true {
			// two coincident points
			return e0, true
		}
		// two vertical lines with the same x

		if math.Min(e1.point0.Y, e1.point1.Y) > math.Max(e0.point0.Y, e0.point1.Y) ||
			math.Max(e1.point0.Y, e1.point1.Y) < math.Min(e0.point0.Y, e0.point1.Y) {
			// no common points
			return Edge{polyclip.Point{}, polyclip.Point{}}, false
		}
		p0 := polyclip.Point{}
		p1 := polyclip.Point{}

		switch {
		case Dy1 >= 0 && Dy2 >= 0:
			p0 = MaxY(e0.point0, e1.point0)
			p1 = MinY(e0.point1, e1.point1)
		case Dy1 >= 0 && Dy2 < 0:
			p0 = MaxY(e0.point0, e1.point1)
			p1 = MinY(e0.point1, e1.point0)
		case Dy1 < 0 && Dy2 < 0:
			p0 = MinY(e0.point0, e1.point0)
			p1 = MaxY(e0.point1, e1.point1)
		case Dy1 < 0 && Dy2 >= 0:
			p0 = MinY(e0.point0, e1.point1)
			p1 = MaxY(e0.point1, e1.point0)
		}
		return Edge{p0, p1}, true
	}

	if almostEqual(y01, y02) && almostEqual(k1, k2) {
		p0 := polyclip.Point{}
		p1 := polyclip.Point{}
		// both edges belongs to the one line

		if math.Min(e1.point0.X, e1.point1.X) > math.Max(e0.point0.X, e0.point1.X) ||
			math.Max(e1.point0.X, e1.point1.X) < math.Min(e0.point0.X, e0.point1.X) {
			// no common points
			return Edge{polyclip.Point{}, polyclip.Point{}}, false
		}
		switch {
		case Dx1 >= 0 && Dx2 >= 0:
			p0 = MaxX(e0.point0, e1.point0)
			p1 = MinX(e0.point1, e1.point1)
		case Dx1 >= 0 && Dx2 < 0:
			p0 = MaxX(e0.point0, e1.point1)
			p1 = MinX(e0.point1, e1.point0)
		case Dx1 < 0 && Dx2 < 0:
			p0 = MinX(e0.point0, e1.point0)
			p1 = MaxX(e0.point1, e1.point1)
		case Dx1 < 0 && Dx2 >= 0:
			p0 = MinX(e0.point0, e1.point1)
			p1 = MaxX(e0.point1, e1.point0)

		}
		return Edge{p0, p1}, true
	}

	if almostEqual(k1, k2) {
		// the edges are parallel
		return Edge{polyclip.Point{}, polyclip.Point{}}, false
	}

	// line1: y1(x) = k1*x + y01
	// line2: y2(x) = K2*x + y02
	// one intersection point

	xp := (y02 - y01) / (k1 - k2)

	if xp >= math.Min(e0.point0.X, e0.point1.X) && xp <= math.Max(e0.point0.X, e0.point1.X) &&
		xp >= math.Min(e1.point0.X, e1.point1.X) && xp <= math.Max(e1.point0.X, e1.point1.X) {
		yp := k1*xp + y01
		return Edge{polyclip.Point{xp, yp}, polyclip.Point{xp, yp}}, true
	}
	return Edge{polyclip.Point{}, polyclip.Point{}}, false
}

func MinX(p0, p1 polyclip.Point) polyclip.Point {
	if math.Min(p0.X, p1.X) == p0.X {
		return p0
	} else {
		return p1
	}
}

func MaxX(p0, p1 polyclip.Point) polyclip.Point {
	if math.Max(p0.X, p1.X) == p0.X {
		return p0
	} else {
		return p1
	}
}

func MinY(p0, p1 polyclip.Point) polyclip.Point {
	if math.Min(p0.Y, p1.Y) == p0.Y {
		return p0
	} else {
		return p1
	}
}

func MaxY(p0, p1 polyclip.Point) polyclip.Point {
	if math.Max(p0.Y, p1.Y) == p0.Y {
		return p0
	} else {
		return p1
	}
}
