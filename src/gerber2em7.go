// Copyright 2018 Vasily Turchenko <turchenkov@gmail.com>. All rights reserved.
// Use of this source code is free

package main

import (
	"blockapertures"
	"bufio"
	"configurator"
	"container/list"
	"errors"
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"image/color"

	"image/png"
	"os"
	"runtime"
	"strconv"
	"strings"
	stor "strings_storage"
	"time"
)

import (
	. "xy"
	. "gerberbasetypes"
	"plotter"
	"render"
	"gerberstates"
	"apertures"
)

// TODO get rid of it
var verboseLevel = flag.Int("v", 3, "verbose level: 0 - minimal, 3 - maximal")

var (

	// configuration base
	viperConfig *viper.Viper

	// global storage of input gerber file strings, the source to feed some processors
	gerberStrings *stor.Storage

	// plotter instance which is responsible for generating the command stream for the target device
	plotterInstance *plotter.Plotter

	// array of steps to be executed to generate PCB
	arrayOfSteps []*gerberstates.State

	// the list of regions
	regionsList *list.List

	// the list of all the apertures
	aperturesList *list.List

	// the map consisting all the aperture blocks
	apertureBlocks map[string]*blockapertures.BlockAperture

	// format specification for the gerber file
	fSpec *FormatSpec

	//render context
	renderContext *render.Render
)

func main() {

	fmt.Println(returnAppInfo(*verboseLevel))

	viperConfig = viper.New()
	configurator.SetDefaults(viperConfig)

	//	configurator.DiagnosticAllCfgPrint(viperConfig)

	cfgFileError := configurator.ProcessConfigFile(viperConfig)
	if cfgFileError != nil {
		fmt.Print("An error has occured: ")
		fmt.Println(cfgFileError)
		fmt.Println("Using built-in defaults.\n")
		configurator.SetDefaults(viperConfig)
	}

	//	configurator.DiagnosticAllCfgPrint(viperConfig)

	var sourceFileName string
	flag.StringVar(&sourceFileName, "i", "", "input file")
	flag.Parse()
	if len(sourceFileName) == 0 {
		fmt.Println("No input file specified.\nUsage:")
		flag.PrintDefaults()
		os.Exit(-1)
	}
	timeStamp := time.Now()
	timeInfo(timeStamp)
	fmt.Println("input file:", sourceFileName, "\n")

	/*
	   Process input string
	*/
	printMemUsage("Memory usage before reading input file:")

	gerberStrings = stor.NewStorage()

	fSpec = new(FormatSpec)

	inFile, err := os.Open(sourceFileName) // For read access.
	defer inFile.Close()
	checkError(err, -1) // read the file into the array of strings

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		var splittedString *[]string
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
	checkError(err, 300)

	fs, err := searchFS(gerberStrings)
	checkError(err, 301)

	fSpec = new(FormatSpec)
	if fSpec.Init(fs, mo) == false {
		fmt.Println("Can not parse:")
		fmt.Println(fs)
		fmt.Println(mo)
		os.Exit(302)
	}

	/* ---------------------- extract apertures and aperture blocks  --------------------- */
	// and aperture macros - TODO!!!!!
	aperturesList = list.New()
	apertureBlocks = make(map[string]*blockapertures.BlockAperture)
	apertureBlockOpened := make([]string, 0)

	gerberStrings2 := stor.NewStorage()

	printMemUsage("Memory usage before extracting apertures:")

	// Aperture processing loop
	gerberStrings.ResetPos()
	for {
		i := gerberStrings.PeekPos()
		gerberString := gerberStrings.String()
		if len(gerberString) == 0 {
			break
		}
		// aperture blocks processing
		if strings.Compare(gerberString, GerberApertureBlockDefEnd) == 0 {
			last := len(apertureBlockOpened) - 1
			if last < 0 {
				panic("No more open aperture blocks left!")
			}
			aperture := new(apertures.Aperture)
			aperture.Code = apertureBlocks[apertureBlockOpened[last]].Code
			aperture.Type = AptypeBlock
			aperture.BlockPtr = apertureBlocks[apertureBlockOpened[last]]
			aperture.BlockPtr.StepsPtr = make([]*gerberstates.State, len(aperture.BlockPtr.BodyStrings)+1)
			aperture.BlockPtr.StepsPtr[0] = gerberstates.NewState()
			apertureBlockOpened = apertureBlockOpened[:last]
			aperturesList.PushBack(aperture) // store correct aperture
			continue
		}
		// new block is met
		if strings.HasPrefix(gerberString, GerberApertureBlockDef) &&
			strings.HasSuffix(gerberString, "*%") {
			// aperture block found
			apBlk := new(blockapertures.BlockAperture)
			apBlk.StartStringNum = i
			apBlk.Code, err = strconv.Atoi(gerberString[4 : len(gerberString)-2])
			apertureBlocks[gerberString] = apBlk
			apertureBlockOpened = append(apertureBlockOpened, gerberString)
			continue
		}

		if len(apertureBlockOpened) != 0 {
			last := len(apertureBlockOpened) - 1
			apertureBlocks[apertureBlockOpened[last]].BodyStrings = append(apertureBlocks[apertureBlockOpened[last]].BodyStrings, gerberString)
			continue
		}
		/*------------------ aperture blocks processing END ----------------- */

		/*------------------ standard apertures processing  ------------------*/
		if strings.HasPrefix(gerberString, GerberApertureDef) &&
			strings.HasSuffix(gerberString, "*%") {
			// possible aperture definition found
			aperture := new(apertures.Aperture)
			apErr := aperture.Init(gerberString[4:len(gerberString)-2], fSpec)
			checkError(apErr, 500)
			aperturesList.PushBack(aperture) // store correct aperture
			continue
		}
		/*------------------- aperture macro processing ------------------------ */
		// TODO
		// all unprocessed above goes here
		gerberStrings2.Accept(gerberString)
	}

	// Global array of commands
	gerberStrings, gerberStrings2 = gerberStrings2, nil

	saveIntermediate(gerberStrings, "before_steps.txt")

	// Main sequence of steps
	arrayOfSteps = make([]*gerberstates.State, gerberStrings.Len()+1)
	// Global list of Regions
	regionsList = list.New()

	//  Aperture blocks must be converted to the steps w/o AB
	//  S&R blocks and regions inside each instance of AB added to the global lists!
	for apBlock := range apertureBlocks {
		bsn := createStepSequence(&apertureBlocks[apBlock].BodyStrings, &apertureBlocks[apBlock].StepsPtr, aperturesList, regionsList, fSpec)
		apertureBlocks[apBlock].StepsPtr = apertureBlocks[apBlock].StepsPtr[:bsn]
	}

	fmt.Println()
	printMemUsage("Memory usage before creating main step sequence:")

	// patch
	// TODO get rid of the patch!
	gerberStringsArray := gerberStrings.ToArray()

	numberOfSteps := createStepSequence(&gerberStringsArray, &arrayOfSteps, aperturesList, regionsList, fSpec)
	arrayOfSteps = arrayOfSteps[:numberOfSteps]

	/* ------------------ aperture blocks to steps ---------------------------*/
	// each D03 must be checked against aperture block

	printMemUsage("Memory usage before unwinding aperture blocks:")

	//	var touch bool = false
	for {
		touch := false
		arrayOfSteps2 := make([]*gerberstates.State, 0)
		for k := 1; k < len(arrayOfSteps); k++ {
			if arrayOfSteps[k].CurrentAp != nil &&
				arrayOfSteps[k].CurrentAp.Type == AptypeBlock &&
				arrayOfSteps[k].Action == OpcodeD03_FLASH {
				for i, bs := range arrayOfSteps[k].CurrentAp.BlockPtr.StepsPtr {
					if i == 0 { // skip root element
						continue
					}
					newStep := new(gerberstates.State)
					newStep.CopyOfWithOffset(bs, arrayOfSteps[k].Coord.GetX(), arrayOfSteps[k].Coord.GetY())
					if i == 1 {
						newStep.PrevCoord = arrayOfSteps[k].PrevCoord
					} else {
						newStep.PrevCoord = arrayOfSteps2[len(arrayOfSteps2)-1].Coord
					}
					arrayOfSteps2 = append(arrayOfSteps2, newStep)
				}
				touch = true
			} else {
				arrayOfSteps2 = append(arrayOfSteps2, arrayOfSteps[k])
			}
		}
		arrayOfSteps, arrayOfSteps2 = arrayOfSteps2, nil
		if touch == false {
			break
		}
	}

	// unwinding SR blocks
	printMemUsage("Memory usage before unwinding SR blocks:")

	// unwind SR blocks
	i := 0
	for i < len(arrayOfSteps) {
		if arrayOfSteps[i].SRBlock != nil {
			insert, excludeLen := unwindSRBlock(&arrayOfSteps, i)
			tailI := i + excludeLen
			tail := arrayOfSteps[tailI:]
			arrayOfSteps = arrayOfSteps[:i]
			arrayOfSteps = append(arrayOfSteps, *insert...)
			arrayOfSteps = append(arrayOfSteps, tail...)
			i += len(*insert)
		} else {
			i++
		}
	}

	// print regions info
	if viperConfig.GetBool(configurator.CfgCommonPrintRegionsInfo) == true {
		j := 0
		for k := regionsList.Front(); k != nil; k = k.Next() {
			fmt.Printf("%+v\n", k.Value)
			j++
		}
		fmt.Println("Total", j, "regions found.")
	}
	// print apertures info
	if viperConfig.GetBool(configurator.CfgCommonPrintAperturesInfo) == true {
		j := 0
		for k := aperturesList.Front(); k != nil; k = k.Next() {
			fmt.Printf("%+v\n", k.Value)
			j++
		}
		fmt.Println("Total", j, "apertures found.")
	}

	fmt.Println("Total", len(arrayOfSteps)-1, "steps to do.")

	var maxX, maxY float64 = 0, 0
	var minX, minY = 1000000.0, 1000000.0
	for k := range arrayOfSteps {
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

	printMemUsage("Memory usage before rendering:")
	timeInfo(timeStamp)
	fmt.Println("Rendering process started\n")

	/*
	   let's render the PCB
	*/
	plotterInstance = plotter.NewPlotter()
	plotterInstance.TakePen(1)
	plotterInstance.SetOutFileName(viperConfig.GetString(configurator.CfgPlotterOutFile))
	renderContext = render.NewRender(plotterInstance, viperConfig, minX, minY, maxX, maxY)
	fmt.Printf("Min. X, Y found: (%f,%f)\n", minX, minY)
	fmt.Printf("Max. X, Y found: (%f,%f)\n", maxX, maxY)
	//renderContext.SetMinXY(minX, minY)
	//renderContext.SetMaxXY(maxX, maxY)

	printMemUsage("Memory usage after render context was initialized:")

	// draw frame by dashed line
	renderContext.DrawFrame()

	k := 0
	for k < len(arrayOfSteps) {
		if arrayOfSteps[k].Action == OpcodeStop {
			break
		}
		ProcessStep(arrayOfSteps[k])
		k++
	}

	if viperConfig.GetBool(configurator.CfgCommonPrintStatistic) == true {
		fmt.Printf("%s%d%s", "The plotter have drawn ", renderContext.LineBresCounter, " straight lines using Brezenham\n")
		fmt.Printf("%s%.0f%s", "Total lenght of straight lines = ", renderContext.LineBresLen*renderContext.XRes, " mm\n")
		fmt.Printf("%s%d%s", "The plotter have drawn ", renderContext.CircleBresCounter, " circles\n")
		fmt.Printf("%s%.0f%s", "Total lenght of circles = ", renderContext.CircleLen*renderContext.XRes, " mm\n")
		fmt.Println("The plotter have drawn", renderContext.FilledRctCounter, "filled rectangles")
		fmt.Println("The plotter have drawn", renderContext.ObRoundCounter, "obrounds (boxes)")
		fmt.Println("The plotter have moved pen", renderContext.MovePenCounters, "times")
		fmt.Printf("%s%.0f%s", "Total move distance = ", renderContext.MovePenDistance*renderContext.XRes, " mm\n")

	}

	if renderContext.YNeedsFlip == true {
		timeInfo(timeStamp)
		fmt.Println("Started flipping (only png image) over X-axis")
		imgLines := renderContext.Img.Bounds().Max.Y - renderContext.Img.Bounds().Min.Y
		pixelsInLine := renderContext.Img.Bounds().Max.X - renderContext.Img.Bounds().Min.X
		steps := imgLines / 2
		for j := 0; j < steps; j++ {
			for i := 0; i < pixelsInLine; i++ {
				tmp := renderContext.Img.At(i, j)
				renderContext.Img.Set(i, j, renderContext.Img.At(i, imgLines-j-1))
				renderContext.Img.Set(i, imgLines-j-1, tmp)
			}
		}
	}

	timeInfo(timeStamp)
	fmt.Println("Rendering process finished")

	// Save to out.png
	if viperConfig.GetBool(configurator.CfgRendererGeneratePNG) == true {
		printMemUsage("Memory usage before png encoding:")
		timeInfo(timeStamp)
		fmt.Println("Generating png image ", renderContext.Img.Bounds().String())
		f, _ := os.OpenFile("G:\\go_prj\\gerber2em7\\src\\out.png", os.O_WRONLY|os.O_CREATE, 0600)
		defer f.Close()
		png.Encode(f, renderContext.Img)
		timeInfo(timeStamp)
		fmt.Println("Image is saved to the file", viperConfig.GetString(configurator.CfgRendererOutFile))
		printMemUsage("Memory usage after png encoding:")
	}
	timeInfo(timeStamp)
	fmt.Println("Saving plotter commands stream to file")
	plotterInstance.Stop()
	timeInfo(timeStamp)
	fmt.Println("Plotter commands are saved to the file", viperConfig.GetString(configurator.CfgPlotterOutFile))
	timeInfo(timeStamp)
	fmt.Println("Exiting")
}

////////////////////////////////////////////////////// end of main ///////////////////////////////////////////////////

// search for format strings
func searchMO(storage *stor.Storage) (string, error) {
	err := errors.New("unit of measurements command not found")
	storage.ResetPos()
	//	for _, s := range gerberStrings {
	s := storage.String()
	for len(s) > 0 {

		if strings.HasPrefix(s, GerberMOIN) || strings.HasPrefix(s, GerberMOMM) {
			return s, nil
		}
		if strings.Compare(s, "G70*") == 0 {
			return GerberMOIN, nil
		}
		if strings.Compare(s, "G71*") == 0 {
			return GerberMOMM, nil
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
		if strings.HasPrefix(s, GerberFormatSpec) {
			return s, nil
		}
		s = storage.String()
	}
	return "", errors.New("_FS_ command not found")
}

/*
	Saves intermediate results from the strings storage to the file
*/
func saveIntermediate(storage *stor.Storage, fileName string) {

	if viperConfig.GetBool(configurator.CfgParserSaveIntermediate) == false {
		return
	}

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = file.Truncate(0)
	if err != nil {
		panic(err)
	}
	storage.ResetPos()
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
func printSqueezedOut(str string) {
	if viperConfig.GetBool(configurator.CfgCommonPrintGerberComments) == true {
		fmt.Println(str)
	}
	return
}

func squeezeString(inString string) string {
	// remove comments and other un-nesessary strings
	// obsolete commands
	// attributes - TODO MAKE USE!!!!
	// strip comments
	if strings.HasPrefix(inString, "G04") || strings.HasPrefix(inString, "G4") { // +09-Jun-2018
		printSqueezedOut("Comment " + inString + " is found")
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
		printSqueezedOut("Obsolete command " + inString + " is found")
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
		printSqueezedOut("Attribute " + inString + " is found")
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


// this function returns application info
func returnAppInfo(verbLevel int) string {
	var header = "Gerber to EM-7052 translation tool\n"
	var version = "Version 0.1.0\n"
	var progDate = "07-Sep-2018\n"
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
// NumberOfSteps - number of the created steps started from 1

func createStepSequence(src *[]string,
	resSteps *[]*gerberstates.State,
	apertl *list.List,
	regl *list.List,
	fSpec *FormatSpec) (NumberOfSteps int) {

	stepNumber := 1 // step number
	stepCompleted := true
	// create the root step with default properties
	(*resSteps)[0] = gerberstates.NewState()
	// process string by string
	var step *gerberstates.State
	for i, s := range *src {
		if stepCompleted == true {
			step = new(gerberstates.State)
			*step = *(*resSteps)[stepNumber-1]
			step.Coord = nil
			step.PrevCoord = nil
		}
		//		fmt.Printf(">>>>>%v  %v\n", stepNumber, arrayOfSteps[stepNumber])
		createStepResult := step.CreateStep(&s, (*resSteps)[stepNumber-1], apertl, regl, i, fSpec)
		switch createStepResult {
		case gerberstates.SCResultNextString:
			fallthrough
		case gerberstates.SCResultSkipString:
			stepCompleted = false
			continue
		case gerberstates.SCResultStepCompleted:
			step.PrevCoord = (*resSteps)[stepNumber-1].Coord
			step.StepNumber = stepNumber
			(*resSteps)[stepNumber] = step
			stepNumber++
			stepCompleted = true
			continue
		case gerberstates.SCResultStop:
			step.StepNumber = stepNumber
			(*resSteps)[stepNumber] = step
			step.Coord = (*resSteps)[stepNumber-1].Coord
			stepNumber++
			stepCompleted = true
			break
		default:
			break
		}
		fmt.Println("Still unknown command: ", s) // print unknown strings
	} // end of input strings parsing
	return stepNumber
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garbage collection cycles completed.
func printMemUsage(header string) {
	if viperConfig.GetBool(configurator.CfgCommonPrintMemoryInfo) == false {
		return
	}

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

func timeInfo(prev time.Time) {
	now := time.Now()
	elapsed := time.Since(prev)
	/*
		"[23:59:04 +2.00001] "
	*/
	out := "["
	hr := strconv.Itoa(now.Hour())
	if len(hr) == 1 {
		hr = "0" + hr
	}
	min := strconv.Itoa(now.Minute())
	if len(min) == 1 {
		min = "0" + min
	}
	sec := strconv.Itoa(now.Second())
	if len(sec) == 1 {
		sec = "0" + sec
	}

	out = out + hr + ":" + min + ":" + sec + " +"
	elapsedSec := (float64(elapsed.Nanoseconds() / (1000 * 1000))) / 1000.0
	out = out + strconv.FormatFloat(elapsedSec, 'f', 3, 64) + "] "
	fmt.Print(out)
}

func unwindSRBlock(steps *[]*gerberstates.State, k int) (*[]*gerberstates.State, int) {
	firstSRStep := (*steps)[k]
	// once came into, no return until sr block stays not fully processed
	kStop := k + firstSRStep.SRBlock.NSteps() // stop value
	numXSteps := firstSRStep.SRBlock.NumX()
	numYSteps := firstSRStep.SRBlock.NumY()
	numberOfStepsInSRBlock := firstSRStep.SRBlock.NSteps() * numXSteps * numYSteps
	SRBlockSteps := make([]*gerberstates.State, numberOfStepsInSRBlock)
	stepCounter := 0
	var addX, addY float64
	for j := 0; j < numYSteps; j++ {
		addY = float64(j) * firstSRStep.SRBlock.DY()
		for i := 0; i < numXSteps; i++ {
			addX = float64(i) * firstSRStep.SRBlock.DX()
			for kk := k; kk < kStop; kk++ {
				SRBlockSteps[stepCounter] = gerberstates.NewState()
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

/*
**************************** step processor *******************************
 */
func ProcessStep(stepData *gerberstates.State) {

	//	stepData.Print()

	var Xp int
	var Yp int
	Xc := transformCoord(stepData.Coord.GetX()-renderContext.MinX, renderContext.XRes)
	Yc := transformCoord(stepData.Coord.GetY()-renderContext.MinY, renderContext.YRes)
	if stepData.PrevCoord == nil {
		Xp = transformCoord(0-renderContext.MinX, renderContext.XRes)
		Yp = transformCoord(0-renderContext.MinY, renderContext.YRes)
	} else {
		Xp = transformCoord(stepData.PrevCoord.GetX()-renderContext.MinX, renderContext.XRes)
		Yp = transformCoord(stepData.PrevCoord.GetY()-renderContext.MinY, renderContext.YRes)
	}

	if stepData.Region != nil {
		// process region
		if renderContext.PolygonPtr == nil {
			renderContext.PolygonPtr = render.NewPolygon()
		}
		if renderContext.AddStepToPolygon(stepData) == stepData.Region.GetNumXY() {
			// we can process region
			renderContext.RenderPolygon()
			renderContext.PolygonPtr = nil
		}
	} else {
		var stepColor color.RGBA
		switch stepData.Action {
		case OpcodeD01_DRAW: // draw
			if stepData.Polarity == PolTypeDark {
				stepColor = renderContext.LineColor
			} else {
				stepColor = renderContext.ClearColor
			}

			var apertureSize int
			if abs(Xc-Xp) < (4*renderContext.PointSizeI) && abs(Yc-Yp) < (4*renderContext.PointSizeI) {
				stepData.IpMode = IPModeLinear
			}
			if stepData.IpMode == IPModeLinear {
				// linear interpolation
				if renderContext.DrawOnlyRegionsMode != true {
					if stepData.CurrentAp.Type == AptypeCircle {
						apertureSize = transformCoord(stepData.CurrentAp.Diameter, renderContext.XRes)
						renderContext.DrawByCircleAperture(Xp, Yp, Xc, Yc, apertureSize, stepColor)
					} else if stepData.CurrentAp.Type == AptypeRectangle {
						// draw with rectangle aperture
						w := transformCoord(stepData.CurrentAp.XSize, renderContext.XRes)
						h := transformCoord(stepData.CurrentAp.YSize, renderContext.YRes)
						renderContext.DrawByRectangleAperture(Xp, Yp, Xc, Yc, w, h, stepColor)
					} else {
						fmt.Println("Error. Only solid drawCircle and solid rectangle may be used to draw.")
						break
					}
				}
			} else {
				// non-linear interpolation
				if renderContext.DrawOnlyRegionsMode != true {
					if stepData.CurrentAp.Type == AptypeCircle {
						apertureSize = transformCoord(stepData.CurrentAp.Diameter, renderContext.XRes)
						var (
							fXp, fYp float64
						)
						if stepData.PrevCoord == nil {
							fXp = transformFloatCoord(0-renderContext.MinX, renderContext.XRes)
							fYp = transformFloatCoord(0-renderContext.MinY, renderContext.YRes)
						} else {
							fXp = transformFloatCoord(stepData.PrevCoord.GetX()-renderContext.MinX, renderContext.XRes)
							fYp = transformFloatCoord(stepData.PrevCoord.GetY()-renderContext.MinY, renderContext.YRes)
						}

						fXc := transformFloatCoord(stepData.Coord.GetX()-renderContext.MinX, renderContext.XRes)
						fYc := transformFloatCoord(stepData.Coord.GetY()-renderContext.MinY, renderContext.YRes)
						fI := transformFloatCoord(stepData.Coord.GetI(), renderContext.XRes)
						fJ := transformFloatCoord(stepData.Coord.GetJ(), renderContext.YRes)

						// Arcs require floats!
						err := renderContext.DrawArc(fXp,
							fYp,
							fXc,
							fYc,
							fI,
							fJ,
							apertureSize,
							stepData.IpMode,
							stepData.QMode,
							// TODO
							renderContext.RegionColor)
						if err != nil {
							stepData.Print()
							checkError(err, 998)
						}
						renderContext.DrawDonut(Xp, Yp, apertureSize, 0, stepColor)
						renderContext.DrawDonut(Xc, Yc, apertureSize, 0, stepColor)
					} else if stepData.CurrentAp.Type == AptypeRectangle {
						fmt.Println("Arc drawing by rectangle aperture is not supported now.")
					} else {
						fmt.Println("Error. Only solid drawCircle and solid rectangle may be used to draw.")
						break
					}
				}
			}
			//
		case OpcodeD02_MOVE: // move
			renderContext.MovePen(Xp, Yp, Xc, Yc, renderContext.MovePenColor)
			//
		case OpcodeD03_FLASH: // flash
			if renderContext.DrawOnlyRegionsMode != true {
				renderContext.MovePen(Xp, Yp, Xc, Yc, renderContext.MovePenColor)
				if stepData.Polarity == PolTypeDark {
					stepColor = renderContext.ApColor
				} else {
					stepColor = renderContext.ClearColor
				}
				w := transformCoord(stepData.CurrentAp.XSize, renderContext.XRes)
				h := transformCoord(stepData.CurrentAp.YSize, renderContext.YRes)
				d := transformCoord(stepData.CurrentAp.Diameter, renderContext.XRes)
				hd := transformCoord(stepData.CurrentAp.HoleDiameter, renderContext.XRes)

				switch stepData.CurrentAp.Type {
				case AptypeRectangle:
					renderContext.DrawFilledRectangle(Xc, Yc, w, h, stepColor)
				case AptypeCircle:
					renderContext.DrawDonut(Xc, Yc, d, hd, stepColor)
				case AptypeObround:
					if w == h {
						renderContext.DrawDonut(Xc, Yc, w, hd, stepColor)
					} else {
						renderContext.DrawObRound(Xc, Yc, w, h, 0, renderContext.ObRoundColor)
					}
				case AptypePoly:
					renderContext.DrawDonut(Xc, Yc, d, hd, renderContext.MissedColor)
					fmt.Println("Polygonal apertures ain't supported.")
				default:
					checkError(errors.New("bad aperture type found"), 501)
					break
				}
			}
		default:
			checkError(errors.New("(renderContext *Render) ProcessStep(stepData *gerbparser.State) internal error. Bad opcode"), 666)
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

func checkError(err error, exitCode int) {
	if err != nil {
		fmt.Println(err)
		os.Exit(exitCode)
	}
}


/* ########################################## EOF #########################################################*/
