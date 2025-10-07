package common

type Fixed int32

func FloatToFixed(f float64) Fixed {
	return Fixed(f * 65536)
}

func (f *Fixed) FixedToFloat() float64 {
	return float64(*f) / 65536
}
