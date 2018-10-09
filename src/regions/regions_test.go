package regions

import "testing"

func TestRegion_IsRegionOpened(t *testing.T) {
	regPtr := NewRegion(100)
	regPtr.G37StringNumber = 0
	a, err := regPtr.IsRegionOpened()
	if err != nil {
		t.Fatal("unexpected error")
	}
	if a == true {
		t.Fatal("region is not opened")
	}
	t.Log("all OK")

	regPtr = nil
	a, err = regPtr.IsRegionOpened()
	if err == nil {
		t.Fatal("must be an error")
	}
	if a == true {
		t.Fatal("region is not opened")
	}
	t.Log("all OK")

}
