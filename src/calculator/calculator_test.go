package calculator

import (
	"strconv"
	"testing"
)

func TestOperation_Calc(t *testing.T) {
	val1 := 222.222
	val2 := 333.333
	op1 := Operand{"op1", val1, nil}
	op2 := Operand{"op2", val2, nil}
	oper1 := Operation{&op1, &op2, Add}
	op3 := Operand{"op3", 0, &oper1}

	oper2 := Operation{&op3, &Operand{"op4", 0.0, nil}, Neg}

	if op3.Calc() != (val1 + val2) {
		t.Fatal("op3.Calc() error!")
	}

	if oper2.Calc() != -(val1 + val2) {
		t.Fatal("oper2.Calc() error!")
	}
}

type testCase struct {
	src string
	ans float64
}

var src = []testCase{
	{"-2x3", -2 * 3},
	{"-2X-3", -2 * -3},
	{"2x3", 2 * 3},
	{"(((-2)))", -2},
	{"2--3", 2- -3},
	{"2/-3.0", 2 / -3.0},
	{"-2--3", -2 - (-3)},
	{"-2+1-1", -2 + 1 - 1},
	{"2+1-1", 2 + 1 - 1},
	{"-2+1--3", -2 + 1 - (-3)},
	{"-6x9/8", -6 * 9 / 8.0},
	{"-6x9/8x8/-4X787.33", -6 * 9 / 8.0 * 8/-4 * 787.33},
	{"-6x9/1x-6x9/2/-6x9/3", -6*9/1*-6*9/2/-6*9/3},
	{"-1", -1},
	{"(-2x(333+444x4343)/555)-(666-(-777x(888x(-999--1000))))+(11-12)", -697593},
}

func TestNewOperand(t *testing.T) {

	for _, s := range src {
		//		operand := NewOperand(s.src, s.src)
		//		NewOperand(s.src, s.src).Calc()
		if NewOperand(s.src, s.src, false).Calc() != s.ans {
			t.Fatal(s.src + " calculation error! got " +
				strconv.FormatFloat(NewOperand(s.src, s.src, false).Calc(), 'f', 10, 64) +
				" expected " + strconv.FormatFloat(s.ans, 'f', 10, 64))
		} else {
			t.Log(s.src + " = " + strconv.FormatFloat(s.ans, 'f', 5, 64))
		}
	}
}

var src2 = []string{
	"(-2x(333+444x4343)/555)-(666-(-777x(888x(-999--1000))))+(11-12)",
}

func TestCalcExpression(t *testing.T) {
	for _, s := range src {
		result := CalcExpression(s.src)
		if result != s.ans {
			t.Fatal(s.src + " calculation error! got " +
				strconv.FormatFloat(result, 'f', 10, 64) +
				" expected " + strconv.FormatFloat(s.ans, 'f', 10, 64))
		} else {
			t.Log(s.src + " = " + strconv.FormatFloat(s.ans, 'f', 5, 64))
		}
	}
}

//func TestTokenizeFormulae(t *testing.T) {
//	for _, s := range src {
//		tf := TokenizeFormulae(s.src)
//		t.Log(s.src)
//		result := ""
//		for v := range tf {
//			result = result + tf[v].String()
//		}
//		t.Log(result)
//		floatResult := CalcTokenizedFormulae(&tf)
//		if  floatResult != s.ans {
//			t.Fatal(s.src + " calculation error! got " +
//				strconv.FormatFloat(floatResult, 'f', 10, 64) +
//				" expected " + strconv.FormatFloat(s.ans, 'f', 10, 64))
//		} else {
//			t.Log(s.src + " = " + strconv.FormatFloat(s.ans, 'f', 5, 64))
//		}
//	}
//}
