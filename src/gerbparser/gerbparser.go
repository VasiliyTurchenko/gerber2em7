/*
Package gerbparser contains functions and structures used for parsing gerber x2 file
*/
package gerbparser

import "fmt"
import "strings"
import "strconv"
import (
	"container/list"
	"errors"
	"math"
	"os"
)

const GerberComment string = "G04"
const GerberEOF string = "M02"
const GerberFormatSpec string = "%FSLA"
const GerberMOIN string = "%MOIN*%"
const GerberMOMM string = "%MOMM*%"

const InchesToMM float64 = 25.4

var GerberCommands = []string{
	"%FS", // Format specification. Sets the coordinate format, e.g. the number of decimals. 4.1
	"%MO", // Mode. Sets the unit to inch or mm. 4.2
	"%AD", // Aperture define. Defines a template based aperture and assigns a D code to it. 4.3
	"%AM", // Aperture macro. Defines a macro aperture template. 4.5
	"%AB", // Aperture block. Defines a block aperture and assigns a D-code to it. 4.6
	"D01", /* Interpolate operation. Outside a region statement D01 creates a draw or arc
	object using the current aperture. Inside it creates a linear or circular contour
	segment. After the D01 command the current point is moved to draw/arc end
	point.4.8 */
	"D02", /* Move operation. D02 does not create a graphics object but moves the current
	point to the coordinate in the D02 command. 4.8 */
	"D03", /* Flash operation. Creates a flash object with the current aperture. After the D03
	command the current point is moved to the flash point. 	4.8 */
	"D",   // Dnn (nn≥10) Sets the current aperture to D code nn. 4.7
	"G01", // Sets the interpolation mode to linear. 4.9
	"G02", // Sets the interpolation mode to clockwise circular. 4.10
	"G03", // Sets the interpolation mode to counterclockwise circular. 4.10
	"G74", // Sets quadrant mode to single quadrant. 4.10
	"G75", // Sets quadrant mode to multi quadrant. 4.10
	"%LP", // Load polarity. Loads the polarity object transformation parameter. 4.11.2
	"%LM", // Load mirror. Loads the mirror object transformation parameter. 4.11.3
	"%LR", // Load rotation. Loads the rotation object transformation parameter. 4.11.4
	"%LS", // Load scale. Loads the scale object transformation parameter. 4.11.5
	"G36", // Starts a region statement. This creates a region by defining its contour. 4.12.
	"G37", // Ends the region statement. 4.12
	"%SR", // Step and repeat. Open or closes a step and repeat statement. 4.13
	"G04", // Comment. 4.14
	"%TF", // Attribute file. Set a file attribute. 5.2
	"%TA", // Attribute aperture. Add an aperture attribute to the dictionary or modify it. 5.3
	"%TO", // Attribute object. Add an object attribute to the dictionary or modify it. 5.4
	"%TD", // Attribute delete. Delete one or all attributes in the dictionary. 5.5
	"M02", // End of file. 4.15
}

// End of line and end of block
//const GerberEOL string = "*"
const GerberEOB string = "*"

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
	Head  string
	MUstr string
	XI    int // digits in the integer part
	XD    int // digits in the fractional part
	YI    int
	YD    int
	MU    float64
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
	fs.MUstr = strings.ToUpper(mu)

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
	valI     string
	valF     string
	valFloat float64
}

func (ap *axisPoint) clear() {
	ap.valI = ""
	ap.valF = ""
	ap.valFloat = 0.0
}

// initializes the point on the axis ax
// n is the number of places for int part
// m is the number of places for frac part
// s is the scale factor 1.0 or 25.4 (mm/inches)
func (ap *axisPoint) init(ins string, n, m int, s float64) bool {
	var result bool = false
	var neg bool = false
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
	ap.valI = strconv.Itoa(ipart)
	ap.valF = strconv.Itoa(fpart)
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
	prev        *XY
	next        *XY
	nodeNumber  uint32
	coordString string // string representation
	x           axisPoint
	y           axisPoint
	// offsets
	i axisPoint
	j axisPoint
}

func (xy *XY) Print() {
	fmt.Println("Node #: ", xy.nodeNumber)
	fmt.Println("\tcurrent object:\t", xy /*, "\n\tprevious one:\t", xy.Prev */)
	//	fmt.Println("\tString representation: ", xy.CoordString)
	//	fmt.Println("\tReal coordinates (x,y): ", xy.X.getfval(), ",", xy.Y.getfval())
	//	fmt.Println("\tReal coordinates (i,j): ", xy.I.getfval(), ",", xy.J.getfval())
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
	var result bool = false
	if prev == nil { // first node
		xy.prev = nil
		xy.next = nil
		xy.nodeNumber = 0
		xy.x.clear()
		xy.y.clear()
	} else {
		*xy = *prev
		xy.prev = prev
		xy.next = nil
		prev.next = xy
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

	/*
		if m2[0] == 'D' {
			return false
		}
	*/
	/*	if p2[0] != 0 {
			return false
		}
	*/
	//	fmt.Println(m2, p2)
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
		//		xy.Prev.Next = nil
	}
	return result
}

/*####################  regions ##################################
 */
type Region struct {
	//	prev    *Region
	//	next    *Region
	//	apType  *Aperture
	startXY         *XY // pointer to start entry
	numberOfXY      int // number of entries
	G36StringNumber int // number of the string with G36 cmd
	G37StringNumber int // number of the string with G37 cmd
	//	numSegments int // number of closed segments
}

// initialises a region object
func (region *Region) Init( /* pr *Region, */ strnum int /*apert *Aperture*/) error {
	if region == nil {
		return errors.New("can not create the contour referenced by null pointer")
	}
	/*	if pr == nil {
			region.prev = region
		} else {
			pr.next = region // not the first element
			region.prev = pr
		}
	*/
	region.G36StringNumber = strnum
	//	region.apType = apert
	region.numberOfXY = 0
	region.G37StringNumber = -1
	return nil
}

// closes the region
func (region *Region) Close(strnum int) error {
	if region == nil {
		return errors.New("can not close the contour referenced by null pointer")
	}
	region.G37StringNumber = strnum
	return nil
}

// sets a start coordinate entry
func (region *Region) SetStartXY(in *XY) {
	region.startXY = in
	region.numberOfXY++
}

// returns a start coordinate entry
func (region *Region) GetStartXY() *XY {
	return region.startXY
}

// increments number of coordinate entries
func (region *Region) IncNumXY() int {
	region.numberOfXY++
	return region.numberOfXY
}

// returns the number of coordinate entries of the contour
func (region *Region) GetNumXY() int {
	return region.numberOfXY
}

// returns true if region is opened
func (region *Region) RegionOpened() (bool, error) {
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
type SR struct {
	srString string
	startXY  *XY
	numX     int
	numY     int
	dX       float64
	dY       float64
	nSteps   int // number of steps in the SR block
}

func (srblock *SR) NumX() int {
	return srblock.numX
}

func (srblock *SR) NumY() int {
	return srblock.numY
}

func (srblock *SR) DX() float64 {
	return srblock.dX
}

func (srblock *SR) DY() float64 {
	return srblock.dY
}

func (srblock *SR) NSteps() int {
	return srblock.nSteps
}

func (srblock *SR) IncNSteps() {
	srblock.nSteps++
}

func (srblock *SR) Init(ins string, fs *FormatSpec) error {
	ins = strings.TrimSpace(ins)
	res, err := ExtractLetterDelimitedFloats(ins, "XYIJ")
	if err != nil {
		return err
	}
	if len(res) != 4 {
		return errors.New("SR.Init: missing SR parameter(s)")
	}
	srblock.numX = int(res['X'])
	if srblock.numX < 1 {
		return errors.New("SR.Init: X count < 1")
	}
	srblock.numY = int(res['Y'])
	if srblock.numY < 1 {
		return errors.New("SR.Init: Y count < 1")
	}
	srblock.dX = res['I'] * fs.ReadMU() // take into account inches or millimeters
	srblock.dY = res['J'] * fs.ReadMU()
	srblock.srString = ins
	return nil
}

func (srblock *SR) StartXY() *XY {
	return srblock.startXY
}

func (srblock *SR) SetStartXY(v *XY) {
	srblock.startXY = v
}

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

/*
################## apertures #############################
*/
// Apertures
const GerberApertureDef = "%ADD"
const GerberApertureMacroDef = "%AM"
const GerberApertureBlockDef = "%AB"
const GerberApertureBlockDefEnd = "%AB*%"

type GerberAptype int

const (
	AptypeCircle    GerberAptype = iota + 1
	AptypeRectangle
	AptypeObround
	AptypePoly
	AptypeMacro
	AptypeBlock
)

type Aperture struct {
	Code         int
	SourceString string
	Type         GerberAptype
	XSize        float64
	YSize        float64
	Diameter     float64
	HoleDiameter float64
	Vertices     int
	RotAngle     float64
	BlockPtr     *BlockAperture
	MacroPtr     *MacroAperture
}

type BlockAperture struct {
	StartStringNum int
	Code           int
	APBodyPtr  []string
	APStepsPtr []*State
}

func (ba *BlockAperture) Print() {
	fmt.Println("\n***** Block aperture *****")
	fmt.Println("\tBlock aperture code:", ba.Code)
	fmt.Println("\tSource strings:")
	for b := range ba.APBodyPtr {
		fmt.Println("\t\t", b, "  ", ba.APBodyPtr[b])
	}
	fmt.Println("\tResulting steps:")
	for b := range ba.APStepsPtr {
		//		fmt.Printf("\t\t%d%s%v\n", b, "  ", ba.APStepsPtr[b])
		ba.APStepsPtr[b].Print()
	}

}

/*
func (ba *BlockAperture) Init(bodylen int) error {
	ba.BlockBody = make([]*string, bodylen)
	ba.Indx = 0
	return nil
}

func (ba *BlockAperture) Add(ins *string) (int, error) {
	if ba.Indx == len(ba.BlockBody) {
		return ba.Indx, errors.New("aperture block is greater than was defined")
	} else {
		ba.BlockBody[ba.Indx] = ins
		ba.Indx++
		return ba.Indx, nil
	}
}
*/
type MacroAperture struct {
	dummy int
}

func (apert *Aperture) GetCode() int {
	return apert.Code
}

func (apert *Aperture) Init(sourceString string, fs *FormatSpec) error {
	//	var result error = nil
	//	sourceString = strings.ToUpper(sourceString)
	sourceString = strings.TrimSpace(sourceString)
	var err error = nil
	apert.SourceString = sourceString

	apertureCodePosition := strings.IndexAny(sourceString, "CROP")
	apert.Code, err = strconv.Atoi(sourceString[:apertureCodePosition])
	if err == nil {
		var tmpVal float64
		tmpSplitted := strings.Split(sourceString[apertureCodePosition+2:], "X")
		for j := range tmpSplitted {
			tmpSplitted[j] = strings.TrimSpace(tmpSplitted[j])
		}
		switch sourceString[apertureCodePosition] {
		case 'C':
			apert.Type = AptypeCircle
			if len(tmpSplitted) == 1 || len(tmpSplitted) == 2 {
				for i, s := range tmpSplitted {
					tmpVal, err = strconv.ParseFloat(s, 32)
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
					tmpVal, err = strconv.ParseFloat(s, 32)
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
					tmpVal, err = strconv.ParseFloat(s, 32)
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
					tmpVal, err = strconv.ParseFloat(s, 32)
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
			err = errors.New("aperture macro is not supported")
			goto fExit
		}
	} else {
		err = errors.New("bad aperture number ")
	}

	apert.HoleDiameter *= fs.ReadMU()
	apert.Diameter *= fs.ReadMU()
	apert.YSize *= fs.ReadMU()
	apert.XSize *= fs.ReadMU()

fExit:
	return err
}

/*
################################## State machine ######################################
*/
type Poltype int

const (
	PoltypeDark  Poltype = iota + 1
	PoltypeClear
)

type Acttype int

const (
	OpcodeD01  Acttype = iota + 1
	OpcodeD02
	OpcodeD03
	OpcodeStop
)

type Quadmode int

const (
	QuadmodeSingle Quadmode = iota + 1
	QuadmodeMulti
)

type IPmode int

const (
	IPModeLinear IPmode = iota + 1
	IPModeCwC
	IPModeCCwC
)

/*
	The State object represents the state of the state machine before processing
	the state.
 */
type State struct {
	// each instance of the State represents an action which must be done
	StepNumber  int     // step number
	Polarity    Poltype // %LPD*% or %LPC*%
	QMode       Quadmode
	CurrentAp   *Aperture // aperture code
	IpMode      IPmode    // interpolation mode
	PrevCoord   *XY
	Coord       *XY
	Action      Acttype
	Region      *Region
	SRBlock     *SR
	OriginForAB *XY // origin for aperture block insertion
}

// diagnostic print
func (step *State) Print() {
	fmt.Println("Step#", step.StepNumber)
	fmt.Println("\tPolarity", step.Polarity)
	fmt.Println("\tQuadrant mode", step.QMode)
	fmt.Println("\tInterpolation mode", step.IpMode)
	if step.CurrentAp != nil {
		fmt.Println("\tAperture", step.CurrentAp.Code)
	} else {
		fmt.Println("\tAperture <nil>")
	}
	fmt.Println("\tAction", step.Action)
	fmt.Println("\tRegion", step.Region)
	fmt.Println("\tS&R block", step.SRBlock)
	if step.PrevCoord != nil {
		fmt.Printf("\t%s%f%s%f\n", "Prev.X=", step.PrevCoord.GetX(), "  Prev.Y=", step.PrevCoord.GetY())
	} else {
		fmt.Println("\tPrev.coord <nil>")
	}
	fmt.Printf("\t%s%f%s%f\n", "X=", step.Coord.GetX(), "  Y=", step.Coord.GetY())
	fmt.Printf("\t%s%f%s%f\n", "I=", step.Coord.GetI(), "  J=", step.Coord.GetJ())
	if step.OriginForAB != nil {
		fmt.Printf("\t%s%f%s%f\n", "Origin X=", step.OriginForAB.GetX(), "  Origin Y=", step.OriginForAB.GetY())
	} else {
		fmt.Println("\tOrigin fro apert.block <nil>")
	}
}

// creates and intializes step object with default values
func NewStep() *State {
	step := new(State)
	xy := new(XY)
	xy.SetX(0)
	xy.SetY(0)
	step.Coord = xy
	step.Polarity = PoltypeDark
	step.IpMode = IPModeLinear
	return step
}

type GerberStringProcessingResult int

const (
	SCResultNextString GerberStringProcessingResult = iota + 1 // need next string to complete step
	SCResultSkipString                                         // string was skipped
	SCResultStepCmpltd                                         // step creation completed
	SCResultStop
)

func (step *State) CreateStep(ins *string, prevstep *State, apertl *list.List, regl *list.List, i int, fSpec *FormatSpec) GerberStringProcessingResult {

	// sequentally fill all the fields
	// after opcode string finalize the step
	if strings.Compare(*ins, "G01*") == 0 || strings.Compare(*ins, "G1*") == 0 { // +09-Jun-2018
		step.IpMode = IPModeLinear
		return SCResultNextString
	}
	if strings.Compare(*ins, "G02*") == 0 || strings.Compare(*ins, "G2*") == 0 { // +09-Jun-2018
		step.IpMode = IPModeCwC
		return SCResultNextString
	}
	if strings.Compare(*ins, "G03*") == 0 || strings.Compare(*ins, "G3*") == 0 { // +09-Jun-2018
		step.IpMode = IPModeCCwC
		return SCResultNextString
	}
	if strings.Compare(*ins, "%LPC*%") == 0 {
		step.Polarity = PoltypeClear
		return SCResultNextString
	}
	if strings.Compare(*ins, "%LPD*%") == 0 {
		step.Polarity = PoltypeDark
		return SCResultNextString
	}
	if strings.Compare(*ins, "G74*") == 0 {
		step.QMode = QuadmodeSingle
		return SCResultNextString
	}
	if strings.Compare(*ins, "G75*") == 0 {
		step.QMode = QuadmodeMulti
		return SCResultNextString
	}
	if strings.Compare("G37*", *ins) == 0 {
		// G37 command is found
		regionOPenedState, err := step.Region.RegionOpened()
		CheckError(err, 401)
		if regionOPenedState == true { // creg is opened
			err = step.Region.Close(i)
			CheckError(err, 402)
			step.Region = nil
		}
		//		rgn = nil
		return SCResultNextString
	}
	//
	if strings.Compare("G36*", *ins) == 0 {
		creg := new(Region)
		err := creg.Init(i)
		CheckError(err, 400)
		regl.PushBack(creg)
		step.Region = creg
		// add coordinates as usual, close creg at G37 command
		return SCResultNextString
	}
	switch {
	case strings.HasSuffix(*ins, "D01*"):
		step.Action = OpcodeD01
	case strings.HasSuffix(*ins, "D02*"):
		step.Action = OpcodeD02
	case strings.HasSuffix(*ins, "D03*"):
		step.Action = OpcodeD03
	}
	if strings.HasSuffix(
		*ins, "D01*") || strings.HasSuffix(
		*ins, "D02*") || strings.HasSuffix(
		*ins, "D03*") {
		xy := new(XY)
		abxy := new(XY)
		s := *ins
		if xy.Init(s[:len(s)-3], fSpec /*step.PrevCoord*/ , prevstep.Coord) != false { // coordinates are recognized successfully
			//			step.PrevCoord = xy
			step.Coord = xy
			step.OriginForAB = abxy
			//				fmt.Println("string:", i, "\tcoordinates(X,Y,I,J):", xy.GetX(), xy.GetY(), xy.GetJ(), xy.GetJ())
			// check if the xy belongs to a region
			if step.Region != nil {
				rs, _ := step.Region.RegionOpened()
				if rs == true {
					// add coordinate entry into creg
					if step.Region.GetNumXY() == 0 {
						// no coordinate entries in the creg
						step.Region.SetStartXY(xy)
					} else {
						step.Region.IncNumXY()
					}
				}
			}
		} else {
			fmt.Println("Error parsing string", i, *ins)
			panic("310")
			os.Exit(310)
		}
		if step.SRBlock != nil {
			step.SRBlock.IncNSteps()
		}
		return SCResultStepCmpltd
	}

	// switch aperture
	s := strings.TrimPrefix(*ins, "G54")
	if strings.HasPrefix(s, "D") && strings.HasSuffix(s, "*") {
		var tc int
		step.CurrentAp = nil
		tc, err := strconv.Atoi(s[1 : len(s)-1])
		CheckError(err, 501)
		for k := apertl.Front(); k != nil; k = k.Next() {
			if k.Value.(*Aperture).GetCode() == tc {
				step.CurrentAp = k.Value.(*Aperture)
				break
			}
		}
		if step.CurrentAp == nil {
			CheckError(errors.New("the aperture does not exist"), 502)
		}
		return SCResultNextString
	}

	if strings.HasPrefix(*ins, "%SRX") {
		fmt.Println("Step and repeat block found at line", i)
		step.SRBlock = new(SR)
		s := *ins
		srerr := step.SRBlock.Init(s[3:len(s)-2], fSpec)
		CheckError(srerr, 550)
		//		SRBlocks = append(SRBlocks, srblock)
		//		srb = srblock
		return SCResultNextString
	}

	if strings.HasPrefix(s, "%SR*%") {
		fmt.Println("Step and repeat block ends at line", i)
		step.SRBlock = nil
		return SCResultNextString
	}

	if strings.Compare(s, "M02*") == 0 || strings.Compare(s, "M00*") == 0 {
		fmt.Println("Stop found at line", i)
		step.Action = OpcodeStop
		// create the special last step
		//		step.StepNumber = prevstep.StepNumber + 1
		step.SRBlock = nil // also closes s&r block
		return SCResultStop
	}

	return SCResultSkipString
}

func CheckError(err error, exitcode int) {
	if err != nil {
		fmt.Println(err)
		os.Exit(exitcode)
	}
}

/*
********************* aperture blocks ******************************
 */
