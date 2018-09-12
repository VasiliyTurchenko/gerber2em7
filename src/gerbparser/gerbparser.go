/*
Package gerbparser contains functions and structures used for parsing gerber x2 file
*/
package gerbparser

import "fmt"
import "strings"
import "strconv"
import (
	"errors"
	"math"
	"os"
)

const GerberFormatSpec string = "%FSLA"
const GerberMOIN string = "%MOIN*%"
const GerberMOMM string = "%MOMM*%"

const InchesToMM float64 = 25.4

// Function checks against non-number characters in the string
func isNumString(ins string) bool {
	v := []byte(ins)
	for _, c := range v {
		if (c < 0x30) || (c > 0x39) {
			return false
		}
	}
	return true
}

/*
############################ format specification #####################
*/

// Format specification object
type FormatSpec struct {
	Head     string
	MUString string
	XI       int // digits in the integer part
	XD       int // digits in the fractional part
	YI       int
	YD       int
	MU       float64
}

// false - unable to parse format string
// true - parsing was successfull
func (fs *FormatSpec) Init(ins, mu string) bool {
	var result bool = false
	var err error
	fs.XI = 0
	fs.XD = 0
	fs.YI = 0
	fs.YD = 0
	fs.Head = strings.ToUpper(ins)
	fs.MUString = strings.ToUpper(mu)

	var tmpxi, tmpxd, tmpyi, tmpyd int // temporary values
	var Xpos, Ypos, suffpos int        // delimiters postions

	if strings.Compare(mu, GerberMOIN) == 0 {
		fs.MU = InchesToMM
	} else if strings.Compare(mu, GerberMOMM) == 0 {
		fs.MU = 1.0
	} else {
		goto fExit
	}

	if (strings.HasPrefix(fs.Head, GerberFormatSpec)) && (strings.HasSuffix(fs.Head, "*%")) {
		Xpos = strings.IndexByte(fs.Head, 'X')
		Ypos = strings.LastIndexByte(fs.Head, 'Y')
		suffpos = strings.LastIndexByte(fs.Head, '*')

		if (Xpos != -1) && (Ypos != -1) {
			if (Xpos < Ypos) && (Ypos < suffpos) {
				if tmpxi, err = strconv.Atoi(fs.Head[Xpos+1 : Xpos+2]); err != nil {
					goto fExit
				}
				if tmpxd, err = strconv.Atoi(fs.Head[Xpos+2 : Ypos]); err != nil {
					goto fExit
				}
				if tmpyi, err = strconv.Atoi(fs.Head[Ypos+1 : Ypos+2]); err != nil {
					goto fExit
				}
				if tmpyd, err = strconv.Atoi(fs.Head[Ypos+2 : suffpos]); err != nil {
					goto fExit
				}
				if (tmpxi != tmpyi) || (tmpxd != tmpyd) {
					goto fExit
				}
				// 4.1.1 gerber format conformance test
				if tmpxi > 6 {
					goto fExit
				}
				if (tmpxd > 7) || (tmpxd < 3) {
					goto fExit
				}
				fs.XI = tmpxi
				fs.XD = tmpxd
				fs.YI = tmpyi
				fs.YD = tmpyd
				result = true
			}
		}

	}
fExit:
	return result
}

func (fs *FormatSpec) ReadXI() int {
	return fs.XI
}
func (fs *FormatSpec) ReadXD() int {
	return fs.XD
}
func (fs *FormatSpec) ReadYI() int {
	return fs.YI
}
func (fs *FormatSpec) ReadYD() int {
	return fs.YD
}
func (fs *FormatSpec) ReadMU() float64 {
	return fs.MU
}

/*
######################### coordinates #########################################
*/
/*
 Coordinates base type
*/
type axisPoint struct {
	valFloat float64
}

func (ap *axisPoint) clear() {
	ap.valFloat = 0.0
}

// initializes the point on the axis ax
// n is the number of places for int part
// m is the number of places for frac part
// s is the scale factor 1.0 or 25.4 (mm/inches)
func (ap *axisPoint) init(ins string, n, m int, s float64) bool {
	var result = false
	var neg = false
	var ws string

	if strings.HasPrefix(ins, "-") {
		neg = true
		ws = strings.TrimPrefix(ins, "-")
	} else {
		if strings.HasPrefix(ins, "+") {
			ws = strings.TrimPrefix(ins, "+")
		} else {
			ws = ins
		}
	}
	if len(ws) > (n + m) {
		return result
	}
	if isNumString(ws) == false {
		return result
	}
	//
	//	fmt.Println("\taxisPoint.init: input values are ", ins, n, m)
	//
	ps := make([]byte, n+m)
	var inso = len(ps) - len(ws)
	for i := 0; i < inso; i++ {
		ps[i] = '0'
	}
	for i := inso; i < len(ps); i++ {
		ps[i] = (byte)(ws[i-inso])
	}

	//	fmt.Println("\taxisPoint.init: ps ", (string)(ps))

	var ipart int
	var fpart int
	var err error

	if ipart, err = strconv.Atoi((string)(ps[0:n])); err != nil {
		return false
	}

	if fpart, err = strconv.Atoi((string)(ps[n : m+n])); err != nil {
		return false
	}
	//	fmt.Println("\taxisPoint.init: int part = ", ap.valI, "; fractional part = ", ap.valF)
	tmpfloat := float64(fpart) / math.Pow10(m)
	tmpfloat += float64(ipart)
	if neg {
		tmpfloat *= -1.0
	}
	ap.valFloat = tmpfloat * float64(s)
	//	fmt.Println("\taxisPoint.init: float value =", ap.valFloat, "\n")
	result = true
	//fExit:
	return result
}

// returns axis point as float64 value
func (ap *axisPoint) getfval() float64 {
	return ap.valFloat
}

type XY struct {
	nodeNumber  uint32
	coordString string // string representation
	x           axisPoint
	y           axisPoint
	// offsets
	i axisPoint
	j axisPoint
}

func NewXY() *XY {
	retVal := new(XY)
	retVal.SetX(0)
	retVal.SetY(0)
	return retVal
}

func (xy *XY) Print() {
	fmt.Println("Node #: ", xy.nodeNumber)
	fmt.Println("\tcurrent object:\t", xy /*, "\n\tprevious one:\t", xy.Prev */)
	//	fmt.Println("\tString representation: ", xy.CoordString)
	//	fmt.Println("\tReal coordinates (x,y): ", xy.X.getfval(), ",", xy.Y.getfval())
	//	fmt.Println("\tReal coordinates (i,j): ", xy.I.getfval(), ",", xy.J.getfval())
}

func (xy *XY) String() string {
	// "XY object # nnn : (xxx, yyy)
	retVal := "XY object #" +
		strconv.Itoa(int(xy.nodeNumber)) +
		": x,y=(" +
		strconv.FormatFloat(xy.x.getfval(), 'f', 5, 64) +
		"," +
		strconv.FormatFloat(xy.y.getfval(), 'f', 5, 64) +
		") " +
		"i,j=(" +
		strconv.FormatFloat(xy.i.getfval(), 'f', 5, 64) +
		"," +
		strconv.FormatFloat(xy.j.getfval(), 'f', 5, 64) +
		")"
	return retVal
}

// tolerance is the radius of the circle around first point
// inisde of which another point will be treated as equal to the first one
func (xy *XY) Equals(another *XY, tolerance float64) bool {
	return (math.Hypot(xy.GetX()-another.GetX(), xy.GetY()-another.GetY())) < tolerance
}

func (xy *XY) GetX() float64 {
	return xy.x.valFloat
}

func (xy *XY) SetX(x float64) {
	xy.x.valFloat = x
}

func (xy *XY) GetY() float64 {
	return xy.y.valFloat
}

func (xy *XY) SetY(y float64) {
	xy.y.valFloat = y
}

func (xy *XY) GetI() float64 {
	return xy.i.valFloat
}

func (xy *XY) SetI(i float64) {
	xy.i.valFloat = i
}

func (xy *XY) GetJ() float64 {
	return xy.j.valFloat
}

func (xy *XY) SetJ(j float64) {
	xy.j.valFloat = j
}

func (xy *XY) Init(sc string, fs *FormatSpec, prev *XY) bool {
	var result = false
	if prev == nil { // first node
		xy.nodeNumber = 0
		xy.x.clear()
		xy.y.clear()
	} else {
		*xy = *prev
		xy.nodeNumber = prev.nodeNumber + 1
	}
	// offsets are not modal
	xy.i.clear()
	xy.j.clear()
	xy.coordString = strings.ToUpper(sc)
	//	fmt.Println("XY.Init -> xy.CoordString :", xy.CoordString)
	xi := fs.ReadXI()
	xd := fs.ReadXD()
	masks := []byte{'X', 'Y', 'I', 'J', 'D'}
	mpos := []int{-1, -1, -1, -1, -1}
	var found int = 0 // found signatures
	for i := range masks {
		mpos[i] = strings.IndexByte(xy.coordString, masks[i])
		//		fmt.Println(masks[i], "at ", mpos[i])
		if mpos[i] != -1 {
			found++
		}
	}
	if mpos[len(mpos)-1] == -1 {
		// eror in string, no trailing D symbol
		return result
	}
	if mpos[len(mpos)-1] != (len(xy.coordString) - 1) {
		// eror in string, trailing D symbol is not at last position
		return result
	}
	m2 := make([]byte, found) // mask array contains only found LETTERS
	p2 := make([]int, found)  // and their positions
	j := 0
	for i := range masks {
		if mpos[i] != -1 {
			p2[j] = mpos[i]
			m2[j] = masks[i]
			j++
		}
	}
	// sort

	i := 0
	for {
		if i < j-1 {
			if p2[i] > p2[i+1] {
				p2[i], p2[i+1] = p2[i+1], p2[i]
				m2[i], m2[i+1] = m2[i+1], m2[i]
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

	sf := fs.ReadMU()
L1:
	for i := range m2 {
		switch m2[i] {
		case 'X':
			// possibly X value detected
			if xy.x.init(xy.coordString[p2[i]+1:p2[i+1]], xi, xd, sf) == false {
				result = false
				break L1
			}
		case 'Y':
			// possibly Y value detected
			if xy.y.init(xy.coordString[p2[i]+1:p2[i+1]], xi, xd, sf) == false {
				result = false
				break L1
			}
		case 'I':
			// possibly I value detected
			if xy.i.init(xy.coordString[p2[i]+1:p2[i+1]], xi, xd, sf) == false {
				result = false
				break L1
			}
		case 'J':
			// possibly J value detected
			if xy.j.init(xy.coordString[p2[i]+1:p2[i+1]], xi, xd, sf) == false {
				result = false
				break L1
			}
		case 'D':
			// trailing symbol found
			result = true
			break L1
		default:
			// nothing was found
			result = false
			break L1
		}
	}
	if result == false {
		// clear all fields and links
		xy.x.clear()
		xy.y.clear()
		xy.i.clear()
		xy.j.clear()
		xy.coordString = ""
	}
	return result
}

/*####################  regions ##################################
 */
type Region struct {
	startXY         *XY // pointer to start entry
	numberOfXY      int // number of entries
	G36StringNumber int // number of the string with G36 cmd
	G37StringNumber int // number of the string with G37 cmd
}

func (region *Region) String() string {

	xyText := "<nil>"
	if region == nil {
		return xyText
	}

	if region.startXY != nil {
		xyText = region.startXY.String()
	}
	return "Region:\n" +
		"\t\tstart point: " + xyText + "\n" +
		"\t\tcontains " + strconv.Itoa(region.numberOfXY) + " vertices\n" +
		"\t\tG36 command is at line " + strconv.Itoa(region.G36StringNumber) + "\n" +
		"\t\tG37 command is at line " + strconv.Itoa(region.G37StringNumber)
}

// creates and initialises a region object
func newRegion(strNum int) *Region {
	retVal := new(Region)
	retVal.G36StringNumber = strNum
	retVal.numberOfXY = 0
	retVal.G37StringNumber = -1
	return retVal
}

// closes the region
func (region *Region) Close(strnum int) error {
	if region == nil {
		return errors.New("can not close the contour referenced by null pointer")
	}
	region.G37StringNumber = strnum

	//	fmt.Println(region.String())

	return nil
}

// sets a start coordinate entry
func (region *Region) setStartXY(in *XY) {
	region.startXY = in
	region.numberOfXY++
}

// returns a start coordinate entry
func (region *Region) getStartXY() *XY {
	return region.startXY
}

// increments number of coordinate entries
func (region *Region) incNumXY() int {
	region.numberOfXY++
	return region.numberOfXY
}

// returns the number of coordinate entries of the contour
func (region *Region) GetNumXY() int {
	return region.numberOfXY
}

// returns true if region is opened
func (region *Region) isRegionOpened() (bool, error) {
	if region == nil {
		return false, errors.New("bad region referenced (by nil ptr)")
	}
	if region.G37StringNumber == -1 {
		return true, nil
	} else {
		return false, nil
	}
}

/*
############################## step and repeat blocks #################################
*/
type SRBlock struct {
	srString string
	numX     int
	numY     int
	dX       float64
	dY       float64
	nSteps   int // number of steps in the SRBlock block
}

func (srblock *SRBlock) String() string {

	if srblock == nil {
		return "<nil>"
	}
	return "Step and repeat block:\n" +
		"\tsource string: " + srblock.srString + "\n" +
		"\tcontains " + strconv.Itoa(srblock.numX) + " repeats along X axis and " + strconv.Itoa(srblock.numY) + " repeats along Y axis\n" +
		"\tnumber of steps in each repetition: " + strconv.Itoa(srblock.nSteps) + "\n" +
		"\tdX=" + strconv.FormatFloat(srblock.dX, 'f', 5, 64) +
		", dy=" + strconv.FormatFloat(srblock.dX, 'f', 5, 64) + "\n"
}

func (srblock *SRBlock) NumX() int {
	return srblock.numX
}

func (srblock *SRBlock) NumY() int {
	return srblock.numY
}

func (srblock *SRBlock) DX() float64 {
	return srblock.dX
}

func (srblock *SRBlock) DY() float64 {
	return srblock.dY
}

func (srblock *SRBlock) NSteps() int {
	return srblock.nSteps
}

func (srblock *SRBlock) incNSteps() {
	srblock.nSteps++
}

func (srblock *SRBlock) Init(ins string, fs *FormatSpec) error {
	ins = strings.TrimSpace(ins)
	res, err := ExtractLetterDelimitedFloats(ins, "XYIJ")
	if err != nil {
		return err
	}
	if len(res) != 4 {
		return errors.New("SRBlock.Init: missing one or some SRBlock parameter(s)")
	}
	srblock.numX = int(res['X'])
	if srblock.numX < 1 {
		return errors.New("SRBlock.Init: X count < 1")
	}
	srblock.numY = int(res['Y'])
	if srblock.numY < 1 {
		return errors.New("SRBlock.Init: Y count < 1")
	}
	srblock.dX = res['I'] * fs.ReadMU() // take into account inches or millimeters
	srblock.dY = res['J'] * fs.ReadMU()
	srblock.srString = ins
	return nil
}

//func (srblock *SRBlock) StartXY() *XY {
//	return srblock.startXY
//}
//
//func (srblock *SRBlock) SetStartXY(v *XY) {
//	srblock.startXY = v
//}

// the function splits the input string by substrings using template's symbols as ordered delimiters and returns
// a map symbol:value

func ExtractLetterDelimitedFloats(ins, template string) (out map[byte]float64, err error) {

	out = make(map[byte]float64)
	p := make([]int, len(template))
	ts := []byte(template)
	for i := range template {
		p[i] = strings.IndexByte(ins, template[i])
	}
	i := 0
	j := len(template) - 1
	for {
		if i < j {
			if p[i] > p[i+1] {
				p[i], p[i+1] = p[i+1], p[i]
				ts[i], ts[i+1] = ts[i+1], ts[i]
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
	/*
		fmt.Println(p)
		fmt.Println(ts)
	*/
	for i := range p {
		if p[i] == -1 {
			continue
		}
		var ii int
		if i == len(p)-1 {
			ii = len(ins)
		} else {
			ii = p[i+1]
		}
		fv, err := strconv.ParseFloat(ins[p[i]+1:ii], 64)
		if err != nil {
			return nil, err
		}
		out[template[i]] = fv
	}

	return out, nil
}

func checkError(err error, exitCode int) {
	if err != nil {
		fmt.Println(err)
		os.Exit(exitCode)
	}
}
