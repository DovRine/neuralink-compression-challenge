package main

import (
	"fmt"
	"unsafe"
)

// Common values and their short representations
var commonValues = map[int]byte{
	10:    1,
	100:   2,
	1000:  3,
	10000: 4,
}

// Reverse mapping for decompression
var reverseCommonValues = map[byte]int{
	1: 10,
	2: 100,
	3: 1000,
	4: 10000,
}

// PackTuple compresses a tuple (element1, element2) into a single uint32 value.
func PackTuple(element1, element2 int) uint32 {
	return (uint32(element1) << 11) | uint32(element2)
}

// UnpackTuple decompresses a packed uint32 value into a tuple (element1, element2).
func UnpackTuple(packedValue uint32) (int, int) {
	element1 := int(packedValue >> 11)
	element2 := int(packedValue & 0x7FF) // 0x7FF is 11 bits of 1s
	return element1, element2
}

// DeltaEncode converts a list of tuples to a list of delta-encoded tuples.
func DeltaEncode(tuples [][2]int) [][2]int {
	if len(tuples) == 0 {
		return nil
	}
	encoded := make([][2]int, len(tuples))
	encoded[0] = tuples[0]
	for i := 1; i < len(tuples); i++ {
		encoded[i][0] = tuples[i][0] - tuples[i-1][0]
		encoded[i][1] = tuples[i][1] - tuples[i-1][1]
	}
	return encoded
}

// DeltaDecode converts a list of delta-encoded tuples back to the original list of tuples.
func DeltaDecode(encoded [][2]int) [][2]int {
	if len(encoded) == 0 {
		return nil
	}
	tuples := make([][2]int, len(encoded))
	tuples[0] = encoded[0]
	for i := 1; i < len(encoded); i++ {
		tuples[i][0] = encoded[i][0] + tuples[i-1][0]
		tuples[i][1] = encoded[i][1] + tuples[i-1][1]
	}
	return tuples
}

// CompressElement compresses an element using the replacement scheme.
func CompressElement(element int) (byte, bool) {
	if val, exists := commonValues[element]; exists {
		return val, true
	}
	return 0, false
}

// DecompressElement decompresses an element using the replacement scheme.
func DecompressElement(element byte) (int, bool) {
	if val, exists := reverseCommonValues[element]; exists {
		return val, true
	}
	return 0, false
}

func main() {
	// Example tuples (frequency, amplitude)
	tuples := [][2]int{{440, 100}, {445, 10000}, {450, 10}, {460, 120}, {470, 130}}

	// Calculate original size
	originalSize := len(tuples) * int(unsafe.Sizeof(tuples[0]))

	// Delta encode tuples
	encodedTuples := DeltaEncode(tuples)

	// Apply replacement scheme and pack tuples
	var packedData []uint32
	var replacements []byte
	for _, tuple := range encodedTuples {
		e1, replaced1 := CompressElement(tuple[0])
		e2, replaced2 := CompressElement(tuple[1])
		if replaced1 && replaced2 {
			replacements = append(replacements, e1, e2)
		} else {
			packed := PackTuple(tuple[0], tuple[1])
			packedData = append(packedData, packed)
		}
	}

	// Calculate packed size
	packedSize := len(packedData)*int(unsafe.Sizeof(packedData[0])) + len(replacements)

	// Calculate compression ratio
	compressionRatio := float64(originalSize) / float64(packedSize)

	// Unpack tuples and apply reverse replacement scheme
	var unpackedEncodedData [][2]int
	for _, packed := range packedData {
		element1, element2 := UnpackTuple(packed)
		unpackedEncodedData = append(unpackedEncodedData, [2]int{element1, element2})
	}

	// Reverse replacement scheme
	for i := 0; i < len(replacements); i += 2 {
		e1, _ := DecompressElement(replacements[i])
		e2, _ := DecompressElement(replacements[i+1])
		unpackedEncodedData = append(unpackedEncodedData, [2]int{e1, e2})
	}

	// Delta decode tuples
	unpackedData := DeltaDecode(unpackedEncodedData)

	// Print results
	fmt.Println("Original:", tuples)
	fmt.Printf("Original size: %d bytes\n", originalSize)
	fmt.Println("Encoded:", encodedTuples)
	fmt.Println("Packed:", packedData)
	fmt.Println("Replacements:", replacements)
	fmt.Printf("Packed size: %d bytes (Compression ratio: %.2f)\n", packedSize, compressionRatio)
	fmt.Println("Unpacked:", unpackedData)
}
