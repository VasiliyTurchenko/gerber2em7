/*
 Generates a stream of EM-7052 commands
*/
package plotter

import (
	"fmt"
	"gerbparser"
	"os"
	"strconv"
	"strings"
)

/*
	Plotter current status and statistic
*/
type Plotter struct {
	selectPenCmds   int
	dropPenCmds     int
	raisePenCmds    int
	moveCmds        int
	currentPosX     int
	currentPosY     int
	outputFile      *os.File
	err             error
	outStringBuffer []string
}

func NewPlotter() *Plotter {
	retval := new(Plotter)
	retval.Init()
	return retval
}

/*
	Initializes Plotter object and genetrates plotter reset command
*/
func (plotter *Plotter) Init() string {
	plotter.currentPosX = 0
	plotter.currentPosY = 0
	plotter.outStringBuffer = make([]string, 0)
	retVal := "J\n"
	plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	return retVal
}

/*
	Deletes unnecessary MA commands
 */
func (plotter *Plotter) squeeze() {
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
func (plotter *Plotter) Stop() {
	_ = plotter.Pen(0)
	_ = plotter.MoveTo(0, 0)
	plotter.squeeze()
	plotter.outputFile, plotter.err = os.OpenFile("G:\\go_prj\\gerber2em7\\src\\out.plotter", os.O_WRONLY|os.O_CREATE, 0600)
	if plotter.err != nil {
		panic(plotter.err)
	}
	defer plotter.outputFile.Close()
	plotter.err = plotter.outputFile.Truncate(0)
	if plotter.err != nil {
		panic(plotter.err)
	}
	for _, s := range plotter.outStringBuffer {
		_, plotter.err = plotter.outputFile.WriteString(s)
		if plotter.err != nil {
			panic(plotter.err)
		}
	}
	plotter.outputFile.Sync()
	plotter.err = plotter.outputFile.Close()
	if plotter.err != nil {
		panic(plotter.err)
	}
	plotter.outStringBuffer = nil
	return
}

func (plotter *Plotter) MoveTo(x, y int) string {
	retVal := plotter.moveTo(x, y)
	plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	return retVal
}

func (plotter *Plotter) moveTo(x, y int) string {
	var retVal string
	plotter.currentPosX = x
	plotter.currentPosY = y
	retVal = "MA " + strconv.Itoa(x) + " , " + strconv.Itoa(y) + "\n"
	return retVal
}

func (plotter *Plotter) DrawLine(x0, y0, x1, y1 int) string {
	var retVal string
	if (plotter.currentPosX != x0) || (plotter.currentPosY != y0) {
		//		fmt.Println("Draw line. Position discrepance detected!")
		retVal = plotter.moveTo(x0, y0)
		plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	}
	retVal = "DA " + strconv.Itoa(x1) + " , " + strconv.Itoa(y1) + "\n"
	plotter.currentPosX = x1
	plotter.currentPosY = y1
	plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	return retVal
}

func (plotter *Plotter) Circle(xc, yc, r int) string {
	retval := plotter.moveTo(xc+r, yc) // move to the rightmost circle point
	plotter.outStringBuffer = append(plotter.outStringBuffer, retval)
	retval = "D C" + strconv.Itoa(r) + " , 0 , 360\n"
	plotter.outStringBuffer = append(plotter.outStringBuffer, retval)
	retval = plotter.moveTo(xc, yc)
	plotter.outStringBuffer = append(plotter.outStringBuffer, retval)
	return retval
}

func (plotter *Plotter) Arc(x0, y0, x1, y1, radius, fi0, fi1 int, ipm gerbparser.IPmode) string {
	var retval string
	if (plotter.currentPosX != x0) || (plotter.currentPosY != y0) {
		fmt.Println("Arc. Position discrepance detected!")
		retval = plotter.moveTo(x0, y0)
		plotter.outStringBuffer = append(plotter.outStringBuffer, retval)
	}
	if ipm == gerbparser.IPModeCwC {
		radius = -radius
	} else {
		radius = radius
	}
	retval = "DC " + strconv.Itoa(radius) + " , " + strconv.Itoa(fi0) + " , " + strconv.Itoa(fi1) + "\n"
	plotter.outStringBuffer = append(plotter.outStringBuffer, retval)
	retval = plotter.moveTo(x1, y1)
	plotter.outStringBuffer = append(plotter.outStringBuffer, retval)
	return retval

}

func (plotter *Plotter) Pen(penNumber int) string {
	if penNumber < 0 || penNumber > 4 {
		panic("Bad pen number specified!")
	}
	retVal := "P" + strconv.Itoa(penNumber) + "\n"
	plotter.outStringBuffer = append(plotter.outStringBuffer, retVal)
	return retVal
}
