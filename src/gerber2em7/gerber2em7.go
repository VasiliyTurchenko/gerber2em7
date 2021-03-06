// Copyright 2018 Vasily Turchenko <turchenkov@gmail.com>. All rights reserved.
// Use of this source code is free

package gerber2em7

import (
	"configurator"
	"container/list"
	"errors"
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	stor "strings_storage"
	"time"
	"unicode"
)

import (
	. "gerberbasetypes"
	glog "glog_t"
	"plotter"
	"render"
	. "xy"
)

import "versiongenerator"

var (

	// configuration base
	viperConfig *viper.Viper

	// global storage of input gerber file strings, the source to feed some processors
	gerberStrings *stor.Storage

	// plotter instance which is responsible for generating the command stream for the target device
	plotterInstance *plotter.PlotterParams

	// array of steps to be executed to generate PCB
	arrayOfSteps []*render.State

	// the list of regions
	regionsList *list.List

	// the list of all the apertures
	aperturesList *list.List

	// the map consisting all the aperture blocks
	apertureBlocks map[string]*render.BlockAperture

	// format specification for the gerber file
	fSpec *FormatSpec

	//render context
	renderContext *render.Render
)

func init() {
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: example -stderrthreshold=[INFO|WARN|FATAL] -log_dir=[string]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func Main() {

	var (
		PlotterFilesFolder = ""
		//		IntermediateFilesFolder = ""
		PNGFilesFolder = ""
		inFileName     = ""
	)

	var sourceFileName string
	flag.StringVar(&sourceFileName, "i", "", "input file")

	flag.Set("stderrthreshold", "ERROR")
	flag.Set("alsologtostderr", "true")
	flag.Set("logtostderr", "true")

	flag.Parse()

	glog.Infoln(returnAppInfo(3))

	viperConfig = viper.New()
	configurator.SetDefaults(viperConfig)

	//	configurator.DiagnosticAllCfgPrint(viperConfig)

	cfgFileError := configurator.ProcessConfigFile(viperConfig)
	if cfgFileError != nil {
		fmt.Print("An error has occured: ")
		fmt.Println(cfgFileError)
		fmt.Println("Using built-in defaults.")
		configurator.SetDefaults(viperConfig)
	}

	//	configurator.DiagnosticAllCfgPrint(viperConfig)

	if len(sourceFileName) == 0 {
		fmt.Println("No input file specified.\nUsage:")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	_, inFileName = filepath.Split(sourceFileName)

	timeStamp := time.Now()

	glog.Infoln(timeInfo(timeStamp)+"input file:", sourceFileName)

	PlotterFilesFolder = filepath.FromSlash(viperConfig.Get(configurator.CfgFoldersPlotterFilesFolder).(string))
	//	IntermediateFilesFolder = filepath.FromSlash(viperConfig.Get(configurator.CfgFoldersIntermediateFilesFolder).(string))
	PNGFilesFolder = filepath.FromSlash(viperConfig.Get(configurator.CfgFoldersPNGFilesFolder).(string))

	//glog.Infoln("folders.PlotterFilesFolder = " + PlotterFilesFolder)
	//glog.Infoln("folders.IntermediateFilesFolder = " + IntermediateFilesFolder)
	//glog.Infoln("folders.PNGFilesFolder = " + PNGFilesFolder)

	/*
	   Process input string
	*/
	printMemUsage("Memory usage before reading input file:")

	gerberStrings = stor.NewStorage()

	fSpec = new(FormatSpec)

	content, err := ioutil.ReadFile(sourceFileName)
	if err != nil {
		checkError(err)
	}
	splittedString := TokenizeGerber(&content)
	// feed the storage
	for _, str := range *splittedString {
		gerberStrings.Accept(squeezeString(strings.ToUpper(str)))
	}
	// save splitted strings to a file
	saveIntermediate(gerberStrings, inFileName+"_pure_gerber.txt")

	// search for format definition strings
	mo, err := searchMO(gerberStrings)
	if err != nil {
		glog.Warning(err)
	}

	fs, err := searchFS(gerberStrings)
	checkError(err)

	fSpec = new(FormatSpec)
	if fSpec.Init(fs, mo) == false {
		glog.Fatalf("Can not parse: %s \\n %s\n", fs, mo)
	}
	printMemUsage("Memory usage before extracting apertures:")
	/* ---------------------- extract aperture macro defs to the am dictionary ----------- */
	render.AMacroDict, gerberStrings = render.ExtractAMDefinitions(gerberStrings)

	if viperConfig.GetBool(configurator.CfgCommonPrintAperturesInfo) == true {
		for i := range render.AMacroDict {
			glog.Info(render.AMacroDict[i].String())
		}
	}

	/* ---------------------- extract apertures and aperture blocks  --------------------- */
	gerberStrings2 := stor.NewStorage()
	aperturesList = list.New()
	apertureBlocks = make(map[string]*render.BlockAperture)
	apertureBlockOpened := make([]string, 0)
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
			lastOpenedAB := len(apertureBlockOpened) - 1
			if lastOpenedAB < 0 {
				//				panic("No more open aperture blocks left!")
				glog.Fatalln("No more open aperture blocks left!")
			}
			aperture := new(render.Aperture)
			aperture.Code = apertureBlocks[apertureBlockOpened[lastOpenedAB]].Code
			aperture.Type = AptypeBlock
			aperture.BlockPtr = apertureBlocks[apertureBlockOpened[lastOpenedAB]]
			aperture.BlockPtr.StepsPtr = make([]*render.State, len(aperture.BlockPtr.BodyStrings)+1)
			aperture.BlockPtr.StepsPtr[0] = render.NewState()
			apertureBlockOpened = apertureBlockOpened[:lastOpenedAB]
			aperturesList.PushBack(aperture) // store correct aperture
			continue
		}
		// new block is met
		if strings.HasPrefix(gerberString, GerberApertureBlockDef) &&
			strings.HasSuffix(gerberString, "*%") {
			// aperture block found
			apBlk := new(render.BlockAperture)
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
			aperturesList.PushBack(render.NewApertureInstance(gerberString, fSpec.ReadMU()))
			continue
		}
		// all unprocessed above goes here
		gerberStrings2.Accept(gerberString)
	}

	// Global array of commands
	gerberStrings, gerberStrings2 = gerberStrings2, nil

	saveIntermediate(gerberStrings, inFileName+"_before_steps.txt")

	// Main sequence of steps
	arrayOfSteps = make([]*render.State, gerberStrings.Len()+1)
	// Global list of Regions
	regionsList = list.New()

	//  Aperture blocks must be converted to the steps w/o AB
	//  S&R blocks and regions inside each instance of AB added to the global lists!
	for apBlock := range apertureBlocks {
		bsn := render.CreateStepSequence(&apertureBlocks[apBlock].BodyStrings,
			&apertureBlocks[apBlock].StepsPtr,
			aperturesList,
			regionsList,
			fSpec)
		apertureBlocks[apBlock].StepsPtr = apertureBlocks[apBlock].StepsPtr[:bsn]
	}

	printMemUsage("Memory usage before creating Main step sequence:")

	// patch
	// TODO get rid of the patch!
	gerberStringsArray := gerberStrings.ToArray()

	numberOfSteps := render.CreateStepSequence(&gerberStringsArray,
		&arrayOfSteps,
		aperturesList,
		regionsList,
		fSpec)
	arrayOfSteps = arrayOfSteps[1:numberOfSteps]

	/* ------------------ aperture blocks to steps ---------------------------*/
	// each D03 must be checked against aperture block

	printMemUsage("Memory usage before unwinding aperture blocks:")

	//	var touch bool = false
	for {
		touch := false
		arrayOfSteps2 := make([]*render.State, 0)
		for k := 0; k < len(arrayOfSteps); k++ {
			if arrayOfSteps[k].CurrentAp != nil &&
				arrayOfSteps[k].CurrentAp.Type == AptypeBlock &&
				arrayOfSteps[k].Action == OpcodeD03_FLASH {
				for i, bs := range arrayOfSteps[k].CurrentAp.BlockPtr.StepsPtr {
					if i == 0 { // skip root element
						continue
					}
					newStep := render.NewState()
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
			insert, tailI := render.UnwindSRBlock(&arrayOfSteps, i)
			lenTail := len(arrayOfSteps) - tailI
			tail := make([]*render.State, lenTail)
			for j := 0; j < lenTail; j++ {
				//				tail[j] = new(render.State)
				tail[j] = render.NewState()
				tail[j].CopyOfWithOffset(arrayOfSteps[tailI+j], 0, 0)
			}
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
			glog.Infoln("\n" + k.Value.(*render.Aperture).String())
			j++
		}
		glog.Infoln("Total", j, "regions found.")
	}
	// print apertures info
	if viperConfig.GetBool(configurator.CfgCommonPrintAperturesInfo) == true {
		j := 0
		for k := aperturesList.Front(); k != nil; k = k.Next() {
			glog.Infoln("\n" + k.Value.(*render.Aperture).String())
			j++
		}
		glog.Infoln("Total", j, "apertures found.")
	}

	glog.Infoln("Total", len(arrayOfSteps)-1, "steps to do.")

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

	glog.Info(timeInfo(timeStamp) + "Rendering process started\n")

	/*
	   let's render the PCB
	*/
	plotterInstance = plotter.NewPlotter()
	plotterInstance.TakePen(1)

	ofNameFromCfg := viperConfig.GetString(configurator.CfgPlotterOutFile)
	if len(ofNameFromCfg) == 0 {
		ofNameFromCfg = inFileName + ".plt"
	}

	outfname := filepath.Join(filepath.ToSlash(PlotterFilesFolder), ofNameFromCfg)
	plotterInstance.SetOutFileName(outfname)
	renderContext = render.NewRender(plotterInstance, viperConfig, minX, minY, maxX, maxY)
	glog.Infof("Min. X, Y found: (%f,%f)\n", minX, minY)
	glog.Infof("Max. X, Y found: (%f,%f)\n", maxX, maxY)

	printMemUsage("Memory usage after render context was initialized:")

	// draw frame by dashed line
	renderContext.DrawFrame()

	k := 0
	for k < len(arrayOfSteps) {
		if arrayOfSteps[k].Action == OpcodeStop {
			break
		}
		//		ProcessStep(arrayOfSteps[k])
		arrayOfSteps[k].Render(renderContext)
		k++
	}

	if viperConfig.GetBool(configurator.CfgCommonPrintStatistic) == true {
		glog.Infof("%s%d%s", "The plotter have drawn ", renderContext.LineBresCounter, " straight lines using Brezenham\n")
		glog.Infof("%s%.0f%s", "Total lenght of straight lines = ", renderContext.LineBresLen*renderContext.XRes, " mm\n")
		glog.Infof("%s%d%s", "The plotter have drawn ", renderContext.CircleBresCounter, " circles\n")
		glog.Infof("%s%.0f%s", "Total lenght of circles = ", renderContext.CircleLen*renderContext.XRes, " mm\n")
		glog.Infoln("The plotter have drawn", renderContext.FilledRctCounter, "filled rectangles")
		glog.Infoln("The plotter have drawn", renderContext.ObRoundCounter, "obrounds (boxes)")
		glog.Infoln("The plotter have moved pen", renderContext.MovePenCounters, "times")
		glog.Infof("%s%.0f%s", "Total move distance = ", renderContext.MovePenDistance*renderContext.XRes, " mm\n")
	}

	if renderContext.YNeedsFlip == true {
		glog.Infoln(timeInfo(timeStamp) + "Started flipping (only png image) over X-axis")
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

	glog.Infoln(timeInfo(timeStamp) + "Rendering process finished")

	// Save to out.png
	if viperConfig.GetBool(configurator.CfgRendererGeneratePNG) == true {
		printMemUsage("Memory usage before png encoding:")

		glog.Infoln(timeInfo(timeStamp)+"Generating png image ", renderContext.Img.Bounds().String())
		/*
			ofNameFromCfg := viperConfig.GetString(configurator.CfgPlotterOutFile)
			if len(ofNameFromCfg) == 0 {
				ofNameFromCfg = inFileName + ".plt"
			}

		*/
		pngNameFromCfg := viperConfig.GetString(configurator.CfgRendererOutFile)
		if len(pngNameFromCfg) == 0 {
			pngNameFromCfg = inFileName + ".png"
		}
		ofname := filepath.Join(filepath.ToSlash(PNGFilesFolder), pngNameFromCfg)
		f, _ := os.OpenFile(ofname, os.O_WRONLY|os.O_CREATE, 0600)
		defer f.Close()
		png.Encode(f, renderContext.Img)

		glog.Infoln(timeInfo(timeStamp)+"Image is saved to the file", ofname)
		printMemUsage("Memory usage after png encoding:")
	}

	glog.Infoln(timeInfo(timeStamp) + "Saving plotter commands stream to file")
	plotterInstance.Stop()
	glog.Infoln(timeInfo(timeStamp)+"Plotter commands are saved to the file", outfname)
	glog.Exitln(timeInfo(timeStamp) + "Exiting")
}

////////////////////////////////////////////////////// end of main ///////////////////////////////////////////////////

// search for format strings
func searchMO(storage *stor.Storage) (string, error) {
	err := errors.New("unit of measurements command not found - MOIN used by default")
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
	return GerberMOIN, err
}

func searchFS(storage *stor.Storage) (string, error) {

	storage.ResetPos()
	//	for _, s := range gerberStrings {

	s := storage.String()
	for len(s) > 0 {
		if strings.HasPrefix(s, "%FST") {
			return s, errors.New("trailing zero omission format is not supported") // + 09-Jun-2018
		}
		if strings.HasPrefix(s, "%FSLI") || strings.HasPrefix(s, "%FSTI") {
			return s, errors.New("incremental coordinates ain't supported") // + 09-Jun-2018
		}

		if strings.HasPrefix(s, GerberFormatSpec) {
			return s, nil
		}
		s = storage.String()
	}
	return "", errors.New("%FS command not found")
}

/*
	Saves intermediate results from the strings storage to the file
*/
func saveIntermediate(storage *stor.Storage, fileName string) {

	if viperConfig.GetBool(configurator.CfgParserSaveIntermediate) == false {
		return
	}

	fileName = filepath.Join(viperConfig.Get(configurator.CfgFoldersIntermediateFilesFolder).(string), fileName)

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		//		panic(err)
		glog.Fatalln(err)
	}
	defer file.Close()
	err = file.Truncate(0)
	if err != nil {
		//		panic(err)
		glog.Fatalln(err)
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
			glog.Fatalln(err)
			//			panic(err)
		}
	}
	file.Sync()
	err = file.Close()
	if err != nil {
		//		panic(err)
		glog.Fatalln(err)
	}
	glog.Infoln("Intermediate file " + fileName + " is saved.")
}

func printSqueezedOut(str string) {
	if viperConfig.GetBool(configurator.CfgCommonPrintGerberComments) == true {
		glog.Infoln(str)
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
	var header = "Gerber to EM-7052 translation tool. "
	var version = "Version 0.2.0. "
	var retVal = "\n"
	switch verbLevel {
	case 3:
		retVal = header + version + versiongenerator.BuildDateTime
	case 2:
		retVal = header + version
	case 1:
		retVal = header
	default:
		retVal = "\n"
	}
	return retVal
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
	glog.Infof(header+" Alloc = %v KB\tTotalAlloc = %v KB\tSys = %v KB\tNumGC = %v\n",
		bToKb(memStats.Alloc),
		bToKb(memStats.TotalAlloc),
		bToKb(memStats.Sys),
		memStats.NumGC)
}

func bToKb(b uint64) uint64 {
	return b / 1024
}

func timeInfo(prev time.Time) string {
	//	now := time.Now()
	elapsed := time.Since(prev)
	/*
		"[23:59:04 +2.00001] "
	*/
	out := "[+"
	//hr := strconv.Itoa(now.Hour())
	//if len(hr) == 1 {
	//	hr = "0" + hr
	//}
	//min := strconv.Itoa(now.Minute())
	//if len(min) == 1 {
	//	min = "0" + min
	//}
	//sec := strconv.Itoa(now.Second())
	//if len(sec) == 1 {
	//	sec = "0" + sec
	//}
	//
	//out = out + hr + ":" + min + ":" + sec + " +"
	elapsedSec := (float64(elapsed.Nanoseconds() / (1000 * 1000))) / 1000.0
	//out = out + strconv.FormatFloat(elapsedSec, 'f', 3, 64) + "] "
	return out + strconv.FormatFloat(elapsedSec, 'f', 3, 64) + "] "
	//fmt.Print(out)
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
	//	panic("math/int.Abs: invalid argument")
	glog.Fatalln("math/int.Abs: invalid argument")
	return 0
}

func checkError(err error) {
	if err != nil {
		glog.Fatalln(err)
	}
}

/* ----- gerber string tokenizer ------------------------------------ */

func TokenizeGerber(buf *[]byte) *[]string {
	retVal := make([]string, 0)
	/*
		1. if we met '%', all the bytes until next '%' stay unchanged.
		Leading and trailing '%' are included in the out string
		2. trim all spaces between trailing '%' and '*' and next non-spase symbol
		3. each stream of bytes with trailing '*' is treated as separate string
	*/
	if len(*buf) < 2 {
		return &[]string{string(*buf)}
	}
	a := 0
	b := len(*buf)
	for a < b {
		if (*buf)[a] == '%' {
			trailerFound := false
			// scan for trailing '%'
			start := a
			a++
			for a < b {
				if (*buf)[a] != '%' {
					a++
				} else {
					trailerFound = true
					break
				}
			}
			if trailerFound == true {
				a++
			}
			filtered := FilterNewLines(string((*buf)[start:a]))
			retVal = append(retVal, filtered)
			continue
		}
		if unicode.IsSpace(rune((*buf)[a])) == true {
			a++
			continue
		}

		// fix strange files with \0 \0 \0...
		if (*buf)[a] == 0x00 {
			a++
			continue
		}

		if (*buf)[a] != '*' {
			trailerFound := false
			// scan for trailing '*'
			start := a
			a++
			for a < b {
				if (*buf)[a] != '*' {
					a++
				} else {
					trailerFound = true
					break
				}
			}
			if trailerFound == true {
				a++
			}
			filtered := FilterNewLines(string((*buf)[start:a]))

			// fix G75G03X0Y0D03* case
			if strings.HasPrefix(filtered, "G04") == false &&
				strings.HasPrefix(filtered, "G4") == false {
				for {
					if len(filtered) > 4 && filtered[0] == 'G' && filtered[3] != '*' {
						retVal = append(retVal, filtered[:3]+"*")
						filtered = filtered[3:]
						continue
					}
					//if len(filtered) > 3 && filtered[0] == 'G' && filtered[2] != '*' {
					//		retVal = append(retVal, filtered[:2]+"*")
					//		filtered = filtered[2:]
					//		continue
					//	}
					retVal = append(retVal, filtered)
					break
				}
			} else {
				retVal = append(retVal, filtered)
			}
			continue
		} else {
			a++
			continue
		}
	}
	return &retVal
}

//filters \n \r symbols from the string
func FilterNewLines(inString string) string {
	retVal := strings.Replace(inString, "\n", "", -1)
	return strings.Replace(retVal, "\r", "", -1)
}

/* ########################################## EOF #########################################################*/
