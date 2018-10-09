package geberlexer

import (
	"glog_t"
	"strconv"
	"strings"
	"unicode"
	"xy"
)

/*
FS Format specification. Sets the coordinate format, e.g. the number of decimals. 4.1
MO Mode. Sets the unit to inch or mm. 4.2
AD Aperture define. Defines a template based aperture and assigns a D code to it. 4.3
AM Aperture macro. Defines a macro aperture template. 4.5
AB Aperture block. Defines a block aperture and assigns a D-code to it. 4.6
Dnn (nn≥10) Sets the current aperture to D code nn. 4.7
D01 Interpolate operation. Outside a region statement D01 creates a draw or arc
object using the current aperture. Inside it creates a linear or circular contour
segment. After the D01 command the current point is moved to draw/arc end
point.
4.8
D02 Move operation. D02 does not create a graphics object but moves the current
point to the coordinate in the D02 command.
4.8
D03 Flash operation. Creates a flash object with the current aperture. After the D03
command the current point is moved to the flash point.
4.8
G01 Sets the interpolation mode to linear. 4.9
G02 Sets the interpolation mode to clockwise circular. 4.10
G03 Sets the interpolation mode to counterclockwise circular. 4.10
G74 Sets quadrant mode to single quadrant. 4.10
G75 Sets quadrant mode to multi quadrant. 4.10
LP Load polarity. Loads the polarity object transformation parameter. 4.11.2
LM Load mirror. Loads the mirror object transformation parameter. 4.11.3
LR Load rotation. Loads the rotation object transformation parameter. 4.11.4
LS Load scale. Loads the scale object transformation parameter. 4.11.5
G36 Starts a region statement. This creates a region by defining its contour. 4.12.
G37 Ends the region statement. 4.12
SR Step and repeat. Open or closes a step and repeat statement. 4.13
G04 Comment. 4.14
TF Attribute file. Set a file attribute. 5.2
TA Attribute aperture. Add an aperture attribute to the dictionary or modify it. 5.3
TO Attribute object. Add an object attribute to the dictionary or modify it. 5.4
TD Attribute delete. Delete one or all attributes in the dictionary. 5.5
M02 End of file. 4.1
*/

/*
G54 Select aperture This historic code optionally precedes an aperture
selection D-code. It has no effect.
Sometimes used.
G55 Prepare for flash This historic code optionally precedes D03 code. It
has no effect.
Very rarely used nowadays.
G70 Set the ‘Unit’ to inch These historic codes perform a function handled by
the MO command. See 4.2.
G71 Set the ‘Unit’ to mm Sometimes used.
G90 Set the ‘Coordinate format’ to ‘Absolute
notation’
These historic codes perform a function handled by
the FS command. See 4.1.
Very rarely used nowadays.
G91 Set the ‘Coordinate format’ to ‘Incremental
notation’
M00 Program stop This historic code has the same effect as M02. See
4.14.
Very rarely used nowadays.
M01 Optional stop This historic code has no effect.
Very rarely used nowadays.
IP Sets the ‘Image polarity’ graphics state
parameter
IP can only be used once, at the beginning of the
file.
Sometimes used, and then usually as %IPPOS*%
to confirm the default – a positive image; it then
has no effect. As it is not clear how %IPNEG*%
must be handled it is probably a waste of time to
try to fully implement it, and sufficient to give a
warning if an image is negative.
AS Sets the ‘Axes correspondence’ graphics
state parameter
These commands can only be used once, at the
beginning of the file. The order of execution is
always MI, SF, OF, IR and AS, independent of
IR Sets ‘Image rotation’ graphics state their order of appearance in the file.
parameter
MI Sets ‘Image mirroring’ graphics state
parameter
Rarely used nowadays. If used it is almost always
to confirm the default value; they have no effect. It
is probably a waste of time to fully implement
these commands; simply ignoring them will
handle the overwhelming majority of Gerber files
correctly; issuing a warning when used with a
non-default value protects the reader in the very
rare cases this occurs.
OF Sets ‘Image offset’ graphics state parameter
SF Sets ‘Scale factor’ graphics state parameter
IN Sets the name of the file image. Has no
effect. It is comment.
These commands can only be used once, at the
beginning of the file.
Use G04 for comments. See 4.14.
Sometimes used.
LN Loads a name. Has no effect. It is a
comment.
Can be used many times in the file.
Use G04 for comments. See 4.14.
Sometimes used.
*/

//go:generate stringer -type=GerberCommandId

const MaxUint = ^uint(0)
const MaxInt = int(MaxUint >> 1)

//type GCommander interface {
//	String() string
//}

type GerberCommandId byte

const (
	AB GerberCommandId = iota
	AD
	AM
	AS
	D
	D01
	D02
	D03
	FS
	G01
	G02
	G03
	G04
	G36
	G37
	G54
	G55
	G70
	G71
	G74
	G75
	G90
	G91
	IN
	IP
	IR
	LM
	LN
	LP
	LR
	LS
	M00
	M01
	M02
	MI
	MO
	OF
	RO
	SF
	SR
	TA
	TD
	TF
	TO
	// must be last
	NOP
)

var GCmdBaseArray = []GerberCommandId{
	D,
	D01,
	D02,
	D03,
	G01,
	G02,
	G03,
	G04,
	G36,
	G37,
	G54,
	G55,
	G70,
	G71,
	G74,
	G75,
	G90,
	G91,
	M00,
	M01,
	M02,
}
var GCmdExtArray = []GerberCommandId{
	AB,
	AD,
	AM,
	AS,
	FS,
	IN,
	IP,
	IR,
	LM,
	LN,
	LP,
	LR,
	LS,
	MI,
	MO,
	OF,
	RO,
	SF,
	SR,
	TA,
	TD,
	TF,
	TO,
}

/* XY initializer */

var lexFS = xy.FormatSpec{"", "", 0, 0, 0, 0, 1.0}

// first initialization
func (gc *GerberCommand) Init() {
	switch gc.cmd {
	case FS:
		lexFS.Init("%FS"+gc.cmdString+"%", "%MOMM*%")
	case MO:
		if gc.cmdString == "IN" {
			lexFS.MUString = "%MOIN*%"
			lexFS.MU = xy.InchesToMM
		}
	case D01, D02, D03:
		gc.xy = xy.NewXY()
		gc.xy.Init(gc.cmdString+"D", &lexFS, nil)
	case G04, TF, TA:

	default:
		glog_t.Error("Init() is not implemented: " + gc.String())
	}
}

var GerberCommandNr int = 0

type GerberCommand struct {
	cmd       GerberCommandId
	cmdString string
	cmdNumber int
	xy        *xy.XY
}

func (gc *GerberCommand) String() string {
	xyStr := ""
	if gc.xy != nil {
		xyStr = `,xy:"` + gc.xy.String()
	}
	return "{cmd#:" + strconv.Itoa(gc.cmdNumber) + ",cmd:\"" + gc.cmd.String() + "\",val:\"" + gc.cmdString + xyStr + "\"}"
}

func NewGerberCommand(cmd GerberCommandId) GerberCommand {
	retVal := new(GerberCommand)
	retVal.cmd = cmd
	retVal.cmdString = "<empty>"
	retVal.cmdNumber = GerberCommandNr
	GerberCommandNr++
	return *retVal
}

type Delim byte

const (
	DataBlockTrailer Delim = '*'
	ExtCmdDelimiter  Delim = '%'
)

func (d Delim) String() string {
	switch d {
	case DataBlockTrailer:
		return "DBEND"
	case ExtCmdDelimiter:
		return "EXTCMD"
	default:
		return string(d)
	}
}

// deletes leading '0'
func FormatGCode(sym string, num string) string {

	if num == "" {
		return sym
	}

	num = strings.TrimLeft(num, "0")

	if len(num) == 1 {
		return sym + "0" + num
	}
	if len(num) == 0 {
		return sym + "00"
	}

	return sym + num
}

func SplitByGCommands2(buf []byte) *[]GerberCommand {
	retVal := make([]GerberCommand, 0)
	extCmdStartPos := -1
	extCmdEndPos := -1
	baseCmdStartPos := -1
	baseCmdEndPos := -1
	for i := range buf {
		// purge CR LF
		if buf[i] == 0x0A || buf[i] == 0x0D {
			continue
		}

		if extCmdStartPos == -1 {
			if unicode.IsSpace(rune((buf)[i])) == true {
				continue
			}
		}

		if buf[i] == byte(ExtCmdDelimiter) {
			if extCmdStartPos == -1 && extCmdEndPos == -1 {
				extCmdStartPos = i
				continue
			}
			if extCmdStartPos != -1 && extCmdEndPos == -1 {
				extCmdEndPos = i
			}
			extCmdString := string(buf[extCmdStartPos+1 : extCmdEndPos])
			extCmd := parseExtCmd(extCmdString)
			if extCmd != nil {
				retVal = append(retVal, *extCmd)
			}
			extCmdStartPos = -1
			extCmdEndPos = -1
			continue
		}
		if extCmdStartPos == -1 {
			if baseCmdStartPos == -1 && baseCmdEndPos == -1 {
				if buf[i] != byte(DataBlockTrailer) {
					baseCmdStartPos = i
				}
				continue
			}
			if buf[i] == byte(DataBlockTrailer) && baseCmdStartPos != -1 {
				baseCmdEndPos = i

				baseCmdString := string(buf[baseCmdStartPos:baseCmdEndPos])

				for {
					// check for concatenated commands like G01D45XnnnYnnnD01
					if len(baseCmdString) > 3 &&
						(baseCmdString[0] == 'G' || baseCmdString[0] == 'D') &&
						baseCmdString[0:3] != "G04" &&
						baseCmdString[0:2] != "G4" {
						notNumPos := MaxInt
						for k := 1; k < len(baseCmdString); k++ {
							if strings.IndexAny(baseCmdString[k:k+1], "01234567890") == -1 {
								notNumPos = k
								break
							}
						}
						if notNumPos < len(baseCmdString) {
							baseCmd := parseBaseCmd(baseCmdString[:notNumPos])
							baseCmdString = baseCmdString[notNumPos:]
							if baseCmd != nil {
								retVal = append(retVal, *baseCmd)
							} else {
								baseCmdStartPos = -1
								baseCmdEndPos = -1
								continue
							}

						} else {
							break
						}
					} else {
						break
					}
				}
				baseCmd := parseBaseCmd(baseCmdString)
				if baseCmd != nil {
					retVal = append(retVal, *baseCmd)
				}
				baseCmdStartPos = -1
				baseCmdEndPos = -1
				continue
			}
		}
	}
	for gc := range retVal {
		retVal[gc].Init()
	}
	return &retVal
}

func parseExtCmd(in string) *GerberCommand {
	if len(in) < 3 {
		return nil
	}
	retVal := new(GerberCommand)
	for j := range GCmdExtArray {
		if in[:2] == GCmdExtArray[j].String() {
			*retVal = NewGerberCommand(GCmdExtArray[j])
			(*retVal).cmdString = in[2:]
			return retVal
		}
	}
	glog_t.Error("The string isn't parsed: ", in)
	return nil
}

func parseBaseCmd(in string) *GerberCommand {
	in = strings.TrimSpace(in)
	if len(in) < 2 {
		return nil
	}
	retVal := new(GerberCommand)
	cmdString := ""
	// filter comments
	if strings.HasPrefix(in, "G04") || strings.HasPrefix(in, "G4") {
		cmd := "G04"
		var cmdString string = ""
		if len(in) > 3 && strings.HasPrefix(in, "G04") {
			cmdString = in[3:]
		} else if len(in) > 2 && strings.HasPrefix(in, "G4") {
			cmdString = in[2:]
		}
		for j := range GCmdBaseArray {
			if cmd == GCmdBaseArray[j].String() {
				*retVal = NewGerberCommand(GCmdBaseArray[j])
				(*retVal).cmdString = cmdString
				return retVal
			}
		}
	}

	// fix implicit D01*
	impl := true
	for m := range in {
		if strings.IndexAny(in[m:m+1], "+-XYIJ0123456789") == -1 {
			impl = false
			break
		}
	}
	if impl == true {
		in = in + "D01"
	}

	// find rightmost byte of D|G|M
	for i := len(in); i > 0; i-- {
		if in[i-1] == 'D' || in[i-1] == 'G' || in[i-1] == 'M' {
			cmd := FormatGCode(in[i-1:i], in[i:])
			if cmd[0] == 'D' && cmd != "D01" && cmd != "D02" && cmd != "D03" {
				cmdString = cmd[1:]
				if _, err := strconv.Atoi(cmdString); err != nil {
					break
				}
				cmd = "D"
			} else {
				cmdString = in[:i-1]
			}
			for k := range cmdString {
				if strings.IndexAny(cmdString[k:k+1], "+-XYIJ0123456789") == -1 {
					cmd = ""
					break
				}
			}
			for j := range GCmdBaseArray {
				if cmd == GCmdBaseArray[j].String() {
					*retVal = NewGerberCommand(GCmdBaseArray[j])
					(*retVal).cmdString = cmdString
					return retVal
				}
			}
		}
	}
	glog_t.Error("The string isn't parsed: ", in)
	return nil
}
