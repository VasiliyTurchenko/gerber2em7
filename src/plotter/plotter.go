/*
 Generates a stream of EM-7052 commands
*/
package plotter

import (
	. "gerberbasetypes"
	glog "glog_t"
	"os"
	"strconv"
	"strings"
)

/*
	Plotter current status and statistic
*/
type PlotterParams struct {
	selectPenCmds   int
	dropPenCmds     int
	raisePenCmds    int
	moveCmds        int
	currentPosX     int
	currentPosY     int
	outFileName     string
	outputFile      *os.File
	err             error
	outStringBuffer []string
}

func NewPlotter() *PlotterParams {
	retVal := new(PlotterParams)
	retVal.Init()
	return retVal
}

type Plotter interface {

	// Initializes plotter by setting up the commands
	InitCommands() error

	//
	Start() string

	// Stops the plotter
	Stop() string

	// moves the tool to position
	MoveTo(x, y int) string

	// draws a line
	DrawLine(x0, y0, x1, y1 int) string

	// draws a circle
	Circle(xc, yc, r int) string

	// draws an arc
	Arc(x0, y0, x1, y1, radius, fi0, fi1 int, ipm IPmode) string

	// takes a pen
	TakePen(penNumber int) string
}

/*
	Initializes Plotter object and generates plotter reset command
*/
func (plotter *PlotterParams) Init() string {
	plotter.currentPosX = 0
	plotter.currentPosY = 0
	plotter.outStringBuffer = make([]string, 0)
	retVal := "J\n"
	plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	return retVal
}

func (plotter *PlotterParams) SetOutFileName(outFileName string) {
	plotter.outFileName = outFileName
}

/*
	Deletes unnecessary MA commands
*/
func (plotter *PlotterParams) squeeze() {
	tmpString := make([]string, 0)
	var lastMA string
	for a := range plotter.outStringBuffer {
		if strings.HasPrefix(plotter.outStringBuffer[a], "MA ") {
			lastMA = plotter.outStringBuffer[a]
		} else {
			if len(lastMA) > 0 {
				tmpString = append(tmpString, lastMA)
			}
			tmpString = append(tmpString, plotter.outStringBuffer[a])
			lastMA = ""
		}
	}
	plotter.outStringBuffer = tmpString
}

/*
	Finalizes command stream and writes file to disk
*/
func (plotter *PlotterParams) Stop() {
	_ = plotter.TakePen(0)
	_ = plotter.MoveTo(0, 0)
	plotter.squeeze()
	plotter.outputFile, plotter.err = os.OpenFile(plotter.outFileName, os.O_WRONLY|os.O_CREATE, 0600)
	if plotter.err != nil {
		glog.Fatal(plotter.err)
	}
	defer plotter.outputFile.Close()
	plotter.err = plotter.outputFile.Truncate(0)
	if plotter.err != nil {
		glog.Fatal(plotter.err)
	}
	for _, s := range plotter.outStringBuffer {
		_, plotter.err = plotter.outputFile.WriteString(s)
		if plotter.err != nil {
			glog.Fatal(plotter.err)
		}
	}
	plotter.outputFile.Sync()
	plotter.err = plotter.outputFile.Close()
	if plotter.err != nil {
		glog.Fatal(plotter.err)
	}
	plotter.outStringBuffer = nil
	return
}

func (plotter *PlotterParams) MoveTo(x, y int) string {
	retVal := plotter.moveTo(x, y)
	plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	return retVal
}

func (plotter *PlotterParams) moveTo(x, y int) string {
	var retVal string
	plotter.currentPosX = x
	plotter.currentPosY = y
	retVal = "MA " + strconv.Itoa(x) + " , " + strconv.Itoa(y) + "\n"
	return retVal
}

func (plotter *PlotterParams) DrawLine(x0, y0, x1, y1 int) string {
	var retVal string
	if (plotter.currentPosX != x0) || (plotter.currentPosY != y0) {
		retVal = plotter.moveTo(x0, y0)
		plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	}
	retVal = "DA " + strconv.Itoa(x1) + " , " + strconv.Itoa(y1) + "\n"
	plotter.currentPosX = x1
	plotter.currentPosY = y1
	plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	return retVal
}

func (plotter *PlotterParams) Circle(xc, yc, r int) string {
	retVal := plotter.moveTo(xc+r, yc) // move to the rightmost circle point
	plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	retVal = "D C" + strconv.Itoa(r) + " , 0 , 360\n"
	plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	retVal = plotter.moveTo(xc, yc)
	plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	return retVal
}

func (plotter *PlotterParams) Arc(x0, y0, x1, y1, radius, fi0, fi1 int, ipm IPmode) string {
	var retVal string
	if (plotter.currentPosX != x0) || (plotter.currentPosY != y0) {
		glog.Error("Arc position discrepance: (currX, currY) (x0, y0) (" +
			strconv.Itoa(plotter.currentPosX) + "," + strconv.Itoa(plotter.currentPosY) + ") (" +
			strconv.Itoa(x0) + "," + strconv.Itoa(y0) + ")")
		retVal = plotter.moveTo(x0, y0)
		plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	}
	if ipm == IPModeCwC {
		radius = -radius
	} else {
		radius = radius
	}
	retVal = "DC " + strconv.Itoa(radius) + " , " + strconv.Itoa(fi0) + " , " + strconv.Itoa(fi1) + "\n"
	plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	retVal = plotter.moveTo(x1, y1)
	plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	return retVal

}

func (plotter *PlotterParams) TakePen(penNumber int) string {
	if penNumber < 0 || penNumber > 4 {
		glog.Fatal("Bad pen number specified!")
	}
	retVal := "P" + strconv.Itoa(penNumber) + "\n"
	plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	return retVal
}
