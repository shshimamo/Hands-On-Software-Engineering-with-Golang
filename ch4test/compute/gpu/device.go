package gpu

type Device struct{}

func NewDevice() *Device {
	return new(Device)
}

func (*Device) Square(in []float32) []float32 {
	panic("not implemented")
}

func (*Device) Sum(in []float32) float32 {
	panic("not implemented")
}
