package regions

import (
	"errors"
	"strconv"
	. "xy"
)

/*####################  regions ##################################
 */
type Region struct {
	startXY         *XY // pointer to start entry
	numberOfXY      int // number of entries
	G36StringNumber int // number of the string with G36 cmd
	G37StringNumber int // number of the string with G37 cmd
}

func (region *Region) String() string {

	xyText := "<nil>"
	if region == nil {
		return xyText
	}

	if region.startXY != nil {
		xyText = region.startXY.String()
	}
	return "Region:\n" +
		"\t\tstart point: " + xyText + "\n" +
		"\t\tcontains " + strconv.Itoa(region.numberOfXY) + " vertices\n" +
		"\t\tG36 command is at line " + strconv.Itoa(region.G36StringNumber) + "\n" +
		"\t\tG37 command is at line " + strconv.Itoa(region.G37StringNumber)
}

// creates and initialises a region object
func NewRegion(strNum int) *Region {
	retVal := new(Region)
	retVal.G36StringNumber = strNum
	retVal.numberOfXY = 0
	retVal.G37StringNumber = -1
	return retVal
}

// closes the region
func (region *Region) Close(strnum int) error {
	if region == nil {
		return errors.New("can not close the contour referenced by null pointer")
	}
	region.G37StringNumber = strnum
	return nil
}

// sets a start coordinate entry
func (region *Region) SetStartXY(in *XY) {
	region.startXY = in
	region.numberOfXY++
}

// returns a start coordinate entry
func (region *Region) getStartXY() *XY {
	return region.startXY
}

// increments number of coordinate entries
func (region *Region) IncNumXY() int {
	region.numberOfXY++
	return region.numberOfXY
}

// returns the number of coordinate entries of the contour
func (region *Region) GetNumXY() int {
	return region.numberOfXY
}

// returns true if region is opened
func (region *Region) IsRegionOpened() (bool, error) {
	if region == nil {
		return false, errors.New("bad region referenced (by nil ptr)")
	}
	if region.G37StringNumber == -1 {
		return true, nil
	} else {
		return false, nil
	}
}
