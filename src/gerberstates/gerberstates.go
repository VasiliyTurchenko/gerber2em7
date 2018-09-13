/*
################################## State machine ######################################
*/
package gerberstates

import (
	"apertures"
	"container/list"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	."xy"
	."gerberbasetypes"
	"srblocks"
	"regions"
//	"apertures"
)

/*
	The State object represents the state of the state machine before processing
	the state.
*/
type State struct {
	// each instance of the State represents an action which must be done
	StepNumber  int     // step number
	Polarity    PolType // %LPD*% or %LPC*%
	QMode       QuadMode
	CurrentAp   *apertures.Aperture // aperture code
	IpMode      IPmode    // interpolation mode
	PrevCoord   *XY
	Coord       *XY
	Action      ActType
	Region      *regions.Region
	SRBlock     *srblocks.SRBlock
	OriginForAB *XY // origin for aperture block insertion
}

// diagnostic print
func (step *State) Print() {
	fmt.Println("Step#", step.StepNumber)
	fmt.Println("\t" + step.Polarity.String())
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
}

// creates and initializes step object with default values
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
		regionOpenedState, err := step.Region.IsRegionOpened()
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
	if strings.HasSuffix(
		*inString, "D01*") || strings.HasSuffix(
		*inString, "D02*") || strings.HasSuffix(
		*inString, "D03*") {
		xy := new(XY)
		abxy := new(XY)
		s := *inString
		if xy.Init(s[:len(s)-3], fSpec /*step.PrevCoord*/, prevStep.Coord) != false { // coordinates are recognized successfully
			//			step.PrevCoord = xy
			step.Coord = xy
			step.OriginForAB = abxy
			//				fmt.Println("string:", i, "\tcoordinates(X,Y,I,J):", xy.GetX(), xy.GetY(), xy.GetJ(), xy.GetJ())
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
			fmt.Println("Error parsing string", i, *inString)
			panic("310")
			os.Exit(310)
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
		checkError(err, 501)
		for k := apertList.Front(); k != nil; k = k.Next() {
			if k.Value.(*apertures.Aperture).GetCode() == tc {
				step.CurrentAp = k.Value.(*apertures.Aperture)
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
		step.SRBlock = new(srblocks.SRBlock)
		s := *inString
		srerr := step.SRBlock.Init(s[3:len(s)-2], fSpec)
		checkError(srerr, 550)
		//		SRBlocks = append(SRBlocks, srblock)
		//		srb = srblock
		return SCResultNextString
	}

	if strings.HasPrefix(s, "%SR*%") {
		fmt.Println("Step and repeat block ends at line", i)
		fmt.Println(step.SRBlock.String())
		step.SRBlock = nil
		return SCResultNextString
	}

	if strings.Compare(s, "M02*") == 0 || strings.Compare(s, "M00*") == 0 {
		fmt.Println("Stop found at line", i)
		step.Action = OpcodeStop
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
