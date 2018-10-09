package gerberdatamodel

import (
	"flag"
	"github.com/akavel/polyclip-go"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strconv"
	"testing"
)

// Generalized with integer
func bresenham(x1, y1, x2, y2 int, col color.Color, img *image.NRGBA) {
	var dx, dy, e, slope int
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
		img.Set(x1, y1, col)

		// Is line an horizontal ?
	case y1 == y2:
		for ; dx != 0; dx-- {
			img.Set(x1, y1, col)
			x1++
		}
		img.Set(x1, y1, col)

		// Is line a vertical ?
	case x1 == x2:
		if y1 > y2 {
			y1, y2 = y2, y1
		}
		for ; dy != 0; dy-- {
			img.Set(x1, y1, col)
			y1++
		}
		img.Set(x1, y1, col)

		// Is line a diagonal ?
	case dx == dy:
		if y1 < y2 {
			for ; dx != 0; dx-- {
				img.Set(x1, y1, col)
				x1++
				y1++
			}
		} else {
			for ; dx != 0; dx-- {
				img.Set(x1, y1, col)
				x1++
				y1--
			}
		}
		img.Set(x1, y1, col)

		// wider than high ?
	case dx > dy:
		if y1 < y2 {
			dy, e, slope = 2*dy, dx, 2*dx
			for ; dx != 0; dx-- {
				img.Set(x1, y1, col)
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
				img.Set(x1, y1, col)
				x1++
				e -= dy
				if e < 0 {
					y1--
					e += slope
				}
			}
		}
		img.Set(x1, y1, col)

		// higher than wide.
	default:
		if y1 < y2 {
			dx, e, slope = 2*dx, dy, 2*dy
			for ; dy != 0; dy-- {
				img.Set(x1, y1, col)
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
				img.Set(x1, y1, col)
				y1--
				e -= dx
				if e < 0 {
					x1++
					e += slope
				}
			}
		}
		img.Set(x1, y1, col)
	}
	return
}

func TestMain(m *testing.M) {
	flag.Set("stderrthreshold", "ERROR")
	flag.Set("alsologtostderr", "true")
	flag.Set("logtostderr", "true")

	flag.Parse()
	os.Exit(m.Run())
}

func TestPoint_String(t *testing.T) {
	p := polyclip.Point{0.99999, -10000.00001}
	t.Log(p.String())
}

func TestGerberCircle_String(t *testing.T) {

	c := GerberCircle{polyclip.Point{0.99999, -10000.00001}, 5.0, color.RGBA{1, 1, 1, 1}}
	sc := c.String()

	cc := NewGerberCircle(polyclip.Point{0.99999, -10000.00001}, 5.0, color.RGBA{1, 1, 1, 1})
	scc := (*cc).String()

	if sc != scc {
		t.Error(sc, "\n", scc)
	} else {
		t.Log(sc)
	}

	clone := cc.Clone()
	scc = (clone).String()

	if sc != scc {
		t.Error("Clone() error! : ", sc, "\n", scc)
	} else {
		t.Log(sc)
	}

	clone.(*GerberCircle).centerPoint.X = 1.0001
	scc = (clone).String()
	if sc == scc {
		t.Error("Clone() error! : ", sc, "\n", scc)
	} else {
		t.Log(sc)
	}
}

func TestGerberCircle_Render(t *testing.T) {

	renderers := []Renderer{
		&Plotter{"Plotter 1"},
		&Plotter{"Plotter 2"},
		&Plotter{"Plotter 3"},
	}

	c := GerberCircle{polyclip.Point{0.99999, -10000.00001}, 5.0, color.RGBA{1, 1, 1, 1}}

	c.Render(&renderers)

}

func TestGerberCircle_ToPoly(t *testing.T) {
	cc := NewGerberCircle(polyclip.Point{1000.0, 1000.0}, 500.0, color.RGBA{255, 0, 0, 255})
	var poly GerberPoly
	for i := 0; i < 100000; i++ {

		poly = cc.ToPoly(10)
		_ = poly
		break
	}

	img := image.NewNRGBA(image.Rect(0, 0, 2000, 2000))
	img.Set(int(poly.originPoint.X), int(poly.originPoint.Y), poly.c)

	xc := int(poly.originPoint.X)
	yc := int(poly.originPoint.Y)
	l := len(poly.PolyCorners)
	for i := 1; i < len(poly.PolyCorners); i++ {
		img.Set(int(poly.PolyCorners[i].X)+xc, int(poly.PolyCorners[i].Y)+yc, poly.c)

		bresenham(int(poly.PolyCorners[i].X)+xc,
			int(poly.PolyCorners[i].Y)+yc,
			int(poly.PolyCorners[i-1].X)+xc,
			int(poly.PolyCorners[i-1].Y)+yc,
			poly.c, img)
	}
	bresenham(int(poly.PolyCorners[0].X)+xc,
		int(poly.PolyCorners[0].Y)+yc,
		int(poly.PolyCorners[l-1].X)+xc,
		int(poly.PolyCorners[l-1].Y)+yc,
		poly.c, img)


	f, _ := os.OpenFile("test_out.png", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	png.Encode(f, img)

}

func TestGerberCircle_ToPoly2(t *testing.T) {
	cc := NewGerberCircle(polyclip.Point{1000.0, 1000.0}, 500.0, color.RGBA{255, 0, 0, 255})
	var poly GerberPoly
	for i := 0; i < 100000; i++ {
		poly = cc.ToPoly2(10)
		_ = poly
	}
	/*
		img := image.NewNRGBA(image.Rect(0, 0, 2000, 2000))
		img.Set(int(poly.originPoint.X), int(poly.originPoint.Y), poly.c)
		xc := int(poly.originPoint.X)
		yc := int(poly.originPoint.Y)
		for i := range poly.PolyCorners {
			img.Set(int(poly.PolyCorners[i].X) + xc, int(poly.PolyCorners[i].Y) + yc, poly.c)
		}
		f, _ := os.OpenFile("test_out.png", os.O_WRONLY|os.O_CREATE, 0600)
		defer f.Close()
		png.Encode(f, img)
	*/
}

func BenchmarkGerberCircle_ToPoly(b *testing.B) {
	cc := NewGerberCircle(polyclip.Point{1000.0, 1000.0}, 500.0, color.RGBA{255, 0, 0, 255})
	for i := 0; i < 100000; i++ {
		poly := cc.ToPoly(10)
		_ = poly
	}
}

func BenchmarkGerberCircle_ToPoly2(b *testing.B) {
	cc := NewGerberCircle(polyclip.Point{1000.0, 1000.0}, 500.0, color.RGBA{255, 0, 0, 255})
	for i := 0; i < 100000; i++ {
		poly := cc.ToPoly2(10)
		_ = poly
	}
}

func TestIntersection(t *testing.T) {

	//edge1 := Edge{Point{-10.0,-10.0}, Point{10.0,10.0}}
	//edge2 := Edge{Point{5.0,-5.0}, Point{-5.0,5.0}}
	//
	//resEdge, intersect := Intersection(edge1, edge2)
	//if intersect == false {
	//	t.Error("Error: edges have 1 intersection point")
	//} else {
	//	if resEdge.point0 != resEdge.point1 {
	//		t.Error("there must be single point")
	//		t.Error(resEdge.point0.String() + resEdge.point1.String())
	//	} else {
	//		t.Log(resEdge.point0.String())
	//	}
	//}

	p1 := polyclip.Point{1.0, 1.0}
	p2 := polyclip.Point{2.0, 2.0}
	p3 := polyclip.Point{9.0, 9.0}
	p4 := polyclip.Point{10.0, 10.0}

	px2 := polyclip.Point{2.0, 1.0}

	py1 := polyclip.Point{2.0, -1.0}
	py2 := polyclip.Point{2.0, 10.0}

	py11 := polyclip.Point{2.0, -2.0}
	py22 := polyclip.Point{2.0, 9.0}

	py3 := polyclip.Point{3.0, -1.0}
	py4 := polyclip.Point{3.0, 10.0}

	type testCase struct {
		e0      Edge
		e1      Edge
		res     Edge
		resBool bool
	}

	testCases := []testCase{
		{Edge{p1, p2}, Edge{p3, p4}, Edge{}, false},
		{Edge{p2, p1}, Edge{p3, p4}, Edge{}, false},
		{Edge{p1, p2}, Edge{p4, p3}, Edge{}, false},
		{Edge{p2, p1}, Edge{p4, p3}, Edge{}, false},

		{Edge{p1, p3}, Edge{p2, p4}, Edge{p2, p3}, true},
		{Edge{p3, p1}, Edge{p2, p4}, Edge{p3, p2}, true},
		{Edge{p1, p4}, Edge{p2, p3}, Edge{p2, p3}, true},
		{Edge{p4, p1}, Edge{p2, p3}, Edge{p3, p2}, true},
		{Edge{p1, p1}, Edge{p1, p4}, Edge{p1, p1}, true},

		// lines are parallel to x axis
		{Edge{p1, px2}, Edge{p1, p4}, Edge{p1, p1}, true},
		{Edge{p1, px2}, Edge{px2, p4}, Edge{px2, px2}, true},
		{Edge{p1, px2}, Edge{p3, p4}, Edge{}, false},

		// lines are parallel to y axis
		{Edge{p1, p3}, Edge{p1, p1}, Edge{p1, p1}, true},
		{Edge{p1, p3}, Edge{p4, p4}, Edge{}, false},
		{Edge{p1, p1}, Edge{p2, p2}, Edge{}, false},
		{Edge{p1, p3}, Edge{p1, p1}, Edge{p1, p1}, true},
		{Edge{py1, py2}, Edge{py3, py4}, Edge{}, false},
		{Edge{py11, py22}, Edge{py11, py22}, Edge{py11, py22}, true},
		{Edge{py22, py11}, Edge{py11, py22}, Edge{py22, py11}, true},

		// point
		{Edge{py22, py22}, Edge{py11, py11}, Edge{}, false},
		{Edge{py22, py22}, Edge{py22, py22}, Edge{py22, py22}, true},

		// orthogonal lines
		{Edge{py11, py22}, Edge{p3, p1}, Edge{polyclip.Point{2.0, 2.0}, polyclip.Point{2.0, 2.0}}, true},
	}

	for i, testcase := range testCases {
		if i == 17 {
			t.Log("")
		}
		resEdge, intersect := Intersection(testcase.e0, testcase.e1)

		if intersect != testcase.resBool {
			t.Error("Error at step " + strconv.Itoa(i))
		} else {
			if resEdge != testcase.res {
				t.Error("Error at step " + strconv.Itoa(i))
				t.Error("Bad edge : " + resEdge.point0.String() + " - " + resEdge.point1.String())
			}
		}
	}
}

func TestGerberCircle_Union1(t *testing.T) {
	cc := NewGerberCircle(polyclip.Point{0.0, 0.0}, 200.0, color.RGBA{255, 0, 0, 255})
	cc2 := NewGerberCircle(polyclip.Point{150.0, 20.0}, 100.0, color.RGBA{255, 0, 0, 255})
	poly := GerberPoly{}
	//
	//poly = cc.Union1(cc2)


	r1 := polyclip.Polygon{cc.ToPoly(10).PolyCorners}
	r2 := polyclip.Polygon{cc2.ToPoly(10).PolyCorners}

	//r2 := polyclip.Polygon{polyclip.Contour{polyclip.Point{100.0,100.0}, polyclip.Point{200.0,200.0},
	//	polyclip.Point{500.0,400.0},},}

	_ = r2


	r3 := r1.Construct(polyclip.XOR, r2)

	_ = r3

	poly.originPoint = polyclip.Point{1000.0, 1000.0}
	poly.c = cc.c
	img := image.NewNRGBA(image.Rect(0, 0, 2000, 2000))
	img.Set(int(poly.originPoint.X), int(poly.originPoint.Y), poly.c)
	xc := int(poly.originPoint.X)
	yc := int(poly.originPoint.Y)
	fill := make(polyclip.Contour, 0)
	for i := range r3 {
		poly.PolyCorners = r3[i]
		l := len(poly.PolyCorners)
		for i := 1; i < len(poly.PolyCorners); i++ {
			img.Set(int(poly.PolyCorners[i].X)+xc, int(poly.PolyCorners[i].Y)+yc, poly.c)

			bresenham(int(poly.PolyCorners[i].X)+xc,
				int(poly.PolyCorners[i].Y)+yc,
				int(poly.PolyCorners[i-1].X)+xc,
				int(poly.PolyCorners[i-1].Y)+yc,
				poly.c, img)
		}
		bresenham(int(poly.PolyCorners[0].X)+xc,
			int(poly.PolyCorners[0].Y)+yc,
			int(poly.PolyCorners[l-1].X)+xc,
			int(poly.PolyCorners[l-1].Y)+yc,
			poly.c, img)
		for _, p := range poly.PolyCorners {
		fill = append(fill, polyclip.Point{p.X + poly.originPoint.X, p.Y + poly.originPoint.Y})
		}
		fill = append(fill, polyclip.Point{poly.PolyCorners[0].X + poly.originPoint.X, poly.PolyCorners[0].Y + poly.originPoint.Y})
	}

	renderOutline(fill, poly.c, img)

	f, _ := os.OpenFile("test_out.png", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	png.Encode(f, img)

}

// renders an outline (a.k.a. polygon).
// edges are straight lines
// coordinates are pixels of rc.Img but in float64
func renderOutline(contours polyclip.Contour, col color.Color, img *image.NRGBA) {

	numVertices := len(contours)
	minY := contours[0].Y
	maxY := contours[0].Y
	for i := range contours {
		if contours[i].Y > maxY {
			maxY = contours[i].Y
		}
		if contours[i].Y < minY {
			minY = contours[i].Y
		}
	}

	var nodes = 0
	var nodeX []int
	nodeX = make([]int, numVertices)
	var pixelY int

	// take into account real plotter pen setPoint size
	startY := int(math.Round(minY))
	stopY := int(math.Round(maxY))

	// fill the inner points of the polygon
	var i = 0

	for pixelY = startY; pixelY < stopY; pixelY++ {
		fPixelY := float64(pixelY)
		nodes = 0
		j := numVertices - 1
		for i = 0; i < numVertices; i++ {
			if ( contours[i].Y < fPixelY && contours[j].Y >= fPixelY ) ||
				(contours[j].Y < fPixelY && contours[i].Y >= fPixelY) {

				nodeX[nodes] = int(math.Round(contours[i].X + (fPixelY-contours[i].Y)/
					(contours[j].Y-contours[i].Y)*(contours[j].X-contours[i].X)))

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
			bresenham(nodeX[i], pixelY, nodeX[i+1], pixelY, color.NRGBA{0.0,255.0, 0.0, 255.0}, img)
		}
	}
	return
}
