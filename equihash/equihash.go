package equihash

import (
	"errors"
)

var hexDigits = [256]int8{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, -1, -1, -1, -1, -1, -1,
	-1, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}

func hexDigit(b byte) int8 {
	return hexDigits[int(b)]
}

var bitLenErr = errors.New("bit len must be >= 8")

func ExpandArray(in, out []byte, bitLen, bytePad uint32) error {
	// prepare state for expansions
	if bitLen < 8 {
		return bitLenErr
	}
	outWidth := (bitLen+7)/8 + bytePad
	bitLenMask := uint32(1)<<bytePad - uint32(1)
	accBits, accVal, j := uint32(0), uint32(0), 0

	// loop to start expansion
	for _, val := range in {
		accVal = (accVal << 8) | uint32(val)
		accBits += 8

		if accBits >= uint32(bitLen) {
			accBits -= bitLen
			for x := 0; uint32(x) < bytePad; x++ {
				out[j+x] = 0
			}
			for x := bytePad; x < outWidth; x++ {
				topMask := (accVal >> (accBits + (8 * (uint32(len(out)) - x - 1))))
				botMask := (bitLenMask >> (8 * (outWidth - x - 1))) & 0xFF
				out[j+int(x)] = byte(topMask & botMask)
			}
			j += int(outWidth)
		}
	}

	return nil
}

func ParseHex(s string) ([]byte, error) {
	out := []byte{}
	for i, c := range s {
		if 
	}
}
