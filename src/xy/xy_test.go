package xy

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

var maxCoord float64
var maxFracLen int
var maxIntLen int

const truncPos int = 8

var fstr = []string{
	"%FSLAX14Y14*%",
	"%FSLAX15Y15*%",
	"%FSLAX16Y16*%",
	"%FSLAX17Y17*%",

	"%FSLAX24Y24*%",
	"%FSLAX25Y25*%",
	"%FSLAX26Y26*%",
	"%FSLAX27Y27*%",

	"%FSLAX34Y34*%",
	"%FSLAX35Y35*%",
	"%FSLAX36Y36*%",
	"%FSLAX37Y37*%",

	"%FSLAX44Y44*%",
	"%FSLAX45Y45*%",
	"%FSLAX46Y46*%",
	"%FSLAX47Y47*%",

	"%FSLAX54Y54*%",
	"%FSLAX57Y57*%",

	"%FSLAX64Y64*%",
	"%FSLAX65Y65*%",
	"%FSLAX66Y66*%",
	"%FSLAX67Y67*%",
}
var mostr string = "%MOMM*%"

var formspec *FormatSpec

type testDataElementStruct struct {
	present, er, neg bool
	val              float64
	comp             string
}

type testDataStruct struct {
	Elements [5]testDataElementStruct
	result   bool // expected result of the test function
}

var testData testDataStruct

/*
func TestFormatSpec_Init(t *testing.T) {

	for i := range fstr {
		formspec := new(FormatSpec)
		if formspec.Init(fstr[i], mostr) == false {
			t.Error("formspec.Init test FAILED!")
		} else {
			fmt.Println("formspec.Init test PASSED!")
		}
	}
}
*/
// generates random bool value
func randBool() bool {
	r := false
	if (rand.Uint32() % 2) == 1 {
		r = true
	}
	return r
}

// generates random extended ASCII code
func randLetter(pc string) byte {
L1:
	c := (byte)(rand.Uint32() & 0xFF)
	if (c < 0x20) || strings.ContainsAny(string(c), pc) {
		goto L1
	}
	return c
}

// function returns a component of the resulting test string
func randComp(i int) string {

	var s string = ""
	var errInNumber bool
	var pc string = "xyijdXYIJD"

	errInNumber = randBool()

	if !testData.Elements[i].present {
		return ""
	}
	if testData.Elements[i].er && !errInNumber {
		s = s + string(randLetter(pc))
	} else {
		s = s + string(pc[i])
	}
	if testData.Elements[i].neg {
		s = s + "-"
	} else {
		s = s + "+"
	}
	//	s = s + "000"
	//	var ipart int
	// ipart = int(testData.Elements[i].val)
	//	s = s + strconv.Itoa(int(testData.Elements[i].val)) //strconv.Itoa(ipart)

	//	pow := maxFracLen //formspec.XD

	//	fpart := int((testData.Elements[i].val - float32(ipart)) * float32(math.Pow10(pow)))
	//	tmps :=  strconv.Itoa(fpart)

	//	tmps := strconv.FormatFloat(float64(testData.Elements[i].val)-math.Floor(float64(testData.Elements[i].val)), 'f', truncPos, 32)

	tmps := strconv.FormatFloat(float64(testData.Elements[i].val), 'f', maxFracLen, 32)

	j := strings.IndexByte(tmps, '.')

	s = s + tmps[:j] + tmps[j+1:]

	//	s = s + (strings.TrimPrefix(tmps, "0."))[:maxFracLen]

	/*
		if len(tmps) < maxFracLen {
			for dl:=0; dl < (maxFracLen - len(tmps)); dl++ {
				tmps = tmps + "0"
			}
		}
		s = s + tmps
	*/
	if testData.Elements[i].neg {
		testData.Elements[i].val *= -1
	}

	if testData.Elements[i].er && errInNumber {
		sp := rand.Intn(len(s)-2) + 1
		out := []byte(s)
		out[sp] = randLetter("0123456789xyijXYIJ+-")
		s = string(out)
	}

	//	fmt.Println(er, "  ", string(pc1), "  ", string(pc2), "  ", s )
	return s
}

// generates a suitable coordinates string to be parsed by function XY.init
func generateXYString() (s string, r bool) {

	for i := range testData.Elements {
		testData.Elements[i].present = randBool()
		testData.Elements[i].neg = randBool()
		testData.Elements[i].er = randBool()
		testData.Elements[i].val = rand.Float64() * maxCoord
		testData.Elements[i].comp = randComp(i)
	}

	if testData.Elements[4].present {
		if testData.Elements[4].er {
			testData.Elements[4].comp = string(randLetter("dD"))
		} else {
			testData.Elements[4].comp = "d"
		}
	} else {
		testData.Elements[4].comp = ""
	}

	r = true

	for i := range testData.Elements {
		s = s + testData.Elements[i].comp
		if testData.Elements[i].present && testData.Elements[i].er {
			r = false
		}
	}
	if (r == false) || (testData.Elements[4].present == false) {
		r = false
	}

	if !testData.Elements[0].present && !testData.Elements[1].present && !testData.Elements[2].present && !testData.Elements[3].present {
		r = false
	}
	testData.result = r
	return s, r
}

/*
func TestXY_Init(t *testing.T) {

	var txy *XY
	txy = new(XY)

	for i := range fstr {
		rand.Seed(int64(rand.Uint64()))

		formspec := new(FormatSpec)
		_ = formspec.Init(fstr[i], mostr)
		maxFracLen = formspec.ReadXD()
		maxIntLen = formspec.ReadXI()
		maxCoord = math.Pow10(maxIntLen)

		fmt.Println("Coord. format:", maxIntLen, ".", maxFracLen)
		fmt.Println("Max.coord.:", maxCoord-1)

		var trueRuns int = 0
		var falseRuns int = 0
		var rr bool

		for i := 0; i < 1000000; i++ {
			cstr, er := generateXYString()

			if testData.result == false {
				falseRuns++
			} else {
				trueRuns++
			}

			if rr = txy.Init(cstr, formspec, nil); rr != er {
				t.Error("xy.Init test FAILED! Iteration:", i, " Test string =", cstr, ".  Expected parsing result =", er, ";  got =", rr)
				t.Error(testData)
				break
			} else {
				//t.Error("xy.Init test PASSED! Iteration:", i, " Test string =", cstr, ".  Expected parsing result =", er, ";  got =", rr)
			}

			if (rr == true) && (er == true) {
				if testData.Elements[0].present {
					if math.Abs(float64(testData.Elements[0].val-txy.GetX())) > math.Pow10(-1*maxFracLen) {
						fmt.Println("cstr = ", cstr)
						fmt.Println("given X = ", testData.Elements[0].val, "\tgot X=", txy.GetX())
						fmt.Println("diff =", testData.Elements[0].val-txy.GetX())
						break
					}
				}

				if testData.Elements[1].present {
					if math.Abs(float64(testData.Elements[1].val-txy.GetY())) > math.Pow10(-1*maxFracLen) {
						fmt.Println("cstr = ", cstr)
						fmt.Println("given Y = ", testData.Elements[1].val, "\tgot Y=", txy.GetY())
						fmt.Println("diff =", (float64(testData.Elements[1].val-txy.GetY()) * math.Pow10(maxFracLen)))
						break
					}
				}

				if testData.Elements[2].present {
					if math.Abs(float64(testData.Elements[2].val-txy.GetI())) > math.Pow10(-1*maxFracLen) {
						fmt.Println("cstr = ", cstr)
						fmt.Println("given I = ", testData.Elements[2].val, "\tgot I=", txy.GetI())
						fmt.Println("diff =", (float64(testData.Elements[2].val-txy.GetX()) * math.Pow10(maxFracLen)))
						break
					}
				}

				if testData.Elements[3].present {
					if math.Abs(float64(testData.Elements[3].val-txy.GetJ())) > math.Pow10(-1*maxFracLen) {
						fmt.Println("cstr = ", cstr)
						fmt.Println("given J = ", testData.Elements[3].val, "\tgot J=", txy.GetJ())
						fmt.Println("diff =", (float64(testData.Elements[3].val-txy.GetJ()) * math.Pow10(maxFracLen)))
						break
					}
				}
			}

			//txy.Print()
		}
		fmt.Println("True runs:", trueRuns, "\tfalse runs:", falseRuns)
	}
}

*/
func Test_ExtractOrderedVals(t *testing.T) {
	s := "B678.68C7897A8098"
	ts := "BCDA"
	r, err := ExtractLetterDelimitedFloats(s, ts)
	if err == nil {
		fmt.Println(r)
	} else {
		fmt.Println(err)
	}

}
