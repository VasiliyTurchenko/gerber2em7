// Code generated by "stringer -type=GerberCommandId"; DO NOT EDIT.

package geberlexer

import "strconv"

const _GerberCommandId_name = "ABADAMASDD01D02D03FSG01G02G03G04G36G37G54G55G70G71G74G75G90G91INIPIRLMLNLPLRLSM00M01M02MIMOOFSFSRTATDTFTONOP"

var _GerberCommandId_index = [...]uint8{0, 2, 4, 6, 8, 9, 12, 15, 18, 20, 23, 26, 29, 32, 35, 38, 41, 44, 47, 50, 53, 56, 59, 62, 64, 66, 68, 70, 72, 74, 76, 78, 81, 84, 87, 89, 91, 93, 95, 97, 99, 101, 103, 105, 108}

func (i GerberCommandId) String() string {
	if i >= GerberCommandId(len(_GerberCommandId_index)-1) {
		return "GerberCommandId(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _GerberCommandId_name[_GerberCommandId_index[i]:_GerberCommandId_index[i+1]]
}