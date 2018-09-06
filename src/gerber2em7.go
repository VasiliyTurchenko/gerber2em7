package main

import (
	"bufio"
	"container/list"
	"errors"
	"flag"
	"fmt"
	"gerbparser"
	"image"
	"image/png"
	"math"
	"os"
	"plotter"

	"render"
	"runtime"
	"strconv"
	"strings"
	"github.com/spf13/viper"
)

const (
	MaxInt   = int(^uint(0) >> 1)
	MinInt   = int(-MaxInt - 1)
	MaxInt32 = int32(math.MaxInt32)
	MinInt32 = int32(math.MinInt32)
	MaxInt64 = int64(math.MaxInt64)
	MinInt64 = int64(math.MinInt64)
)

var ErrbadFS = errors.New("format is not specified or format string parsing error")

var verboselevel = flag.Int("v", 3, "verbose level: 0 - minimal, 3 - maximal")

//var infile = flag.String("i", "", "input file (with path)")
//var outfile = flag.String("o", "", "output file (with path)")
//var logfile = flag.String("l", "", "log file (with path)")

var totalStrings int = 0
var gerberStrings []string // all the strings

//var xy *gerbparser.XY // first coordinates record

var fSpec *gerbparser.FormatSpec
var apert *gerbparser.Aperture

//var creg *gerbparser.Region

var plt *plotter.Plotter

var arrayOfSteps []*gerbparser.State // state machine
//var SRBlocks []*gerbparser.SR        // SR blocks
var regl *list.List   // regions
var apertl *list.List // apertures
var apertblocks map[string]*gerbparser.BlockAperture

var maxX, maxY float64 = 0, 0
var minX, minY float64 = 1000000.0, 1000000.0

var s

func main() {

	viper.SetConfigName("config")     // no need to include file extension
	viper.AddConfigPath(".")  // set the path of your config file

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	} else {
		fmt.Println(viper.GetString("parser.splittedfile"))

	}




		flag.Parse()
	fmt.Println(*verboselevel)
	fmt.Println(returnAppInfo(*verboselevel))
	/*
		fmt.Println("Input inFile: " + *infile)
		fmt.Println("Output inFile: " + *outfile)
		fmt.Println("Log inFile: " + *logfile)
	*/
	/*
	   Process input string
	*/
	PrintMemUsage("Memory usage before reading input file:")

	gerberStrings = make([]string, 0)
	fSpec = new(gerbparser.FormatSpec)

	inFile, err := os.Open("G:\\go_prj\\gerber2em7\\src\\test.g2") // For read access.
	defer closeFile(inFile)
	CheckError(err, -1) // read the file into the array of strings
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		gerberStrings = append(gerberStrings, strings.ToUpper(scanner.Text()))
		totalStrings++
	}
	// // split concatenated command strings AAAAAD01*BBBBBBD02*GNN*D03*etc
	gerberStrings = *splitStrings(&gerberStrings)

	// save splitted strings to a file
	saveIntermediate(&gerberStrings, "splitted.txt")

	// remove comments and other non-nesessary strings
	gerberStrings = *squeezeStrings(&gerberStrings)

	// search for format definition strings
	mo, err := searchMO()
	CheckError(err, 300)

	fs, err := searchFS()
	CheckError(err, 301)

	fSpec = new(gerbparser.FormatSpec)
	if fSpec.Init(fs, mo) == false {
		fmt.Println("Can not parse:")
		fmt.Println(fs)
		fmt.Println(mo)
		os.Exit(302)
	}

	/* ---------------------- extract apertures and aperture blocks  --------------------- */
	// and aperture macros - TODO!!!!!
	apertl = list.New()
	apertblocks = make(map[string]*gerbparser.BlockAperture)
	abopened := make([]string, 0)

	gerberStrings2 := make([]string, 0) // where to put strings to be source of the steps

	PrintMemUsage("Memory usage before extracting apertures:")

	for i, s := range gerberStrings {
		/*------------------ aperture blocks processing ----------------- */
		if strings.Compare(s, "%AB*%") == 0 {
			last := len(abopened) - 1
			if last < 0 {
				panic("No more open aperture blocks left!")
			}
			//			apertblocks[abopened[last]].APBodyPtr = append(apertblocks[abopened[last]].APBodyPtr, &gerberStrings[i])

			bapert := new(gerbparser.Aperture)
			bapert.Code = apertblocks[abopened[last]].Code
			bapert.Type = gerbparser.AptypeBlock
			bapert.BlockPtr = apertblocks[abopened[last]]
			bapert.BlockPtr.APStepsPtr = make([]*gerbparser.State, len(bapert.BlockPtr.APBodyPtr)+1)
			bapert.BlockPtr.APStepsPtr[0] = gerbparser.NewStep()
			abopened = abopened[:last]
			apertl.PushBack(bapert) // store correct aperture

			/*
				if last == 0 {
					for i := range apertblocks {
						fmt.Println("Aperture block #", i)
						fmt.Println("Aperture code", apertblocks[i].Code)
						fmt.Println("\tstarts at", apertblocks[i].StartStringNum)
						for j := range apertblocks[i].APBodyPtr {
							fmt.Println("\t\tAB string", j, "value=", apertblocks[i].APBodyPtr[j])
						}
					}
				}
			*/
			continue
		}
		// new block is met
		if strings.HasPrefix(s, gerbparser.GerberApertureBlockDef) &&
			strings.HasSuffix(s, "*%") {
			// aperture block found
			cab := new(gerbparser.BlockAperture)
			cab.StartStringNum = i
			//			cab.APBodyPtr = append(cab.APBodyPtr, &gerberStrings[i])
			cab.Code, err = strconv.Atoi(gerberStrings[i][4 : len(gerberStrings[i])-2])

			/*			cab.APStepsPtr = make([]*gerbparser.State, 0)
						cab.APStepsPtr = append(cab.APStepsPtr, gerbparser.NewStep()) // add root step of this block
			*/
			apertblocks[gerberStrings[i]] = cab
			abopened = append(abopened, gerberStrings[i])
			continue
		}

		if len(abopened) != 0 {
			last := len(abopened) - 1
			apertblocks[abopened[last]].APBodyPtr = append(apertblocks[abopened[last]].APBodyPtr, gerberStrings[i])
			continue
		}
		/*------------------ aperture blocks processing END ----------------- */

		/*------------------ standard apertures processing  ------------------*/
		if strings.HasPrefix(s, gerbparser.GerberApertureDef) &&
			strings.HasSuffix(s, "*%") {
			// possible aperture definition found
			apert = new(gerbparser.Aperture)
			apErr := apert.Init(s[4:len(s)-2], fSpec)
			CheckError(apErr, 500)
			apertl.PushBack(apert) // store correct aperture
			continue
		}

		/*------------------- aperture macro processing ------------------------ */

		// TODO

		// all unprocessed above goes here
		gerberStrings2 = append(gerberStrings2, s)
	}
	// Global array of commands
	gerberStrings = gerberStrings2
	gerberStrings2 = nil
	saveIntermediate(&gerberStrings, "before_steps.txt")
	// Main sequence of steps
	arrayOfSteps = make([]*gerbparser.State, len(gerberStrings)+1)
	// Global Step and Repeat blocks array
	//	SRBlocks = make([]*gerbparser.SR, 0)
	// Global list of Regions
	regl = list.New()

	//  Aperture blocks must be converted to the steps w/o AB
	//  S&R blocks and regions inside each instance of AB added to the global lists!
	for ablock := range apertblocks {
		bsn := CreateStepSequence(&apertblocks[ablock].APBodyPtr, &apertblocks[ablock].APStepsPtr, apertl, regl, fSpec)
		apertblocks[ablock].APStepsPtr = apertblocks[ablock].APStepsPtr[:bsn]
		//		apertblocks[ablock].Print()
	}

	fmt.Println()
	PrintMemUsage("Memory usage before creating main step sequence:")

	// func CreateStepSequence(src *[]string, resSteps *[]*gerbparser.State, apertl *list.List, regl *list.List, fSpec *gerbparser.FormatSpec) (stepnum int)
	stepnum := CreateStepSequence(&gerberStrings, &arrayOfSteps, apertl, regl, fSpec)
	/*
	   /////
	   // state machine current states
	   	var stepnum = 1 // step number
	   	var stepCompleted = true
	   // create the root step with default properties
	   	arrayOfSteps[0] = gerbparser.NewStep()
	   	// process string by string
	   	var step *gerbparser.State
	   	for i, s := range gerberStrings {
	   		if stepCompleted == true {
	   			step = new(gerbparser.State)
	   			*step = *arrayOfSteps[stepnum-1]
	   			step.Coord = nil
	   			step.PrevCoord = nil
	   		}
	   		//		fmt.Printf(">>>>>%v  %v\n", stepnum, arrayOfSteps[stepnum])
	   		procres := step.CreateStep(&gerberStrings[i], arrayOfSteps[stepnum-1], apertl, regl, i, fSpec)
	   		switch procres {
	   		case gerbparser.SCResultNextString:
	   			fallthrough
	   		case gerbparser.SCResultSkipString:
	   			stepCompleted = false
	   			continue
	   		case gerbparser.SCResultStepCmpltd:
	   			step.PrevCoord = arrayOfSteps[stepnum-1].Coord
	   			arrayOfSteps[stepnum] = step
	   			stepnum++
	   			stepCompleted = true
	   			continue
	   		case gerbparser.SCResultStop:
	   			arrayOfSteps[stepnum] = step
	   			step.Coord = arrayOfSteps[stepnum-1].Coord
	   			stepnum++
	   			stepCompleted = true
	   			break
	   		default:
	   			break
	   		}

	   		fmt.Println("Still unknown command: ",s) // print unknown strings
	   	} // end of input strings parsing
	   ////
	*/
	arrayOfSteps = arrayOfSteps[:stepnum]

	fmt.Println("+++++++++++++++++ Unwinded steps ++++++++++++++++")
	/* ------------------ aperture blocks to steps ---------------------------*/
	// each D03 must be checked against aperture block

	PrintMemUsage("Memory usage before unwinding aperture blocks:")

	//	var touch bool = false
	for {
		touch := false
		arrayOfSteps2 := make([]*gerbparser.State, 0)
		for k := 1; k < len(arrayOfSteps); k++ {
			if arrayOfSteps[k].CurrentAp.Type == gerbparser.AptypeBlock &&
				arrayOfSteps[k].Action == gerbparser.OpcodeD03 {
				for i, bs := range arrayOfSteps[k].CurrentAp.BlockPtr.APStepsPtr {
					if i == 0 { // skip root element
						continue
					}
					ns := new(gerbparser.State)
					*ns = *bs
					newxy := new(gerbparser.XY)
					newxy.SetX(bs.Coord.GetX() + arrayOfSteps[k].Coord.GetX())
					newxy.SetY(bs.Coord.GetY() + arrayOfSteps[k].Coord.GetY())
					ns.Coord = newxy
					if i == 1 {
						ns.PrevCoord = arrayOfSteps[k].PrevCoord
					} else {
						ns.PrevCoord = arrayOfSteps2[len(arrayOfSteps2)-1].Coord
					}

					arrayOfSteps2 = append(arrayOfSteps2, ns)
				}
				touch = true
			} else {
				arrayOfSteps2 = append(arrayOfSteps2, arrayOfSteps[k])
			}
			//			arrayOfSteps2[len(arrayOfSteps2)-1].Print()
		}
		arrayOfSteps = arrayOfSteps2
		arrayOfSteps2 = nil
		if touch == false {
			break
		}
	}
	/*
		for i := range arrayOfSteps {
			arrayOfSteps[i].StepNumber = i
			arrayOfSteps[i].Print()
		}
	*/
	// print region info
	j := 0
	for k := regl.Front(); k != nil; k = k.Next() {
		fmt.Printf("%+v\n", k.Value)
		j++

	}
	fmt.Println("Total", j, "regions found.")

	j = 0
	for k := apertl.Front(); k != nil; k = k.Next() {
		fmt.Printf("%+v\n", k.Value)
		j++
	}
	fmt.Println("Total", j, "apertures found.")

	fmt.Println("Total", len(arrayOfSteps)-1, "steps to do.")

	for k := range arrayOfSteps {
		//		fmt.Printf("%+v\n", arrayOfSteps[k])
		if arrayOfSteps[k].Coord.GetX() > maxX {
			maxX = arrayOfSteps[k].Coord.GetX()
		}
		if arrayOfSteps[k].Coord.GetX() < minX {
			minX = arrayOfSteps[k].Coord.GetX()
		}
		if arrayOfSteps[k].Coord.GetY() > maxY {
			maxY = arrayOfSteps[k].Coord.GetY()
		}
		if arrayOfSteps[k].Coord.GetY() < minY {
			minY = arrayOfSteps[k].Coord.GetY()
		}
	}

	PrintMemUsage("Memory usage before rendering:")

	/*
	   let's render the PCB
	*/
	plt = new(plotter.Plotter)
	plt.Init()
	plt.Pen(1)

	render.PlCfg.SetDrawContoursMode()
	render.PlCfg.SetDrawSolidsMode()
	//	render.PlCfg.SetDrawMovesMode()
	render.PlCfg.SetNotDrawMovesMode()
	render.PlCfg.SetDrawOnlyRegionsMode()
	render.PlCfg.SetDrawAllMode()

	context := new(render.Render)

	render.Stat.CircleBresCounter = 0
	render.Stat.LineBresCounter = 0
	fmt.Println("Min. X, Y found:", minX, minY)
	fmt.Println("Max. X, Y found:", maxX, maxY)

	context.Init(plt)
	context.SetMinXY(minX, minY)

	context.Img = image.NewNRGBA(image.Rect(context.LimitsX0, context.LimitsY0, context.LimitsX1, context.LimitsY1))

	//	k := 0
	k := 1
	for {
		if k == len(arrayOfSteps) {
			break
		}
		stepData := arrayOfSteps[k]

		/* unwind step and repeat block(s)*/
		if stepData.SRBlock != nil {
			// first time we've met sr
			kstop := k + stepData.SRBlock.NSteps() // stop value

			//			var srStep *gerbparser.State
			fakeprev := new(gerbparser.XY)
			fakeprev.SetX(0)
			fakeprev.SetY(0)
			modc := make([]gerbparser.State, stepData.SRBlock.NSteps()*stepData.SRBlock.NumX()*stepData.SRBlock.NumY())
			ii := stepData.SRBlock.NumX()
			jj := stepData.SRBlock.NumY()
			modccnt := 0
			var addX, addY float64
			for j := 0; j < jj; j++ {
				addY = float64(j) * arrayOfSteps[k].SRBlock.DY()
				for i := 0; i < ii; i++ {
					addX = float64(i) * arrayOfSteps[k].SRBlock.DX()
					for kk := k; kk < kstop; kk++ {
						//						srStep = arrayOfSteps[kk]
						if kk == k {
							//							srStep.PrevCoord = fakeprev
							modc[modccnt].PrevCoord = fakeprev
						} else {
							modc[modccnt].PrevCoord = modc[modccnt-1].Coord

						}
						modc[modccnt].Action = arrayOfSteps[kk].Action
						modc[modccnt].Region = arrayOfSteps[kk].Region
						modc[modccnt].SRBlock = arrayOfSteps[kk].SRBlock
						modc[modccnt].IpMode = arrayOfSteps[kk].IpMode
						modc[modccnt].QMode = arrayOfSteps[kk].QMode
						modc[modccnt].CurrentAp = arrayOfSteps[kk].CurrentAp
						modc[modccnt].Polarity = arrayOfSteps[kk].Polarity
						modc[modccnt].StepNumber = arrayOfSteps[kk].StepNumber
						modc[modccnt].Coord = new(gerbparser.XY)
						modc[modccnt].Coord.SetX(arrayOfSteps[kk].Coord.GetX() + addX)
						modc[modccnt].Coord.SetY(arrayOfSteps[kk].Coord.GetY() + addY)
						modc[modccnt].Coord.SetI(arrayOfSteps[kk].Coord.GetI())
						modc[modccnt].Coord.SetJ(arrayOfSteps[kk].Coord.GetJ())
						//						srStep.Coord = modc[modccnt].Coord
						//						modc[modccnt].Print()
						context.StepProcessor(&modc[modccnt])
						modccnt++
					}
				}
			}
			k = kstop
			continue
		}

		if stepData.Action == gerbparser.OpcodeStop {
			break
		}

		context.StepProcessor(stepData)
		k++
	}

	fmt.Printf("%s%d%s", "The plotter have drawn ", render.Stat.LineBresCounter, " straight lines using Brezenham\n")
	fmt.Printf("%s%.0f%s", "Total lenght of straight lines = ", render.Stat.LineBresLen*context.XRes, " mm\n")
	fmt.Printf("%s%d%s", "The plotter have drawn ", render.Stat.CircleBresCounter, " circles\n")
	fmt.Printf("%s%.0f%s", "Total lenght of circles = ", render.Stat.CircleLen*context.XRes, " mm\n")
	fmt.Println("The plotter have drawn", render.Stat.FilledRctCounter, "filled rectangles")
	fmt.Println("The plotter have drawn", render.Stat.ObRoundCounter, "obrounds (boxes)")
	fmt.Println("The plotter have moved pen", render.Stat.MovePenCounters, "times")
	fmt.Printf("%s%.0f%s", "Total move distance = ", render.Stat.MovePenDistance*context.XRes, " mm\n")

	// Save to out.png
	f, _ := os.OpenFile("G:\\go_prj\\gerber2em7\\src\\out.png", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()

	PrintMemUsage("Memory usage before  png encoding:")

	png.Encode(f, context.Img)

	plt.Stop()
}

////////////////////////////////////////////////////// end of main ///////////////////////////////////////////////////

// close file
func closeFile(f *os.File) {
	f.Close()
}

// search for format strings
func searchMO() (string, error) {
	err := errors.New("unit of measurements command not found")
	for _, s := range gerberStrings {
		if strings.HasPrefix(s, gerbparser.GerberMOIN) || strings.HasPrefix(s, gerbparser.GerberMOMM) {
			return s, nil
		}
		if strings.Compare(s, "G70*") == 0 {
			return gerbparser.GerberMOIN, nil
		}
		if strings.Compare(s, "G71*") == 0 {
			return gerbparser.GerberMOMM, nil
		}
	}
	return "", err
}

func searchFS() (string, error) {
	for _, s := range gerberStrings {
		if strings.HasPrefix(s, "%FST") {
			return s, errors.New("trailing zero omission format is not supported") // + 09-Jun-2018
		}
		if strings.HasPrefix(s, "%FSLI") {
			return s, errors.New("incremental coordinates ain't supported") // + 09-Jun-2018
		}
		if strings.HasPrefix(s, gerbparser.GerberFormatSpec) {
			return s, nil
		}

	}
	return "", errors.New("_FS_ command not found")
}

/*
func abs(x int) int {
	switch {
	case x >= 0:
		return x
	case x >= MinInt:
		return -x
	}
	panic("math/int.Abs: invalid argument")
}
*/

/*
	Saves intermediate results placed in the strings array to the file
 */
func saveIntermediate(buffer *[]string, fileName string) {

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = file.Truncate(0)
	if err != nil {
		panic(err)
	}
	for i := range *buffer {
		_, err = file.WriteString((*buffer)[i] + "\n")
		if err != nil {
			panic(err)
		}
	}
	file.Sync()
	err = file.Close()
	if err != nil {
		panic(err)
	}
}


func splitStrings(a *[]string) *[]string {
	// split concatenated command strings AAAAAD01*BBBBBBD02*GNN*D03*etc
	splitted := make([]string, 0)
	for i := range *a {
		if (strings.HasPrefix((*a)[i], "%") && strings.HasSuffix((*a)[i], "%")) ||
			strings.HasPrefix((*a)[i], "G04") {
			// do not split
			splitted = append(splitted, (*a)[i])
		} else {
			for _, tspl := range strings.SplitAfter((*a)[i], "*") {
				if len(tspl) > 0 {
					for {
						n := strings.IndexByte(tspl, 'G')
						if n == -1 {
							splitted = append(splitted, tspl)
							break
						} else {
							splitted = append(splitted, tspl[n:n+3]+"*")
							tspl = tspl[n+3:]
						}
					}
				}
			}
		}
	}
	return &splitted
}

func squeezeStrings(a *[]string) *[]string {
	// remove comments and other un-nesessary strings
	// obsolete commands
	// attributes - TODO MAKE USE!!!!
	squeezed := make([]string, 0)
	for i := range *a {
		// strip comments
		if strings.HasPrefix((*a)[i], "G04") || strings.HasPrefix((*a)[i], "G4") { // +09-Jun-2018
			fmt.Println("Comment", (*a)[i], " is found at line", i)
			continue
		}
		// strip some obsolete commands
		if strings.HasPrefix((*a)[i], "%AS") ||
			strings.HasPrefix((*a)[i], "%IR") ||
			strings.HasPrefix((*a)[i], "%MI") ||
			strings.HasPrefix((*a)[i], "%OF") ||
			strings.HasPrefix((*a)[i], "%SF") ||
			strings.HasPrefix((*a)[i], "%IN") ||
			strings.HasPrefix((*a)[i], "%LN") {
			fmt.Println("Obsolete command", (*a)[i], " is found at line", i)
			continue
		}
		if strings.Compare((*a)[i], "%SRX1Y1I0J0*%") == 0 { //  +09-Jun-2018
			continue
		}
		// strip attributes - TODO!!!!!
		if strings.HasPrefix((*a)[i], "%TF") ||
			strings.HasPrefix((*a)[i], "%TA") ||
			strings.HasPrefix((*a)[i], "%TO") ||
			strings.HasPrefix((*a)[i], "%TD") {
			fmt.Println("Attribute", (*a)[i], " is found at line", i)
			continue
		}
		if strings.Compare((*a)[i], "*") == 0 {
			continue
		}
		if strings.Compare((*a)[i], "G54*") == 0 {
			continue
		}
		squeezed = append(squeezed, (*a)[i])
	}
	return &squeezed
}

func CheckError(err error, exitcode int) {
	if err != nil {
		fmt.Println(err)
		os.Exit(exitcode)
	}
}

// this function returns application info
func returnAppInfo(verbLevel int) string {
	var header = "Gerber to EM-7052 translation tool\n"
	var version = "Version 0.0.1\n"
	var progDate = "15-May-2018\n"
	var retVal = "\n"
	switch verbLevel {
	case 3:
		retVal = header + version + progDate
	case 2:
		retVal = header + version
	case 1:
		retVal = header
	default:
		retVal = "\n"
	}
	return retVal
}

// the function creates a full step sequence using src *[]string as source
// src *[]string - pointer to the source string array
// resSteps *[]*gerbparser.State - pointer to the resulting array of the steps, array size must be enough to hold all the staps
// apertl *list.List - pointer to the global aperture list
// regl *list.List - pointer to the global regions list
// fSpec *gerbparser.FormatSpec - pointer to the format specif. object
// stepnum - number of the created steps started from 1

func CreateStepSequence(src *[]string, resSteps *[]*gerbparser.State, apertl *list.List, regl *list.List, fSpec *gerbparser.FormatSpec) (stepnum int) {

	stepnum = 1 // step number
	stepCompleted := true
	// create the root step with default properties
	(*resSteps)[0] = gerbparser.NewStep()
	// process string by string
	var step *gerbparser.State
	for i, s := range *src {
		if stepCompleted == true {
			step = new(gerbparser.State)
			*step = *(*resSteps)[stepnum-1]
			step.Coord = nil
			step.PrevCoord = nil
		}
		//		fmt.Printf(">>>>>%v  %v\n", stepnum, arrayOfSteps[stepnum])
		procres := step.CreateStep( /*&gerberStrings[i]*/ &s, (*resSteps)[stepnum-1], apertl, regl, i, fSpec)
		switch procres {
		case gerbparser.SCResultNextString:
			fallthrough
		case gerbparser.SCResultSkipString:
			stepCompleted = false
			continue
		case gerbparser.SCResultStepCmpltd:
			step.PrevCoord = (*resSteps)[stepnum-1].Coord
			step.StepNumber = stepnum
			(*resSteps)[stepnum] = step
			stepnum++
			stepCompleted = true
			continue
		case gerbparser.SCResultStop:
			step.StepNumber = stepnum
			(*resSteps)[stepnum] = step
			step.Coord = (*resSteps)[stepnum-1].Coord
			stepnum++
			stepCompleted = true
			break
		default:
			break
		}

		fmt.Println("Still unknown command: ", s) // print unknown strings
	} // end of input strings parsing

	return stepnum
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage(header string) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Println(header)
	fmt.Printf("Alloc = %v KB", bToKb(memStats.Alloc))
	fmt.Printf("\tTotalAlloc = %v KB", bToKb(memStats.TotalAlloc))
	fmt.Printf("\tSys = %v KB", bToKb(memStats.Sys))
	fmt.Printf("\tNumGC = %v\n", memStats.NumGC)
}

func bToKb(b uint64) uint64 {
	return b / 1024
}

/* ########################################## EOF #########################################################*/
