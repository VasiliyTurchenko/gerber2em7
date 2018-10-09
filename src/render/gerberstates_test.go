package render

import (
	"gerberbasetypes"
	"testing"
)

func TestApTransParameters_String(t *testing.T) {
	apt := ApTransParameters{Polarity: gerberbasetypes.PolTypeClear,
		Mirroring: gerberbasetypes.MirrorXY,
		Rotation:  98.0,
		Scale:     0.01}
	t.Log(apt.String())
}
