package gerbparser

import (
	"strings"
	"testing"
)

func TestAMPrimitiveType_String(t *testing.T) {

	testArray := [8]AMPrimitiveType{AMPrimitive_Comment, AMPrimitive_Circle, AMPrimitive_VectLine, AMPrimitive_CenterLine, AMPRimitive_OutLine, AMPrimitive_Polygon,
		AMPrimitive_Moire, AMPrimitive_Thermal}

	answers := [9]string{
		"comment",
		"circle",
		"vector line",
		"center line",
		"outline",
		"polygon",
		"moire",
		"thermal",
		"unknown",
	}

	for i, s := range testArray {
		if strings.Compare(s.String(), answers[i]) != 0 {
			t.Error("Error! " + s.String() + "!=" + answers[i])
		}
	}
}

func TestAMPrimitive_String(t *testing.T) {
	aMPrimitive := AMPrimitivePolygon{AMPrimitive_Polygon, []string {"0zzz", "1xxxxxx", "", "MODIFIER",}}
	t.Logf(aMPrimitive.String())
}

func TestApertureMacro_String(t *testing.T) {
	aMPrimitive1 := AMPrimitivePolygon{AMPrimitive_Polygon,
	[]string {"polygon modifier 1", "polygon modifier 2", "polygon modifier 3", "polygon modifier 4",}}
	aMPrimitive2 := AMPrimitiveCircle{AMPrimitive_Circle,
	[]string {"circle modifier 1", "circle modifier 2", "circle modifier 3", "circle modifier 4",}}
	aMPrimitive3 := AMPrimitiveThermal{AMPrimitive_Thermal,
	[]string {"thermal modifier 1", "thermal modifier 2", "thermal modifier 3", "thermal modifier 4",}}
	aMPrimitive4 := AMPrimitiveComment{AMPrimitive_Comment,
		[]string {}}

	testApertureMacro := ApertureMacro{"Aperture macro name",
	[]string{"1st comment string", "2nd comment string","3rd comment string",},
	[]AMVariable{ AMVariable{"AM VAR Name1", "val1", 0},
	AMVariable{"AM VAR Name2", "val2", 2}, AMVariable{"AM VAR Name2","",0},},
	[]AMPrimitive{aMPrimitive1, aMPrimitive2, aMPrimitive3, aMPrimitive4}}

	t.Log(testApertureMacro.String())

}

func TestNewApertureMacro(t *testing.T) {
	srcString1 := "%AMBox*\n0 Rectangle with rounded corners, with rotation*\n0 The origin of the aperture is it’s center*\n0 $1 X-size*\n0 $2 Y-size*\n" +
		"0 $3 Rounding radius*\n0 $4 Rotation angle, in degrees counterclockwise*\n0 Add two overlapping rectangle primitives as box body*\n" +
		"21,1,$1,$2-$3-$3,0,0,$4*\n21,1,$2-$3-$3,$2,0,0,$4*\n0 Add four circle primitives for the rounded corners*\n$5=$1/2*\n$6=$2/2*\n$7" + "=" + "2X$3*\n1,1,$7,$5-$3,$6-$3,$4*\n" +
		"1,1,$7,-$5+$3,$6-$3,$4*\n1,1,$7,-$5+$3,-$6+$3,$4*\n1,1,$7,$5-$3,-$6+$3,$4*%"

	srcString2 := "%AMTRIANGLE_30*\n4,1,3,\n1,-1,\n1,1,\n2,1,\n1,-1,\n30*%"

	srcString3 := "%AMTARGET*\n1,1,$1,0,0*\n$1=$1x0.8*\n1,0,$1,0,0*\n$1=$1x0.8*\n1,1,$1,0,0*\n$1=$1x0.8*\n1,0,$1,0,0*\n$1=$1x0.8*\n1,1,$1,0,0*\n$1=$1x0.8*\n1,0,$1,0,0*%"
	/*
	%AMTARGET*
	1,1,$1,0,0*
	$1=$1x0.8*
	1,0,$1,0,0*
	$1=$1x0.8*
	1,1,$1,0,0*
	$1=$1x0.8*
	1,0,$1,0,0*
	$1=$1x0.8*
	1,1,$1,0,0*
	$1=$1x0.8*
	1,0,$1,0,0*%
	 */

	srcString4 := "%AMTEST*\n1,1,$1,$2,$3*\n$4=$1x0.75*\n$5=($2+100)x1.75*\n1,0,$4,$5,$3*%\n"
	/*
	%AMTEST*
	1,1,$1,$2,$3*
	$4=$1x0.75*
	$5=($2+100)x1.75*
	1,0,$4,$5,$3*%
	 */

	srcStrings := []string{srcString1, srcString2, srcString3, srcString4}
	for i:= range srcStrings {
		aMacro, err := NewApertureMacro(srcStrings[i])
		if err != nil {
			t.Error("Error creating new aperture macro")
			t.Error(err)
		} else {
			t.Log(aMacro.String())
		}
	}

}
