package geberlexer

import (
	"strings"
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

type GCommander interface {
	String() string
}

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
	SF
	SR
	TA
	TD
	TF
	TO
	// must be last
	NOP
)

var GCmdBaseArray = []GerberCommand{
	{D, ""},
	{D01, ""},
	{D02, ""},
	{D03, ""},
	{G01, ""},
	{G02, ""},
	{G03, ""},
	{G04, ""},
	{G36, ""},
	{G37, ""},
	{G54, ""},
	{G55, ""},
	{G70, ""},
	{G71, ""},
	{G74, ""},
	{G75, ""},
	{G90, ""},
	{G91, ""},
	{M00, ""},
	{M01, ""},
	{M02, ""},
}
var GCmdExtArray = []GerberCommand{
	{AB, ""},
	{AD, ""},
	{AM, ""},
	{AS, ""},
	{FS, ""},
	{IN, ""},
	{IP, ""},
	{IR, ""},
	{LM, ""},
	{LN, ""},
	{LP, ""},
	{LR, ""},
	{LS, ""},
	{MI, ""},
	{MO, ""},
	{OF, ""},
	{SF, ""},
	{SR, ""},
	{TA, ""},
	{TD, ""},
	{TF, ""},
	{TO, ""},
}

var BadGerberCommand = GerberCommand{NOP, ""}

type GerberCommand struct {
	cmd       GerberCommandId
	cmdString string
}

func (gc *GerberCommand) String() string {
	return "{command:\"" + gc.cmd.String() + "\",val:\"" + gc.cmdString + "\"}"
}

func (gc *GerberCommand) Clone() GerberCommand {
	retVal := new(GerberCommand)
	retVal.cmd = gc.cmd
	retVal.cmdString = copys2(gc.cmdString)
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

func SplitByGCommands(buf *[]byte) *[]GerberCommand {
	retVal := make([]GerberCommand, 0)
	ExtBegin := false
	cmdDetected := false
	cc := ""
	cbody := ""
	baseNum := ""
	var AddCmd = GerberCommand{NOP, ""}
	for i := range *buf {
		// purge CR LF
		if (*buf)[i] == 0x0A || (*buf)[i] == 0x0D {
			continue
		}
		if ExtBegin == false && (*buf)[i] == byte(ExtCmdDelimiter) {
			ExtBegin = true
			continue
		}
		if ExtBegin == true && cmdDetected == false {
			cc = cc + string((*buf)[i])
			if len(cc) == 2 {
				// look foe the command
				cmdDetected = true
				for j := range GCmdExtArray {
					if cc == GCmdExtArray[j].cmd.String() {
						AddCmd = GCmdExtArray[j].Clone()
						break
					}
					AddCmd = BadGerberCommand.Clone()
				}
			}
			continue
		}
		if ExtBegin == true && cmdDetected == true {
			if (*buf)[i] == byte(ExtCmdDelimiter) {
				retVal = append(retVal, AddCmd)
				ExtBegin = false
				cmdDetected = false
				AddCmd.cmdString = ""
				AddCmd.cmd = NOP
				cc = ""
			} else {
				AddCmd.cmdString += string((*buf)[i])
			}
			continue
		}
		// base commands
		if cmdDetected == false {
			c := (*buf)[i]
			if c == 'M' || c == 'G' || 	c == 'D' {
				cc = string(c)
				cmdDetected = true
			} else {
				cbody = cbody + string(c)
			}
			continue
		}
		if cmdDetected == true {
			c := (*buf)[i]
			if c == ' ' || c == byte(DataBlockTrailer) {
				cc = FormatGCode(cc, baseNum)
				if cc[0] == 'D' && cc != "D01" && cc != "D02" && cc != "D03" {
					cbody = cc[1:]
					cc = "D"
				}
				for j := range GCmdBaseArray {
					if cc == GCmdBaseArray[j].cmd.String() {
						AddCmd = GCmdBaseArray[j].Clone()
						AddCmd.cmdString = cbody
						break
					}
					AddCmd = BadGerberCommand.Clone()
					AddCmd.cmdString = cc
				}
				retVal = append(retVal, AddCmd)
				cmdDetected = false
				baseNum = ""
				cbody = ""
				cc = ""
			} else {
				baseNum = baseNum + string(c)
			}
			continue
		}
		panic("Dead end reached!")
	}
	return &retVal
}

func copys2(a string) string {
	if len(a) == 0 {
		return ""
	}
	return a[0:1] + a[1:]
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
