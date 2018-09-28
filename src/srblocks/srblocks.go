/*
The file contains functions and structures used for parsing gerber x2 file
*/
package srblocks

import (
	"errors"
	"hash/fnv"
	"strconv"
	"strings"
	. "xy"
)

/*
############################## step and repeat blocks #################################
*/
type SRBlock struct {
	srString string
	srId	 string
	numX     int
	numY     int
	dX       float64
	dY       float64
	nSteps   int // number of steps in the SRBlock block
}

func (srblock *SRBlock) String() string {

	if srblock == nil {
		return "<nil>"
	}
	return "Step and repeat block:\n" +
		"\tsource string: " + srblock.srString + "\n" +
		"\thash Id: " + srblock.srId + "\n" +
		"\tcontains " + strconv.Itoa(srblock.numX) + " repetition(s) along X axis and " + strconv.Itoa(srblock.numY) + " repetition(s) along Y axis\n" +
		"\tnumber of steps in each repetition: " + strconv.Itoa(srblock.nSteps) + "\n" +
		"\tdX=" + strconv.FormatFloat(srblock.dX, 'f', 5, 64) +
		", dy=" + strconv.FormatFloat(srblock.dY, 'f', 5, 64) + "\n"
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

func (srblock *SRBlock) IncNSteps() {
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

	hs := ins + strconv.Itoa(srblock.numX) + strconv.Itoa(srblock.numY) + strconv.FormatFloat(srblock.dX,'f', 5, 64) + strconv.FormatFloat(srblock.dY,'f', 5, 64)
	h := fnv.New32a()
	h.Write([]byte(hs))
	srblock.srId = strconv.FormatInt(int64(h.Sum32()), 10)
	return nil
}

//func (srblock *SRBlock) StartXY() *XY {
//	return srblock.startXY
//}
//
//func (srblock *SRBlock) SetStartXY(v *XY) {
//	srblock.startXY = v
//}

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
