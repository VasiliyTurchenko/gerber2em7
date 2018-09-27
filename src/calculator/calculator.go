package calculator

import (
	"strconv"
	"strings"
)

type Calculator interface {
	Calc() float64
}

type Operand struct {
	variableName string
	value        float64
	operation    *Operation
}

func (op *Operand) Calc() float64 {
	if op.operation == nil {
		return op.value
	}
	return op.operation.Calc()
}

type Operation struct {
	firstOperand  *Operand
	secondOperand *Operand
	operation     OpCode
}

type OpCode int

const (
	Nop OpCode = iota
	Add OpCode = iota + 1
	Sub
	Mul
	Div
	Neg
	Plus
)

func (oc OpCode) String() string {
	switch oc {
	case Add, Plus:
		return "+ "
	case Sub, Neg:
		return "- "
	case Mul:
		return "x "
	case Div:
		return "/ "
	case Nop:
		return "<nop> "
	default:
		return "bad OpCode "
	}
	return ""
}

func (op *Operation) Calc() float64 {
	if op.firstOperand == nil {
		panic("calculator: first operand = nil")
	}
	switch op.operation {
	case Neg:
		return -op.firstOperand.Calc()
	case Plus:
		return op.firstOperand.Calc()
	}
	if op.secondOperand == nil {
		panic("calculator: second operand = nil")
	}
	switch op.operation {
	case Add:
		return op.firstOperand.Calc() + op.secondOperand.Calc()
	case Sub:
		return op.firstOperand.Calc() - op.secondOperand.Calc()
	case Mul:
		return op.firstOperand.Calc() * op.secondOperand.Calc()
	case Div:
		return op.firstOperand.Calc() / op.secondOperand.Calc()
	default:
		panic("calculator: bad opcode " + strconv.Itoa(int(op.operation)))
	}
}

func NewOperand(str string, name string, inv bool) *Operand {
	retVal := new(Operand)
	(*retVal).variableName = name
	str = strings.TrimSpace(str)
	panicString := "NewOperand() is unable to parse " + str
	opSymbols := "+-xX/"
	opPos1 := strings.IndexAny(str, opSymbols)
	if opPos1 == len(str)-1 {
		panic(panicString)
	}
	opPos2 := -1
	if opPos1 != -1 {
		opPos2 = strings.IndexAny(str[opPos1+1:], opSymbols)
		if opPos2 != -1 {
			opPos2 = opPos2 + opPos1 + 1
		}
	}
	var err error

	if (opPos1 == -1) || (opPos1 == 0 && opPos2 == -1) {
		// try to represent str as constant value
		retVal.value, err = strconv.ParseFloat(str, 64)
		if err != nil {
			panic(panicString)
		}
		if inv == true {
			retVal.value = 1 / retVal.value
		}
		return retVal
	}
	if opPos1 == 0 {
		opPos1 = opPos2
	}
	var opCode OpCode
	switch str[opPos1] {
	case '+':
		opCode = Add
	case '-':
		opCode = Sub
	case '/':
		opCode = Div
	case 'x', 'X':
		opCode = Mul
	}
	op1 := NewOperand(str[:opPos1], "", inv)

	// div to mul conversion
	needInv := false
	if opCode == Div {
		opCode = Mul
		needInv = true
	}

	op2 := NewOperand(str[opPos1+1:], "", needInv)
	retVal.operation = &Operation{op1, op2, opCode}
	return retVal
}

type Stack struct {
	data []int
}

func NewStack() *Stack {
	return &Stack{}
}

func (stack *Stack) Push(val int) {
	(*stack).data = append((*stack).data, val)
}

func (stack *Stack) Pop() int {
	slen := len((*stack).data)
	if slen == 0 {
		panic("Pop() error: stack is empty")
	}
	retVal := (*stack).data[slen-1]
	(*stack).data = (*stack).data[:slen-1]
	return retVal
}

//
func CalcExpression(str string) float64 {

//	log.SetFlags(log.Lshortfile)

	stack := NewStack()
	varStorage := make(map[string]float64)
	var tempVarId = 0
	var valName = ""
	const LEFT_PAR = '('
	const RIGHT_PAR = ')'
	var r rune
	var i int
	str = "(" + str + ")"
	for {
		for i, r = range str {
			if r == LEFT_PAR {
				stack.Push(i)
				continue
			}
			if r == RIGHT_PAR {
				lPar := stack.Pop()
				substring := str[lPar+1 : i]
//				log.Println(substring)
				tf := TokenizeFormulae(substring, &varStorage)
				val := CalcTokenizedFormulae(&tf)
				valName = "$$" + strconv.Itoa(tempVarId)
				tempVarId++
				varStorage[valName] = val
//				log.Println("Variable " + valName + " = " + strconv.FormatFloat(val, 'f', 10, 64))
				str = strings.Replace(str, str[lPar:i+1], valName, 1)
//				log.Println("Reduced str = " + str)
				break
			}
		}
		if strings.Compare(str, valName) == 0 {
			break
		}
	}
	return varStorage[str]
}

type TokenizedFormula struct {
	value     float64
	operation OpCode
}

func (tf *TokenizedFormula) String() string {
	return strconv.FormatFloat((*tf).value, 'f', 10, 64) + " " + (*tf).operation.String()
}

func TokenizeFormulae(str string, varStorage *map[string]float64 ) []TokenizedFormula {
	retVal := make([]TokenizedFormula, 0)
	runeStr := []rune(str)
	tokenStart := true
	needInvNext := false
	needNegNext := false
	convString := ""
	var opCode OpCode
	var floatVal float64
	var err error
	panicString := "TokenizeFormulae() is unable to parse " + str
	for i := 0; i < len(runeStr); i++ {
		if tokenStart == true {
			if runeStr[i] == '+' || runeStr[i] == '-' {
				tokenStart = false
				convString = convString + string(runeStr[i])
				continue
			}
		}
		var needNN, needIN bool
		switch runeStr[i] {
		case '+':
			opCode = Add
			needNN = false
			needIN = false
		case '-':
			opCode = Add
			needNN = true
			needIN = false
		case '/':
			opCode = Mul
			needIN = true
			needNN = false
		case 'x', 'X':
			opCode = Mul
			needIN = false
			needNN = false
		default:
			convString = convString + string(runeStr[i])
			tokenStart = false
			continue
		}
		if strings.HasPrefix(convString, "$") {
			floatVal = (*varStorage)[convString]
		} else {
			floatVal, err = strconv.ParseFloat(convString, 64)
			if err != nil {
				panic(panicString)
			}
		}
		if needInvNext {
			floatVal = 1 / floatVal
		}
		if needNegNext {
			floatVal = -floatVal
		}
		retVal = append(retVal, TokenizedFormula{floatVal, opCode})
		tokenStart = true
		convString = ""
		needInvNext = needIN
		needNegNext = needNN
	}
	// last token ...
	if strings.HasPrefix(convString, "$") {
		floatVal = (*varStorage)[convString]
	} else {

		floatVal, err = strconv.ParseFloat(convString, 64)
		if err != nil {
			panic(panicString)
		}
	}
	if needInvNext {
		floatVal = 1 / floatVal
		needInvNext = false
	}
	if needNegNext {
		floatVal = -floatVal
		needNegNext = false
	}
	retVal = append(retVal, TokenizedFormula{floatVal, Nop})

	return retVal
}

func CalcTokenizedFormulae(tf *[]TokenizedFormula) float64 {
	retVal := 0.0
	mulVal := 1.0
	startMul := false
	i := 0
	ii := len(*tf) - 1
	for {
		if (*tf)[i].operation == Mul || startMul == true {
			mulVal = mulVal * (*tf)[i].value
			startMul = true
			if i == ii {
				retVal = retVal + mulVal
				break
			} else {
				i++
				continue
			}
		}
		if startMul == true {
			// the last multiplicator
			mulVal = mulVal * (*tf)[i].value
			retVal = retVal + mulVal
			mulVal = 1.0
			startMul = false
		} else {
			retVal = retVal + (*tf)[i].value
		}
		if i == ii {
			break
		}
		i++
	}
	return retVal
}
