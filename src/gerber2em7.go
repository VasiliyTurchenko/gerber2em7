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
	stor "strings_storage"

	"render"
	"runtime"
	"strconv"
	"strings"
)

const (
	MaxInt   = int(^uint(0) >> 1)
	MinInt   = int(-MaxInt - 1)
	MaxInt32 = int32(math.MaxInt32)
	MinInt32 = int32(math.MinInt32)
	MaxInt64 = int64(math.MaxInt64)
	MinInt64 = int64(math.MinInt64)
)

// var ErrbadFS = errors.New("format is not specified or format string parsing error")

var verboselevel = flag.Int("v", 3, "verbose level: 0 - minimal, 3 - maximal")

//var infile = flag.String("i", "", "input file (with path)")
//var outfile = flag.String("o", "", "output file (with path)")
//var logfile = flag.String("l", "", "log file (with path)")

//var totalStrings int = 0

//var gerberStrings []string // all the strings

var gerberStrings *stor.Storage

var fSpec *gerbparser.FormatSpec
var aperture *gerbparser.Aperture

var plotterInstance *plotter.Plotter

var arrayOfSteps []*gerbparser.State // state machine
//var SRBlocks []*gerbparser.SR        // SR blocks
var regionsList *list.List   // regions
var aperturesList *list.List // apertures
var apertureBlocks map[string]*gerbparser.BlockAperture

var maxX, maxY float64 = 0, 0
var minX, minY = 1000000.0, 1000000.0

//var s

func main() {

	/*
		viper.SetConfigName("config")     // no need to include file extension
		viper.AddConfigPath(".")  // set the path of your config file

		err := viper.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		} else {
			fmt.Println(viper.GetString("parser.splittedfile"))

		}


	*/

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

	//	gerberStrings = make([]string, 0)
	gerberStrings = stor.NewStorage()

	fSpec = new(gerbparser.FormatSpec)

	inFile, err := os.Open("G:\\go_prj\\gerber2em7\\src\\test.g2") // For read access.
	defer inFile.Close()
	CheckError(err, -1) // read the file into the array of strings

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		var splittedString *[]string
		//		gerberStrings = append(gerberStrings, strings.ToUpper(scanner.Text()))

		rawString := strings.ToUpper(scanner.Text())
		// split concatenated command strings AAAAAD01*BBBBBBD02*GNN*D03*etc
		splittedString = splitString(rawString)
		// feed the storage
		for _, str := range *splittedString {
			gerberStrings.Accept(squeezeString(str))
		}
	}
	// save splitted strings to a file
	saveIntermediate(gerberStrings, "splitted.txt")

	// search for format definition strings
	mo, err := searchMO(gerberStrings)
	CheckError(err, 300)

	fs, err := searchFS(gerberStrings)
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
	aperturesList = list.New()
	apertureBlocks = make(map[string]*gerbparser.BlockAperture)
	apertureBlockOpened := make([]string, 0)

	//	gerberStrings2 := make([]string, 0) // where to put strings to be source of the steps

	gerberStrings2 := stor.NewStorage()

	PrintMemUsage("Memory usage before extracting apertures:")

	// Aperture processing loop
	gerberStrings.ResetPos()
	//	for i, gerberString := range gerberStrings {

	for {
		i := gerberStrings.PeekPos()
		gerberString := gerberStrings.String()

		if len(gerberString) == 0 {
			break
		}

		//fmt.Println("i=", i)

		// aperture blocks processing
		if strings.Compare(gerberString, gerbparser.GerberApertureBlockDefEnd) == 0 {
			last := len(apertureBlockOpened) - 1
			if last < 0 {
				panic("No more open aperture blocks left!")
			}
			aperture := new(gerbparser.Aperture)
			aperture.Code = apertureBlocks[apertureBlockOpened[last]].Code
			aperture.Type = gerbparser.AptypeBlock
			aperture.BlockPtr = apertureBlocks[apertureBlockOpened[last]]
			aperture.BlockPtr.APStepsPtr = make([]*gerbparser.State, len(aperture.BlockPtr.APBodyPtr)+1)
			aperture.BlockPtr.APStepsPtr[0] = gerbparser.NewStep()
			apertureBlockOpened = apertureBlockOpened[:last]
			aperturesList.PushBack(aperture) // store correct aperture
			continue
		}
		// new block is met
		if strings.HasPrefix(gerberString, gerbparser.GerberApertureBlockDef) &&
			strings.HasSuffix(gerberString, "*%") {
			// aperture block found
			apBlk := new(gerbparser.BlockAperture)
			apBlk.StartStringNum = i
			apBlk.Code, err = strconv.Atoi(gerberString[4 : len(gerberString)-2])
			apertureBlocks[gerberString] = apBlk
			apertureBlockOpened = append(apertureBlockOpened, gerberString)
			continue
		}

		if len(apertureBlockOpened) != 0 {
			last := len(apertureBlockOpened) - 1
			apertureBlocks[apertureBlockOpened[last]].APBodyPtr = append(apertureBlocks[apertureBlockOpened[last]].APBodyPtr, gerberString)
			continue
		}
		/*------------------ aperture blocks processing END ----------------- */

		/*------------------ standard apertures processing  ------------------*/
		if strings.HasPrefix(gerberString, gerbparser.GerberApertureDef) &&
			strings.HasSuffix(gerberString, "*%") {
			// possible aperture definition found
			aperture = new(gerbparser.Aperture)
			apErr := aperture.Init(gerberString[4:len(gerberString)-2], fSpec)
			CheckError(apErr, 500)
			aperturesList.PushBack(aperture) // store correct aperture
			continue
		}

		/*------------------- aperture macro processing ------------------------ */

		// TODO

		// all unprocessed above goes here
		gerberStrings2.Accept(gerberString)
	}

	// Global array of commands
	gerberStrings = gerberStrings2
	gerberStrings2 = nil

	saveIntermediate(gerberStrings, "before_steps.txt")

	// Main sequence of steps
	arrayOfSteps = make([]*gerbparser.State, gerberStrings.Len()+1) //len(gerberStrings)+1)
	// Global Step and Repeat blocks array
	//	SRBlocks = make([]*gerbparser.SR, 0)
	// Global list of Regions
	regionsList = list.New()

	//  Aperture blocks must be converted to the steps w/o AB
	//  S&R blocks and regions inside each instance of AB added to the global lists!
	for ablock := range apertureBlocks {
		bsn := CreateStepSequence(&apertureBlocks[ablock].APBodyPtr, &apertureBlocks[ablock].APStepsPtr, aperturesList, regionsList, fSpec)
		apertureBlocks[ablock].APStepsPtr = apertureBlocks[ablock].APStepsPtr[:bsn]
		//		apertureBlocks[ablock].Print()
	}

	fmt.Println()
	PrintMemUsage("Memory usage before creating main step sequence:")

	// patch
	// TODO get rid of the patch!

	gerberStringsArray := gerberStrings.ToArray()

	// func CreateStepSequence(src *[]string, resSteps *[]*gerbparser.State, aperturesList *list.List, regionsList *list.List, fSpec *gerbparser.FormatSpec) (stepnum int)
	stepnum := CreateStepSequence(&gerberStringsArray, &arrayOfSteps, aperturesList, regionsList, fSpec)
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
	   		procres := step.CreateStep(&gerberStrings[i], arrayOfSteps[stepnum-1], aperturesList, regionsList, i, fSpec)
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
	for k := regionsList.Front(); k != nil; k = k.Next() {
		fmt.Printf("%+v\n", k.Value)
		j++

	}
	fmt.Println("Total", j, "regions found.")

	j = 0
	for k := aperturesList.Front(); k != nil; k = k.Next() {
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
	//	plotterInstance = new(plotter.Plotter)
	//	plotterInstance.Init()

	plotterInstance = plotter.NewPlotter()
	plotterInstance.Pen(1)

	render.PlCfg.SetDrawContoursMode()
	render.PlCfg.SetDrawSolidsMode()
	//	render.PlCfg.SetDrawMovesMode()
	render.PlCfg.SetNotDrawMovesMode()
	render.PlCfg.SetDrawOnlyRegionsMode()
	render.PlCfg.SetDrawAllMode()

	renderContext := new(render.Render)

	render.Stat.CircleBresCounter = 0
	render.Stat.LineBresCounter = 0
	fmt.Println("Min. X, Y found:", minX, minY)
	fmt.Println("Max. X, Y found:", maxX, maxY)

	renderContext.Init(plotterInstance)
	renderContext.SetMinXY(minX, minY)

	renderContext.Img = image.NewNRGBA(image.Rect(renderContext.LimitsX0, renderContext.LimitsY0, renderContext.LimitsX1, renderContext.LimitsY1))

	//	k := 0
	k := 1
	for k < len(arrayOfSteps) {
		stepToDo := arrayOfSteps[k]

		/* unwind step and repeat block(s)*/
		if stepToDo.SRBlock != nil {
			// first time we've met sr
			kStop := k + stepToDo.SRBlock.NSteps() // stop value

			//			var srStep *gerbparser.State
			fakePrev := new(gerbparser.XY)
			fakePrev.SetX(0)
			fakePrev.SetY(0)
			modc := make([]gerbparser.State, stepToDo.SRBlock.NSteps()*stepToDo.SRBlock.NumX()*stepToDo.SRBlock.NumY())
			ii := stepToDo.SRBlock.NumX()
			jj := stepToDo.SRBlock.NumY()
			modccnt := 0
			var addX, addY float64
			for j := 0; j < jj; j++ {
				addY = float64(j) * arrayOfSteps[k].SRBlock.DY()
				for i := 0; i < ii; i++ {
					addX = float64(i) * arrayOfSteps[k].SRBlock.DX()
					for kk := k; kk < kStop; kk++ {
						//						srStep = arrayOfSteps[kk]
						if kk == k {
							//							srStep.PrevCoord = fakePrev
							modc[modccnt].PrevCoord = fakePrev
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
						renderContext.StepProcessor(&modc[modccnt])
						modccnt++
					}
				}
			}
			k = kStop
			continue
		}

		if stepToDo.Action == gerbparser.OpcodeStop {
			break
		}

		renderContext.StepProcessor(stepToDo)
		k++
	}

	fmt.Printf("%s%d%s", "The plotter have drawn ", render.Stat.LineBresCounter, " straight lines using Brezenham\n")
	fmt.Printf("%s%.0f%s", "Total lenght of straight lines = ", render.Stat.LineBresLen*renderContext.XRes, " mm\n")
	fmt.Printf("%s%d%s", "The plotter have drawn ", render.Stat.CircleBresCounter, " circles\n")
	fmt.Printf("%s%.0f%s", "Total lenght of circles = ", render.Stat.CircleLen*renderContext.XRes, " mm\n")
	fmt.Println("The plotter have drawn", render.Stat.FilledRctCounter, "filled rectangles")
	fmt.Println("The plotter have drawn", render.Stat.ObRoundCounter, "obrounds (boxes)")
	fmt.Println("The plotter have moved pen", render.Stat.MovePenCounters, "times")
	fmt.Printf("%s%.0f%s", "Total move distance = ", render.Stat.MovePenDistance*renderContext.XRes, " mm\n")

	// Save to out.png
	f, _ := os.OpenFile("G:\\go_prj\\gerber2em7\\src\\out.png", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()

	PrintMemUsage("Memory usage before  png encoding:")

	png.Encode(f, renderContext.Img)

	plotterInstance.Stop()
}

////////////////////////////////////////////////////// end of main ///////////////////////////////////////////////////

// close file
func closeFile(f *os.File) {
	f.Close()
}

// search for format strings
func searchMO(storage *stor.Storage) (string, error) {
	err := errors.New("unit of measurements command not found")
	storage.ResetPos()
	//	for _, s := range gerberStrings {
	s := storage.String()
	for len(s) > 0 {

		if strings.HasPrefix(s, gerbparser.GerberMOIN) || strings.HasPrefix(s, gerbparser.GerberMOMM) {
			return s, nil
		}
		if strings.Compare(s, "G70*") == 0 {
			return gerbparser.GerberMOIN, nil
		}
		if strings.Compare(s, "G71*") == 0 {
			return gerbparser.GerberMOMM, nil
		}
		s = storage.String()
	}
	return "", err
}

func searchFS(storage *stor.Storage) (string, error) {

	storage.ResetPos()
	//	for _, s := range gerberStrings {

	s := storage.String()
	for len(s) > 0 {
		if strings.HasPrefix(s, "%FST") {
			return s, errors.New("trailing zero omission format is not supported") // + 09-Jun-2018
		}
		if strings.HasPrefix(s, "%FSLI") {
			return s, errors.New("incremental coordinates ain't supported") // + 09-Jun-2018
		}
		if strings.HasPrefix(s, gerbparser.GerberFormatSpec) {
			return s, nil
		}
		s = storage.String()
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
	Saves intermediate results from the strings storage to the file
*/
func saveIntermediate(storage *stor.Storage, fileName string) {

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = file.Truncate(0)
	if err != nil {
		panic(err)
	}
	for {
		//	for i := range *buffer {
		str := storage.String()
		if len(str) == 0 {
			storage.ResetPos()
			break
		}
		_, err = file.WriteString(str + "\n")
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

func splitString(rawString string) *[]string {
	// split concatenated command string AAAAAD01*BBBBBBD02*GNN*D03*etc
	splittedStrings := make([]string, 0)
	if (strings.HasPrefix(rawString, "%") && strings.HasSuffix(rawString, "%")) ||
		strings.HasPrefix(rawString, "G04") {
		// do not split
		splittedStrings = append(splittedStrings, rawString)
	} else {
		for _, tmpSplitted := range strings.SplitAfter(rawString, "*") {
			if len(tmpSplitted) > 0 {
				for {
					n := strings.IndexByte(tmpSplitted, 'G')
					if n == -1 {
						splittedStrings = append(splittedStrings, tmpSplitted)
						break
					} else {
						splittedStrings = append(splittedStrings, tmpSplitted[n:n+3]+"*")
						tmpSplitted = tmpSplitted[n+3:]
					}
				}
			}
		}
	}
	return &splittedStrings
}

func squeezeString(inString string) string {
	// remove comments and other un-nesessary strings
	// obsolete commands
	// attributes - TODO MAKE USE!!!!
	// strip comments
	if strings.HasPrefix(inString, "G04") || strings.HasPrefix(inString, "G4") { // +09-Jun-2018
		fmt.Println("Comment", inString, " is found")
		return ""
	}
	// strip some obsolete commands
	if strings.HasPrefix(inString, "%AS") ||
		strings.HasPrefix(inString, "%IR") ||
		strings.HasPrefix(inString, "%MI") ||
		strings.HasPrefix(inString, "%OF") ||
		strings.HasPrefix(inString, "%SF") ||
		strings.HasPrefix(inString, "%IN") ||
		strings.HasPrefix(inString, "%LN") {
		fmt.Println("Obsolete command", inString, " is found")
		return ""
	}
	if strings.Compare(inString, "%SRX1Y1I0J0*%") == 0 { //  +09-Jun-2018
		return ""
	}
	// strip attributes - TODO!!!!!
	if strings.HasPrefix(inString, "%TF") ||
		strings.HasPrefix(inString, "%TA") ||
		strings.HasPrefix(inString, "%TO") ||
		strings.HasPrefix(inString, "%TD") {
		fmt.Println("Attribute", inString, " is found")
		return ""
	}
	if strings.Compare(inString, "*") == 0 {
		return ""
	}
	if strings.Compare(inString, "G54*") == 0 {
		return ""
	}
	return inString
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
// aperturesList *list.List - pointer to the global aperture list
// regionsList *list.List - pointer to the global regions list
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
