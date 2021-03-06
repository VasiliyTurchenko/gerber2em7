// Base types for Gerber parsing and processing
//  go:generate stringer -type=GerberApType
package gerberbasetypes

const (
	MaxInt = int(^uint(0) >> 1)
	MinInt = int(-MaxInt - 1)
)

// Apertures
const GerberApertureDef = "%ADD"
const GerberApertureMacroDef = "%AM"
const GerberApertureBlockDef = "%AB"
const GerberApertureBlockDefEnd = "%AB*%"

type GerberApType int

const (
	AptypeCircle GerberApType = iota + 1
	AptypeRectangle
	AptypeObround
	AptypePoly
	AptypeMacro
	AptypeBlock
)

func (ga GerberApType) String() string {
	switch ga {
	case AptypeCircle:
		return "circle aperture"
	case AptypeRectangle:
		return "rectangle aperture"
	case AptypeObround:
		return "obround (box) aperture"
	case AptypePoly:
		return "polygon aperture"
	case AptypeMacro:
		return "macro aperture"
	case AptypeBlock:
		return "block aperture"
	default:
	}
	return "Unknown aperture type"

}

type PolType int

const (
	PolTypeDark PolType = iota + 1
	PolTypeClear
)

func (p PolType) String() string {
	switch p {
	case PolTypeDark:
		return "Polarity: dark"
	case PolTypeClear:
		return "Polarity: clear"
	default:

	}
	return "Unknown polarity"
}

type ActType int

const (
	OpcodeD01_DRAW ActType = iota + 1
	OpcodeD02_MOVE
	OpcodeD03_FLASH
	OpcodeStop
)

func (act ActType) String() string {
	switch act {
	case OpcodeD01_DRAW:
		return "Opcode D01 (DRAW)"
	case OpcodeD02_MOVE:
		return "Opcode D02 (MOVE)"
	case OpcodeD03_FLASH:
		return "Opcode D03 (FLASH)"
	case OpcodeStop:
		return "Opcode Stop"
	default:

	}
	return "Unknown Opcode"
}

type QuadMode int

const (
	QuadModeSingle QuadMode = iota + 1
	QuadModeMulti
)

func (q QuadMode) String() string {
	switch q {
	case QuadModeSingle:
		return "QuadMode: Single"
	case QuadModeMulti:
		return "QuadMode: Multi"
	default:

	}
	return "Unknown QuadMode"
}

type IPmode int

const (
	IPModeLinear IPmode = iota + 1
	IPModeCwC
	IPModeCCwC
)

func (ipm IPmode) String() string {
	switch ipm {
	case IPModeLinear:
		return "Linear interpolation"
	case IPModeCwC:
		return "Clockwise interpolation"
	case IPModeCCwC:
		return "Counter-clockwise interpolation"
	default:

	}
	return "Unknown interpolation"
}

type Mirror int

const (
	NoMirror Mirror = iota + 1
	MirrorX
	MirrorY
	MirrorXY
)

func (m Mirror) String() string {
	switch m {
	case NoMirror:
		return "No mirroring"
	case MirrorX:
		return "Mirroring X"
	case MirrorY:
		return "Mirroring Y"
	case MirrorXY:
		return "Mirroring X,Y"
	default:

	}
	return "Unknown mirroring type"
}
