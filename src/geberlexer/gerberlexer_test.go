package geberlexer

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	//	"io/ioutil"
	"strconv"
	"testing"
)

func TestGerberCommand_String(t *testing.T) {
	cmd := []GerberCommand{
		{FS, "the body of FS Command", 0, nil},
		{MO, "the body of MO Command", 0, nil},
		{AM, "the body of AM Command", 0, nil},
		{G04, "COMMENT!!!", 555, nil},
		{M02, "STOP", 777, nil},
	}
	for i := range cmd {
		t.Logf("%s\n", cmd[i].String())
	}
}

func TestGerberCommandId_String(t *testing.T) {
	for i := range GCmdBaseArray {
		t.Logf("%s\n", GCmdBaseArray[i].String())
	}
	for i := range GCmdExtArray {
		t.Logf("%s\n", GCmdExtArray[i].String())
	}

}

type testdata struct {
	input  string
	answer string
}

var td = []testdata{

	{`G04 Title: $Id$, silkscreen component side *
G04 Creator: pcb 1.6.3 *
G04 CreationDate: Wed Mar  8 07:54:16 2000 UTC *
G04 For: neil *
G04 Format: Gerber/RS-274X *
G04 PCB-Dimensions: 3875 2425 *
G04 PCB-Coordinate-Origin: lower left *
G04 Color: R0 G0 B0 *
*
%FSLAX23Y23*%
%MOIN*%
%ADD11C,0.015*%
%ADD12C,0.025*%
%ADD13C,0.040*%
%ADD14C,0.070*%
%ADD15C,0.100*%
%ADD16C,0.010*%
%ADD17C,0.008*%
%ADD18R,0.070X0.070*%
%ADD19R,0.100X0.100*%
%ADD20C,0.110*%
%ADD21C,0.140*%
%ADD22R,0.080X0.080*%
%ADD23R,0.110X0.110*%
%ADD24C,0.080*%
%ADD25R,0.060X0.060*%
%ADD26R,0.090X0.090*%
%ADD27C,0.060*%
%ADD28C,0.090*%
%ADD29C,0.075*%
%ADD30C,0.105*%
%ADD31C,0.150*%
%ADD32C,0.180*%
%ADD33C,0.180X0.150*%
%ADD34C,0.009*%
%IPPOS*%
G01*
G54D16*X40Y1950D02*G75G03I160J0X40Y1951D01*G01*
M02*
`, ""},
}

func TestSplitByGCommands(t *testing.T) {

	dir := "G:/go_prj/gerber2em7/gerbers"

	files, err := ioutil.ReadDir(filepath.FromSlash(dir))
	if err != nil {
		t.Fatal(err)
	}

	var content []byte

	for _, f := range files {
		fmt.Println("BEGIN FILE ------------------------------------------------------")
		//		ff := f.Name()
		ff := "PCB_Fabrication_Data_in_Gerber_Example_1_Copper$L1.gbr"

		fmt.Println(f.Name())
		content, err = ioutil.ReadFile(filepath.Join(filepath.FromSlash(dir), ff))
		if err != nil {
			t.Fatal(err)
		}
		result := SplitByGCommands2(content)

		for i := range *result {
			t.Logf("%s\n", (*result)[i].String())
		}
		break
	}

}

func TestSplitByGCommandsArr(t *testing.T) {
	var contents string
	for j := range td {
		contents = contents + td[j].input
	}
	content := []byte(contents)

	result := SplitByGCommands2(content)

	if len(*result) != len(td) {
		t.Error("len(*result) != len(td) :" + strconv.Itoa(len(*result)) + "!=" + strconv.Itoa(len(td)))
		for i := range *result {
			t.Logf("%s\n", (*result)[i].String())
		}
	} else {
		for i := range *result {
			if (*result)[i].String() != td[i].answer {
				t.Error("\ninput: " + td[i].input + "\ncorr: " + td[i].answer + "\nres: " + (*result)[i].String())
			} else {
				t.Logf("OK\ninput: %s\ncorr: %s\nres : %s", td[i].input, td[i].answer, (*result)[i].String())
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
		"F", "10"}
	for i := 0; i < len(testarray); i = i + 2 {
		t.Log(testarray[i] + testarray[i+1] + "->" + FormatGCode(testarray[i], testarray[i+1]))
	}
}

var commentTest = []string{
	"", "", "n",
	"bcvbcvb", "", "n",
	"G", "", "n",
	"G1234", "", "n",
	"G4", "", "y",
	"G4zzzzzzzzxzxzxccdewe5r", "zzzzzzzzxzxzxccdewe5r", "y",
	"G04 bfgdgf", " bfgdgf", "y",
	"G00004bfgdgf", " bfgdgf", "y",
	"G400004bfgdgf", " bfgdgf", "y",
}

func TestIsGerberComment(t *testing.T) {

}
