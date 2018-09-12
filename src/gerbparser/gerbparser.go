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
	retVal:= "XY object # " +
		strconv.Itoa(int(xy.nodeNumber)) +
		" :(" +
		strconv.FormatFloat(xy.x.getfval(),'f', 5,64) +
		"," +
		strconv.FormatFloat(xy.x.getfval(),'f',5, 64)
	return retVal
}

// tolerance is the radius of the circle around first point
// inisde of which another point will be treated as equal to the first one
func (xy *XY) Equals(another *XY, tolerance float64) bool {
	return ( math.Hypot(xy.GetX() - another.GetX(), xy.GetY() - another.GetY()) ) < tolerance
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
	startXY  *XY
	numX     int
	numY     int
	dX       float64
	dY       float64
	nSteps   int // number of steps in the SRBlock block
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

func (srblock *SRBlock) StartXY() *XY {
	return srblock.startXY
}

func (srblock *SRBlock) SetStartXY(v *XY) {
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
	BodyStrings    []string
	StepsPtr       []*State
}

func (ba *BlockAperture) Print() {
	fmt.Println("\n***** Block aperture *****")
	fmt.Println("\tBlock aperture code:", ba.Code)
	fmt.Println("\tSource strings:")
	for b := range ba.BodyStrings {
		fmt.Println("\t\t", b, "  ", ba.BodyStrings[b])
	}
	fmt.Println("\tResulting steps:")
	for b := range ba.StepsPtr {
		ba.StepsPtr[b].Print()
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
type polType int

const (
	PolTypeDark polType = iota + 1
	PolTypeClear
)

type Acttype int

const (
	OpcodeD01_DRAW Acttype = iota + 1
	OpcodeD02_MOVE
	OpcodeD03_FLASH
	OpcodeStop
)

type QuadMode int

const (
	QuadModeSingle QuadMode = iota + 1
	QuadModeMulti
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
	Polarity    polType // %LPD*% or %LPC*%
	QMode       QuadMode
	CurrentAp   *Aperture // aperture code
	IpMode      IPmode    // interpolation mode
	PrevCoord   *XY
	Coord       *XY
	Action      Acttype
	Region      *Region
	SRBlock     *SRBlock
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
		fmt.Println("\tOrigin from apert.block <nil>")
	}
}

// creates and intializes step object with default values
func NewState() *State {
	state := new(State)
	state.Coord = NewXY()
	state.Polarity = PolTypeDark
	state.IpMode = IPModeLinear
	return state
}

func (step *State) CopyOfWithOffset(another *State, addX float64, addY float64) {
	step.Action = another.Action
	step.Region = another.Region
	step.SRBlock = another.SRBlock
	step.IpMode = another.IpMode
	step.QMode = another.QMode
	step.CurrentAp = another.CurrentAp
	step.Polarity = another.Polarity
	step.StepNumber = another.StepNumber
	step.Coord = new(XY)
	step.Coord.SetX(another.Coord.GetX() + addX)
	step.Coord.SetY(another.Coord.GetY() + addY)
	step.Coord.SetI(another.Coord.GetI())
	step.Coord.SetJ(another.Coord.GetJ())
}



type GerberStringProcessingResult int

const (
	SCResultNextString    GerberStringProcessingResult = iota + 1 // need next string to complete step
	SCResultSkipString                                            // string was skipped
	SCResultStepCompleted                                         // step creation completed
	SCResultStop
)

func (step *State) CreateStep(
	inString *string,
	prevStep *State,
	apertList *list.List,
	regionsList *list.List,
	i int,
	fSpec *FormatSpec) GerberStringProcessingResult {

	// sequentally fill all the fields
	// after opcode string finalize the step
	if strings.Compare(*inString, "G01*") == 0 || strings.Compare(*inString, "G1*") == 0 { // +09-Jun-2018
		step.IpMode = IPModeLinear
		return SCResultNextString
	}
	if strings.Compare(*inString, "G02*") == 0 || strings.Compare(*inString, "G2*") == 0 { // +09-Jun-2018
		step.IpMode = IPModeCwC
		return SCResultNextString
	}
	if strings.Compare(*inString, "G03*") == 0 || strings.Compare(*inString, "G3*") == 0 { // +09-Jun-2018
		step.IpMode = IPModeCCwC
		return SCResultNextString
	}
	if strings.Compare(*inString, "%LPC*%") == 0 {
		step.Polarity = PolTypeClear
		return SCResultNextString
	}
	if strings.Compare(*inString, "%LPD*%") == 0 {
		step.Polarity = PolTypeDark
		return SCResultNextString
	}
	if strings.Compare(*inString, "G74*") == 0 {
		step.QMode = QuadModeSingle
		return SCResultNextString
	}
	if strings.Compare(*inString, "G75*") == 0 {
		step.QMode = QuadModeMulti
		return SCResultNextString
	}
	if strings.Compare("G37*", *inString) == 0 {
		// G37 command is found
		regionOpenedState, err := step.Region.isRegionOpened()
		checkError(err, 401)
		if regionOpenedState == true { // creg is opened
			err = step.Region.Close(i)
			checkError(err, 402)
			step.Region = nil
		}
		return SCResultNextString
	}
	//
	if strings.Compare("G36*", *inString) == 0 {
		creg := newRegion(i)
		regionsList.PushBack(creg)
		step.Region = creg
		// add coordinates as usual, close creg at G37 command
		return SCResultNextString
	}
	switch {
	case strings.HasSuffix(*inString, "D01*"):
		step.Action = OpcodeD01_DRAW
	case strings.HasSuffix(*inString, "D02*"):
		step.Action = OpcodeD02_MOVE
	case strings.HasSuffix(*inString, "D03*"):
		step.Action = OpcodeD03_FLASH
	}
	if strings.HasSuffix(
		*inString, "D01*") || strings.HasSuffix(
		*inString, "D02*") || strings.HasSuffix(
		*inString, "D03*") {
		xy := new(XY)
		abxy := new(XY)
		s := *inString
		if xy.Init(s[:len(s)-3], fSpec /*step.PrevCoord*/ , prevStep.Coord) != false { // coordinates are recognized successfully
			//			step.PrevCoord = xy
			step.Coord = xy
			step.OriginForAB = abxy
			//				fmt.Println("string:", i, "\tcoordinates(X,Y,I,J):", xy.GetX(), xy.GetY(), xy.GetJ(), xy.GetJ())
			// check if the xy belongs to a region
			if step.Region != nil {
				rs, _ := step.Region.isRegionOpened()
				if rs == true {
					// add coordinate entry into creg
					if step.Region.GetNumXY() == 0 {
						// no coordinate entries in the creg
						step.Region.setStartXY(xy)
					} else {
						step.Region.incNumXY()
					}
				}
			}
		} else {
			fmt.Println("Error parsing string", i, *inString)
			panic("310")
			os.Exit(310)
		}
		if step.SRBlock != nil {
			step.SRBlock.incNSteps()
		}
		return SCResultStepCompleted
	}

	// switch aperture
	s := strings.TrimPrefix(*inString, "G54")
	if strings.HasPrefix(s, "D") && strings.HasSuffix(s, "*") {
		var tc int
		step.CurrentAp = nil
		tc, err := strconv.Atoi(s[1 : len(s)-1])
		checkError(err, 501)
		for k := apertList.Front(); k != nil; k = k.Next() {
			if k.Value.(*Aperture).GetCode() == tc {
				step.CurrentAp = k.Value.(*Aperture)
				break
			}
		}
		if step.CurrentAp == nil {
			checkError(errors.New("the aperture does not exist"), 502)
		}
		return SCResultNextString
	}

	if strings.HasPrefix(*inString, "%SRX") {
		fmt.Println("Step and repeat block found at line", i)
		step.SRBlock = new(SRBlock)
		s := *inString
		srerr := step.SRBlock.Init(s[3:len(s)-2], fSpec)
		checkError(srerr, 550)
		//		SRBlocks = append(SRBlocks, srblock)
		//		srb = srblock
		return SCResultNextString
	}

	if strings.HasPrefix(s, "%SRBlock*%") {
		fmt.Println("Step and repeat block ends at line", i)
		step.SRBlock = nil
		return SCResultNextString
	}

	if strings.Compare(s, "M02*") == 0 || strings.Compare(s, "M00*") == 0 {
		fmt.Println("Stop found at line", i)
		step.Action = OpcodeStop
		// create the special last step
		//		step.StepNumber = prevStep.StepNumber + 1
		step.SRBlock = nil // also closes s&r block
		return SCResultStop
	}

	return SCResultSkipString
}

func checkError(err error, exitCode int) {
	if err != nil {
		fmt.Println(err)
		os.Exit(exitCode)
	}
}

/*
********************* aperture blocks ******************************
 */

