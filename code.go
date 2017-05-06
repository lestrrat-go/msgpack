package msgpack

// Byte returns the byte representation of the Code
func (t Code) Byte() byte {
	return byte(t)
}

// IsMapFamily returns true if the given code is equivalent to
// one of the `map` family in msgpack
func IsMapFamily(c Code) bool {
	b := c.Byte()
	return (b >= FixMap0.Byte() && b <= FixMap15.Byte()) ||
		b == Map16.Byte() ||
		b == Map32.Byte()
}
