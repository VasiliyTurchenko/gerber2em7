package blockapertures

import (
	"fmt"
	"gerberstates"
)

type BlockAperture struct {
	StartStringNum int
	Code           int
	BodyStrings    []string
	StepsPtr       []*gerberstates.State
}

//Print info
func (ba *BlockAperture) Print() {
	fmt.Println("\n***** Block aperture *****")
	fmt.Println("\tBlock aperture code:", ba.Code)
	fmt.Println("\tSource strings:")
	for b := range ba.BodyStrings {
		fmt.Println("\t\t", b, "  ", ba.BodyStrings[b])
	}
	fmt.Println("\tResulting steps:")
	for b := range ba.StepsPtr {
		ba.StepsPtr[b].Print()
	}
}
