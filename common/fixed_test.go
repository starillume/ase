package common

import "testing"

func TestFloatToFixedAndBack(t *testing.T) {
	values := []float64{0, 1, -1, 3.14159, -2.71828, 123.456}

	for _, v := range values {
		fixed := FloatToFixed(v)
		back := fixed.FixedToFloat()

		diff := v - back
		if diff < 0 {
			diff = -diff
		}

		if diff > 1e-5 {
			t.Errorf("conversion mismatch for %f: got %f (fixed=%d)", v, back, fixed)
		}
	}
}

func TestFloatToFixed_RoundTripZero(t *testing.T) {
	fixed := FloatToFixed(0)
	if fixed != 0 {
		t.Errorf("expected 0, got %d", fixed)
	}

	if val := fixed.FixedToFloat(); val != 0 {
		t.Errorf("expected 0.0, got %f", val)
	}
}
