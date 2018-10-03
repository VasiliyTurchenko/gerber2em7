package geberlexer

import (
	"strconv"
	"testing"
)

func TestGerberCommand_String(t *testing.T) {
	cmd := []GerberCommand{
		{FS, "the body of FS Command"},
		{MO, "the body of MO Command"},
		{AM, "the body of AM Command"},
		{G04, "COMMENT!!!"},
		{M02, "STOP"},
	}
	for i := range cmd {
		t.Logf("%s\n",cmd[i].String())
	}
}

func TestGerberCommandId_String(t *testing.T) {
	for i := range GCmdBaseArray {
		t.Logf("%s\n",GCmdBaseArray[i].cmd.String())
	}
	for i := range GCmdExtArray {
		t.Logf("%s\n",GCmdExtArray[i].cmd.String())
	}

}

type testdata struct {
	input string
	answer string
}

var td = []testdata{
	{"X4794700Y2202900D01*\n", "{command:\"D01\",val:\"X4794700Y2202900\"}"},
	{"X4794700Y2202900D02*", "{command:\"D02\",val:\"X4794700Y2202900\"}"},
	{"X4794700Y2202900D03*\n", "{command:\"D03\",val:\"X4794700Y2202900\"}"},
	{"X4794700Y2202900D100*\n\n\n", "{command:\"D\",val:\"100\"}"},
	{"D01*", "{command:\"D01\",val:\"\"}"},
	{"D02*\n", "{command:\"D02\",val:\"\"}"},
	{"D03*", "{command:\"D03\",val:\"\"}"},
	{"D04*", "{command:\"D\",val:\"04\"}"},
	{"D000999*\n", "{command:\"D\",val:\"999\"}"},
	{"G01*", "{command:\"G01\",val:\"\"}"},
	{"G1*", "{command:\"G01\",val:\"\"}"},
	{"G02*", "{command:\"G02\",val:\"\"}"},
	{"G03*", "{command:\"G03\",val:\"\"}"},
	{"G04 ZZZ*", "{command:\"G04\",val:\"ZZZ\"}"},
	{"G05*", "{command:\"NOP\",val:\"G05\"}"},
	{"G5*", "{command:\"NOP\",val:\"G05\"}"},
	{"X4794700Y2202900D03*", "{command:\"D03\",val:\"X4794700Y2202900\"}"},
	{"X4794700Y2202900D03*", "{command:\"D03\",val:\"X4794700Y2202900\"}"},
}

func TestSplitByGCommands(t *testing.T) {
	//var sf = "G:\\go_prj\\gerber2em7\\gerbers\\am-test.gbx"
	//sf = "G:\\go_prj\\gerber2em7\\gerbers\\PCB_Fabrication_Data_in_Gerber_Example_2_Copper$L2$Inr.gbr"
	//content, err := ioutil.ReadFile(sf)
	//if err != nil {
	//	t.Fatal(err)
	//}
//	t.Log(string(content))

	var contents string
	for j := range td {
		contents = contents + td[j].input
	}
	content := []byte(contents)

	result := SplitByGCommands(&content)

	if len(*result) != len(td) {
		t.Error("len(*result) != len(td) :" + strconv.Itoa(len(*result)) +"!=" + strconv.Itoa(len(td)) )
		for i := range *result {
			t.Logf("%s\n",(*result)[i].String())
		}
	} else {
		for i := range *result {
			if (*result)[i].String() != td[i].answer {
				t.Error("\ninput: " + td[i].input + "\ncorr: " + td[i].answer + "\nres: " + (*result)[i].String())
			}
		}
	}
}

func TestFormatGCode(t *testing.T) {
	var testarray = []string{
		"D", "00",
		"D", "00000",
		"D", "001000",
		"B", "1",
		"G", "ss00ss",
		"EE", "",
		"D", "1",
		"D", "9",
		"F", "10",}
	for i:=0; i < len(testarray); i = i+2 {
		t.Log(testarray[i] + testarray[i+1] + "->" + FormatGCode(testarray[i],testarray[i+1]))
	}
}
