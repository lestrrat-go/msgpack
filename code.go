package msgpack

func (t Code) Byte() byte {
	return byte(t)
}

func IsMapFamily(c Code) bool {
	b := c.Byte()
	return (b >= FixMap0.Byte() && b <= FixMap15.Byte()) ||
		b == Map16.Byte() ||
		b == Map32.Byte()
}
