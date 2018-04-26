package equihash

import (
	"errors"
	"unicode"
)

var (
	errLen      = errors.New("slices not same len")
	errBitLen   = errors.New("bit len < 8")
	errOutWidth = errors.New("incorrect outwidth size")
	hexDigits   = [256]int8{
		1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
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
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	}
)

const (
	wordSize = 32
	wordMask = (1 << wordSize) - 1
)

func hexDigit(r rune) int8 {
	return -hexDigits[int(r)]
}

func parseHex(s string) []byte {
	out := []byte{}
	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}

		c := hexDigit(r)
		n := c << 4
		if c == -1 {
			break
		}
		n = n | c
		out = append(out, byte(n))
	}
	return out
}

func xor(a, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, errLen
	}
	out := make([]byte, len(a))
	for i, val := range a {
		out[i] = val ^ b[i]
	}
	return out, nil
}

func expandArray(in []byte, outLen, bitLen, bytePad int) ([]byte, error) {
	if bitLen < 8 {
		return nil, errBitLen
	}
	if wordSize < 7+bitLen {
		return nil, errBitLen
	}
	outWidth := (bitLen+7)/8 + bytePad
	if outLen != 8*outWidth*len(in)/bitLen {
		return nil, errOutWidth
	}
	bitLenMask := (1 << uint(bitLen)) - 1
	accBits, accVal, j := 0, 0, 0
	out := make([]byte, outLen)
	for i := 0; i < len(in); i++ {
		accVal = ((accVal << 8) & wordMask) | int(in[i])
		accBits += 8

		if accBits >= bitLen {
			accBits -= bitLen
			for x := bytePad; x < outWidth; x++ {
				a := accVal >> (uint(accBits + (8 * (outWidth - x - 1))))
				b := (bitLenMask >> uint((8 * (outWidth - x - 1)))) & 0xFF
				out[j+x] = byte(a) & byte(b)
			}
			j += outWidth
		}
	}

	return out, nil
}

func compressArray(in []byte, outLen, bitLen, bytePad int) ([]byte, error) {
	if bitLen < 8 {
		return nil, errBitLen
	}
	if wordSize < 7+bitLen {
		return nil, errBitLen
	}

	inWidth := (bitLen+7)/8 + bytePad
	if outLen != 8*inWidth*len(in)/bitLen {
		return nil, errOutWidth
	}
	bitLenMask := (1 << uint(bitLen)) - 1
	accBits, accVal, j := 0, 0, 0
	out := make([]byte, outLen)
	for i := 0; i < outLen; i++ {
		if accBits < 8 {
			accVal = accVal << uint(bitLen)
			for x := bytePad; x < inWidth; x++ {
				a := in[j+x]
				b := (bitLenMask >> uint(8*(inWidth-x-1))) & 0xFF
				mask := a & byte(b)
				accVal = accVal | int(mask)
			}
			j += inWidth
			accBits += bitLen
		}

		accBits -= 8
		out[i] = byte((accVal >> uint(accBits)) & 0xFF)
	}
	return out, nil
}
