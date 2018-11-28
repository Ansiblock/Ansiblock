package utils

// TODO check on littleendian system
// misleading name, better ByteToUint64
func ByteToInt64(b []byte, st int) int64 {
	return int64(uint64(b[st+7]) | uint64(b[st+6])<<8 | uint64(b[st+5])<<16 | uint64(b[st+4])<<24 |
		uint64(b[st+3])<<32 | uint64(b[st+2])<<40 | uint64(b[st+1])<<48 | uint64(b[st])<<56)
}

func Uint64toByte(v uint64) []byte {
	result := make([]byte, 8)
	for i := 0; i < 8; i++ {
		shiftCount := uint(8 * (7 - i))
		result[i] = byte(v >> shiftCount)
	}
	return result
}
