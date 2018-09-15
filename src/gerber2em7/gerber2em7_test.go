package gerber2em7

import (
	"fmt"
	"strconv"
	"testing"
)

func TestTokenizeGerber(t *testing.T) {
	testCase1 := []byte ("\n\t\t   0000****%1111*******%\n22222")
	testCase2 := []byte ("%11\n\r\tkkkkkkk%\n")

	ans := TokenizeGerber(&testCase1)
	for i := range *ans {
		fmt.Println(strconv.Itoa(i) + ": " + (*ans)[i])
	}

	ans = TokenizeGerber(&testCase2)
	for i := range *ans {
		fmt.Println(strconv.Itoa(i) + ": " + (*ans)[i])
	}


}
