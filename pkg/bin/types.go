package bin

// TODO: integration uint24 to struct parser

func Uint24(u24 [3]uint8) uint {
	i := uint(u24[2]) | (uint(u24[1]) << 8) | (uint(u24[0]) << 16)
	if (i & 0x800000) > 0 {
		i |= 0xFF000000
	}
	return i
}
