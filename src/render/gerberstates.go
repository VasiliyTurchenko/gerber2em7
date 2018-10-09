/*
################################## State machine ######################################
*/
package render

import (
	"container/list"
	"errors"
	"fmt"
	. "gerberbasetypes"
	glog "glog_t"
	"image/color"
	"regions"
	"srblocks"
	"strconv"
	"strings"
	. "xy"
)

var stateIdSeed int

func init() {
	stateIdSeed = 0
}

func getNewId() int {
	retVal := stateIdSeed
	stateIdSeed += 1
	return retVal
}

/*
	The State object represents the state of the state machine before processing
	the state.
*/

// the parameters loaded by LS, LM, LR, LP commands
// affect the current aperture when flashing
type ApTransParameters struct {
	Polarity  PolType
	Mirroring Mirror
	Rotation  float64
	Scale     float64
}

func (atp *ApTransParameters) String() string {
	return atp.Polarity.String() + "; " +
		atp.Mirroring.String() +
		"; Rotation=" + strconv.FormatFloat(atp.Rotation, 'f', 5, 64) + "deg.; Scale=" +
		strconv.FormatFloat(atp.Scale, 'f', 5, 64)
}

type State struct {
	// each instance of the State represents an action which must be done
	StepNumber int // step number
	//	Polarity      PolType // %LPD*% or %LPC*%
	QMode         QuadMode
	CurrentAp     *Aperture // aperture code
	IpMode        IPmode    // interpolation mode
	PrevCoord     *XY
	Coord         *XY
	Action        ActType
	Region        *regions.Region
	SRBlock       *srblocks.SRBlock
	OriginForAB   *XY // origin for aperture block insertion
	ApTransParams ApTransParameters
	StateId       int
}

// diagnostic print
func (step *State) Print() {
	fmt.Println("Step#", step.StepNumber)
	fmt.Println("\t" + step.QMode.String())
	fmt.Println("\t" + step.IpMode.String())
	if step.CurrentAp != nil {
		fmt.Println("\tAperture", step.CurrentAp.Code)
	} else {
		fmt.Println("\tAperture <nil>")
	}
	fmt.Println("\t" + step.Action.String())
	fmt.Println("\t" + step.Region.String())
	fmt.Println("\t" + step.SRBlock.String())
	fmt.Print("\tPrev.coordinates:")
	if step.PrevCoord != nil {
		fmt.Println("\t" + step.PrevCoord.String())
	} else {
		fmt.Println("\t<nil>")
	}
	fmt.Print("\tCurrent coordinates:")
	if step.PrevCoord != nil {
		fmt.Println("\t" + step.Coord.String())
	} else {
		fmt.Println("\t<nil>")
	}
	fmt.Print("\tOrigin for aperture block:")
	if step.OriginForAB != nil {
		fmt.Println("\t" + step.OriginForAB.String())
	} else {
		fmt.Println("\t<nil>")
	}
	fmt.Println(step.ApTransParams.String())
}

// creates and initializes step object with default values
func NewState() *State {
	state := new(State)
	state.Coord = NewXY()
	//	state.Polarity = PolTypeDark
	state.IpMode = IPModeLinear
	state.ApTransParams = ApTransParameters{Polarity: PolTypeDark,
		Mirroring: NoMirror,
		Rotation:  0.0,
		Scale:     1.0}
	state.StateId = getNewId()
	return state
}

func (step *State) CopyOfWithOffset(another *State, addX float64, addY float64) {
	step.Action = another.Action
	step.Region = another.Region
	step.SRBlock = another.SRBlock
	step.IpMode = another.IpMode
	step.QMode = another.QMode
	step.CurrentAp = another.CurrentAp
	//	step.Polarity = another.Polarity
	step.StepNumber = another.StepNumber
	step.Coord = new(XY)
	step.Coord.SetX(another.Coord.GetX() + addX)
	step.Coord.SetY(another.Coord.GetY() + addY)
	step.Coord.SetI(another.Coord.GetI())
	step.Coord.SetJ(another.Coord.GetJ())
	step.ApTransParams.Polarity = another.ApTransParams.Polarity
	step.ApTransParams.Scale = another.ApTransParams.Scale
	step.ApTransParams.Rotation = another.ApTransParams.Rotation
	step.ApTransParams.Mirroring = another.ApTransParams.Mirroring
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
	//if strings.Compare(*inString, "%LPC*%") == 0 {
	//	step.Polarity = PolTypeClear
	//		return SCResultNextString
	//	}
	//	if strings.Compare(*inString, "%LPD*%") == 0 {
	//		step.Polarity = PolTypeDark
	//		return SCResultNextString
	//}

	// + 01-Oct-2018
	if strings.Compare(*inString, "%LPC*%") == 0 {
		step.ApTransParams.Polarity = PolTypeClear
		return SCResultNextString
	}
	if strings.Compare(*inString, "%LPD*%") == 0 {
		step.ApTransParams.Polarity = PolTypeDark
		return SCResultNextString
	}

	if strings.Compare(*inString, "%LMN*%") == 0 {
		step.ApTransParams.Mirroring = NoMirror
		return SCResultNextString
	}

	if strings.Compare(*inString, "%LMX*%") == 0 {
		step.ApTransParams.Mirroring = MirrorX
		return SCResultNextString
	}

	if strings.Compare(*inString, "%LMY*%") == 0 {
		step.ApTransParams.Mirroring = MirrorY
		return SCResultNextString
	}

	if strings.Compare(*inString, "%LMXY*%") == 0 {
		step.ApTransParams.Mirroring = MirrorXY
		return SCResultNextString
	}

	if strings.HasPrefix(*inString, "%LR") == true {
		end := strings.Index(*inString, "*")
		val, err := strconv.ParseFloat((*inString)[3:end], 64)
		if err != nil {
			glog.Fatalln(*inString + "unrecoginzed!")
		}
		step.ApTransParams.Rotation = val
		return SCResultNextString
	}

	if strings.HasPrefix(*inString, "%LS") == true {
		end := strings.Index(*inString, "*")
		val, err := strconv.ParseFloat((*inString)[3:end], 64)
		if err != nil {
			glog.Fatalln(*inString + "unrecoginzed!")
		}
		step.ApTransParams.Scale = val
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
		regionOpenedState, err := step.Region.IsRegionOpened()
		checkError(err)
		if regionOpenedState == true { // creg is opened
			err = step.Region.Close(i)
			checkError(err)
			step.Region = nil
		}
		return SCResultNextString
	}
	//
	if strings.Compare("G36*", *inString) == 0 {
		creg := regions.NewRegion(i)
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

	// implicit draw

	if strings.HasSuffix(*inString, "D01*") == false &&
		strings.HasSuffix(*inString, "D02*") == false &&
		strings.HasSuffix(*inString, "D03*") == false &&
		strings.HasSuffix(*inString, "*") == true &&
		(strings.HasPrefix(*inString, "X") || strings.HasPrefix(*inString, "Y")) {
		glog.Warningln("Implicit DRAW command found in " + *inString)
		step.Action = OpcodeD01_DRAW
		*inString = strings.TrimSuffix(*inString, "*") + "D01*"
	}

	if strings.HasSuffix(
		*inString, "D01*") || strings.HasSuffix(
		*inString, "D02*") || strings.HasSuffix(
		*inString, "D03*") {
		xy := new(XY)
		abxy := new(XY)
		s := *inString
		if xy.Init(s[:len(s)-3], fSpec, prevStep.Coord) != false { // coordinates are recognized successfully
			step.Coord = xy
			step.OriginForAB = abxy
			// check if the xy belongs to a region
			if step.Region != nil {
				rs, _ := step.Region.IsRegionOpened()
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
			glog.Fatalln("Error parsing string", i, *inString)
		}
		if step.SRBlock != nil {
			step.SRBlock.IncNSteps()
		}
		return SCResultStepCompleted
	}

	// switch aperture
	s := strings.TrimPrefix(*inString, "G54")
	if strings.HasPrefix(s, "D") && strings.HasSuffix(s, "*") {
		var tc int
		step.CurrentAp = nil
		tc, err := strconv.Atoi(s[1 : len(s)-1])
		checkError(err)
		for k := apertList.Front(); k != nil; k = k.Next() {
			if k.Value.(*Aperture).GetCode() == tc {
				step.CurrentAp = k.Value.(*Aperture)
				break
			}
		}
		if step.CurrentAp == nil {
			checkError(errors.New("the aperture " + strconv.Itoa(tc) + " does not exist"))
		}
		return SCResultNextString
	}

	// + 28-09-2018
	if strings.HasPrefix(s, "%SRX1Y1I0") {
		glog.Infoln(step.SRBlock.String()+"ends at line", i)
		step.SRBlock = nil
		return SCResultNextString
	}

	if strings.HasPrefix(*inString, "%SRX") {
		glog.Infoln("Step and repeat block found at line", i)
		step.SRBlock = new(srblocks.SRBlock)
		s := *inString
		srerr := step.SRBlock.Init(s[3:len(s)-2], fSpec)
		checkError(srerr)
		//		SRBlocks = append(SRBlocks, srblock)
		//		srb = srblock
		return SCResultNextString
	}

	if strings.HasPrefix(s, "%SR*%") {
		glog.Infoln("\n"+step.SRBlock.String()+"ends at line", i)
		step.SRBlock = nil
		return SCResultNextString
	}

	if strings.Compare(s, "M02*") == 0 || strings.Compare(s, "M00*") == 0 {
		glog.Infoln("Stop found at line", i)
		step.Action = OpcodeStop
		step.SRBlock = nil // also closes s&r block
		return SCResultStop
	}

	glog.Warningln("skipped: " + *inString)
	return SCResultSkipString
}

func checkError(err error) {
	if err != nil {
		glog.Fatalln(err)
	}
}

// the function creates a full step sequence using src *[]string as source
// src *[]string - pointer to the source string array
// resSteps *[]*gerbparser.State - pointer to the resulting array of the steps, array size must be enough to hold all the staps
// aperturesList *list.List - pointer to the global aperture list
// regionsList *list.List - pointer to the global regions list
// fSpec *gerbparser.FormatSpec - pointer to the format specif. object
// NumberOfSteps - number of the created steps started from 1

func CreateStepSequence(src *[]string,
	resSteps *[]*State,
	apertl *list.List,
	regl *list.List,
	fSpec *FormatSpec) (NumberOfSteps int) {

	stepNumber := 1 // step number
	stepCompleted := true
	// create the root step with default properties
	(*resSteps)[0] = NewState()
	// process string by string
	var step *State
	for i, s := range *src {
		if stepCompleted == true {
			//			step = new(State)
			step = NewState()
			//			*step = *(*resSteps)[stepNumber-1]
			step.CopyOfWithOffset((*resSteps)[stepNumber-1], 0, 0)
			step.Coord = nil
			step.PrevCoord = nil
		}
		createStepResult := step.CreateStep(&s, (*resSteps)[stepNumber-1], apertl, regl, i, fSpec)
		switch createStepResult {
		case SCResultNextString:
			fallthrough
		case SCResultSkipString:
			stepCompleted = false
			continue
		case SCResultStepCompleted:
			step.PrevCoord = (*resSteps)[stepNumber-1].Coord
			step.StepNumber = stepNumber
			(*resSteps)[stepNumber] = step
			stepNumber++
			stepCompleted = true
			continue
		case SCResultStop:
			step.StepNumber = stepNumber
			(*resSteps)[stepNumber] = step
			step.Coord = (*resSteps)[stepNumber-1].Coord
			stepNumber++
			stepCompleted = true
			break
		default:
			break
		}
		//		glog.Warningln("Still unknown command: ", s)
	} // end of input strings parsing
	return stepNumber
}

func UnwindSRBlock(steps *[]*State, k int) (*[]*State, int) {
	firstSRStep := (*steps)[k]
	// once came into, no return until sr block stays not fully processed
	kStop := k + firstSRStep.SRBlock.NSteps() // stop value
	numXSteps := firstSRStep.SRBlock.NumX()
	numYSteps := firstSRStep.SRBlock.NumY()
	numberOfStepsInSRBlock := firstSRStep.SRBlock.NSteps() * numXSteps * numYSteps
	SRBlockSteps := make([]*State, numberOfStepsInSRBlock)
	stepCounter := 0
	var addX, addY float64
	for j := 0; j < numYSteps; j++ {
		addY = float64(j) * firstSRStep.SRBlock.DY()
		for i := 0; i < numXSteps; i++ {
			addX = float64(i) * firstSRStep.SRBlock.DX()
			for kk := k; kk < kStop; kk++ {
				SRBlockSteps[stepCounter] = NewState()
				if kk == k {
					SRBlockSteps[stepCounter].PrevCoord = NewXY()
				} else {
					SRBlockSteps[stepCounter].PrevCoord = SRBlockSteps[stepCounter-1].Coord
				}
				SRBlockSteps[stepCounter].CopyOfWithOffset((*steps)[kk], addX, addY)
				stepCounter++
			}
		}
	}
	return &SRBlockSteps, kStop
}

func (step *State) Render(rc *Render) {

	// polygons are not affected by aperture transformation parameters
	if step.Region != nil {
		// process region
		if rc.PolygonPtr == nil {
			rc.PolygonPtr = NewPolygon(step.Region.G36StringNumber)
		}
		if rc.AddStepToPolygon(step) == step.Region.GetNumXY() {
			// we can process region
			rc.RenderPolygon()
			rc.PolygonPtr = nil
		}
		return
	}

	var Xp int
	var Yp int
	Xc := transformCoord(step.Coord.GetX()-rc.MinX, rc.XRes)
	Yc := transformCoord(step.Coord.GetY()-rc.MinY, rc.YRes)
	if step.PrevCoord == nil {
		Xp = transformCoord(0-rc.MinX, rc.XRes)
		Yp = transformCoord(0-rc.MinY, rc.YRes)
	} else {
		Xp = transformCoord(step.PrevCoord.GetX()-rc.MinX, rc.XRes)
		Yp = transformCoord(step.PrevCoord.GetY()-rc.MinY, rc.YRes)
	}

	if step.Action == OpcodeD02_MOVE {
		rc.MovePen(Xp, Yp, Xc, Yc, rc.MovePenColor)
		return
	}

	if rc.DrawOnlyRegionsMode == true {
		return
	}

	var stepColor color.RGBA

	if step.ApTransParams.Polarity == PolTypeDark {
		stepColor = rc.LineColor
	} else {
		stepColor = rc.ClearColor
	}

	var apertureSize int

	if step.Action == OpcodeD01_DRAW && step.IpMode == IPModeLinear {
		// linear interpolation

		if step.CurrentAp.Type == AptypeCircle {
			//						apertureSize = transformCoord(step.CurrentAp.Diameter, renderContext.XRes)
			apertureSize = transformCoord(step.CurrentAp.Diameter*step.ApTransParams.Scale,
				rc.XRes)
			rc.DrawByCircleAperture(Xp, Yp, Xc, Yc, apertureSize, stepColor)
		} else if step.CurrentAp.Type == AptypeRectangle {
			// draw with rectangle aperture
			w := transformCoord(step.CurrentAp.XSize*step.ApTransParams.Scale,
				rc.XRes)
			h := transformCoord(step.CurrentAp.YSize*step.ApTransParams.Scale,
				rc.YRes)
			rc.DrawByRectangleAperture(Xp, Yp, Xc, Yc, w, h, stepColor)
		} else {
			glog.Fatalln("Error. Only solid drawCircle and solid rectangle may be used to draw.")
		}

		return
	}

	// IPModeCwC, IPModeCwCC
	if step.Action == OpcodeD01_DRAW {
		// non-linear interpolation

		if step.CurrentAp.Type == AptypeCircle {
			apertureSize = transformCoord(step.CurrentAp.Diameter*step.ApTransParams.Scale,
				rc.XRes)
			var (
				fXp, fYp float64
			)
			if step.PrevCoord == nil {
				fXp = transformFloatCoord(0-rc.MinX, rc.XRes)
				fYp = transformFloatCoord(0-rc.MinY, rc.YRes)
			} else {
				fXp = transformFloatCoord(step.PrevCoord.GetX()-rc.MinX, rc.XRes)
				fYp = transformFloatCoord(step.PrevCoord.GetY()-rc.MinY, rc.YRes)
			}

			fXc := transformFloatCoord(step.Coord.GetX()-rc.MinX, rc.XRes)
			fYc := transformFloatCoord(step.Coord.GetY()-rc.MinY, rc.YRes)
			fI := transformFloatCoord(step.Coord.GetI(), rc.XRes)
			fJ := transformFloatCoord(step.Coord.GetJ(), rc.YRes)

			// Arcs require floats!
			err := rc.DrawArc(fXp,
				fYp,
				fXc,
				fYc,
				fI,
				fJ,
				apertureSize,
				step.IpMode,
				step.QMode,
				// TODO
				rc.RegionColor)
			if err != nil {
				step.Print()
				checkError(err)
			}
			rc.DrawDonut(Xp, Yp, apertureSize, 0, stepColor)
			rc.DrawDonut(Xc, Yc, apertureSize, 0, stepColor)
		} else if step.CurrentAp.Type == AptypeRectangle {
			glog.Fatalln("Arc drawing by rectangle aperture is not supported now.")
		} else {
			glog.Fatalln("Only solid drawCircle and solid rectangle may be used to draw.")

		}
		return
	}

	if step.Action == OpcodeD03_FLASH { // flash
		if rc.DrawOnlyRegionsMode != true {
			rc.MovePen(Xp, Yp, Xc, Yc, rc.MovePenColor)
			if step.ApTransParams.Polarity == PolTypeDark {
				step.CurrentAp.Render(Xc, Yc, rc)
			} else {
				glog.Errorln("Flash by clear polarity is not supported yet.")
			}
		}
		return
	}
	// we must not reach this point
	checkError(errors.New("(renderContext *Render) ProcessStep(step *gerbparser.State) internal error. Bad opcode"))
}
