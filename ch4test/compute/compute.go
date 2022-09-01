package compute

type Device interface {
	Square([]float32) []float32
	Sum([]float32) float32
}

func SumOfSquares(c Device, in []float32) float32 {
	sq := c.Square(in)
	return c.Sum(sq)
}
