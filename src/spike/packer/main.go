package main

import (
	"bytes"
	"container/heap"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"unsafe"
)

// Huffman coding structures and functions
type HuffmanNode struct {
	value    int
	freq     int
	left     *HuffmanNode
	right    *HuffmanNode
	code     string
	index    int
}

type HuffmanHeap []*HuffmanNode

func (h HuffmanHeap) Len() int           { return len(h) }
func (h HuffmanHeap) Less(i, j int) bool { return h[i].freq < h[j].freq }
func (h HuffmanHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }
func (h *HuffmanHeap) Push(x interface{}) {
	n := len(*h)
	node := x.(*HuffmanNode)
	node.index = n
	*h = append(*h, node)
}
func (h *HuffmanHeap) Pop() interface{} {
	old := *h
	n := len(old)
	node := old[n-1]
	old[n-1] = nil
	node.index = -1
	*h = old[0 : n-1]
	return node
}

func buildHuffmanTree(freqMap map[int]int) *HuffmanNode {
	h := &HuffmanHeap{}
	heap.Init(h)
	for value, freq := range freqMap {
		heap.Push(h, &HuffmanNode{value: value, freq: freq})
	}
	for h.Len() > 1 {
		left := heap.Pop(h).(*HuffmanNode)
		right := heap.Pop(h).(*HuffmanNode)
		newNode := &HuffmanNode{
			freq:  left.freq + right.freq,
			left:  left,
			right: right,
		}
		heap.Push(h, newNode)
	}
	return heap.Pop(h).(*HuffmanNode)
}

func generateHuffmanCodes(node *HuffmanNode, code string, codeMap map[int]string) {
	if node == nil {
		return
	}
	if node.left == nil && node.right == nil {
		node.code = code
		codeMap[node.value] = code
	}
	generateHuffmanCodes(node.left, code+"0", codeMap)
	generateHuffmanCodes(node.right, code+"1", codeMap)
}

func encodeData(data []int, codeMap map[int]string) string {
	encoded := ""
	for _, value := range data {
		encoded += codeMap[value]
	}
	return encoded
}

func decodeData(encoded string, root *HuffmanNode) []int {
	node := root
	decoded := []int{}
	for _, bit := range encoded {
		if bit == '0' {
			node = node.left
		} else {
			node = node.right
		}
		if node.left == nil && node.right == nil {
			decoded = append(decoded, node.value)
			node = root
		}
	}
	return decoded
}

// Delta encoding functions
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

// WAV file parsing functions
func readWAVFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data []byte
	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}
		data = append(data, buffer[:n]...)
	}
	return data, nil
}

func parseWAVHeader(data []byte) (uint16, uint16, uint32, uint32, []byte, error) {
	reader := bytes.NewReader(data)

	var chunkID [4]byte
	var chunkSize uint32
	var format [4]byte

	var subchunk1ID [4]byte
	var subchunk1Size uint32
	var audioFormat uint16
	var numChannels uint16
	var sampleRate uint32
	var byteRate uint32
	var blockAlign uint16
	var bitsPerSample uint16

	var subchunk2ID [4]byte
	var subchunk2Size uint32

	if err := binary.Read(reader, binary.LittleEndian, &chunkID); err != nil {
		return 0, 0, 0, 0, nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &chunkSize); err != nil {
		return 0, 0, 0, 0, nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &format); err != nil {
		return 0, 0, 0, 0, nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &subchunk1ID); err != nil {
		return 0, 0, 0, 0, nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &subchunk1Size); err != nil {
		return 0, 0, 0, 0, nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &audioFormat); err != nil {
		return 0, 0, 0, 0, nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &numChannels); err != nil {
		return 0, 0, 0, 0, nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &sampleRate); err != nil {
		return 0, 0, 0, 0, nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &byteRate); err != nil {
		return 0, 0, 0, 0, nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &blockAlign); err != nil {
		return 0, 0, 0, 0, nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &bitsPerSample); err != nil {
		return 0, 0, 0, 0, nil, err
	}

	for {
		if err := binary.Read(reader, binary.LittleEndian, &subchunk2ID); err != nil {
			return 0, 0, 0, 0, nil, err
		}
		if err := binary.Read(reader, binary.LittleEndian, &subchunk2Size); err != nil {
			return 0, 0, 0, 0, nil, err
		}

		if string(subchunk2ID[:4]) == "data" {
			break
		} else {
			// Skip unknown chunk
			if _, err := reader.Seek(int64(subchunk2Size), 1); err != nil {
				return 0, 0, 0, 0, nil, err
			}
		}
	}

	audioData := make([]byte, subchunk2Size)
	if err := binary.Read(reader, binary.LittleEndian, &audioData); err != nil {
		return 0, 0, 0, 0, nil, err
	}

	return numChannels, bitsPerSample, sampleRate, subchunk2Size, audioData, nil
}

func convertAudioToTuples(numChannels uint16, bitsPerSample uint16, audioData []byte) [][2]int {
	var tuples [][2]int

	sampleSize := int(bitsPerSample / 8)
	numSamples := len(audioData) / (int(numChannels) * sampleSize)

	for i := 0; i < numSamples; i++ {
		for ch := 0; ch < int(numChannels); ch++ {
			offset := (i*int(numChannels) + ch) * sampleSize
			var value int

			switch sampleSize {
			case 1:
				value = int(audioData[offset])
			case 2:
				value = int(int16(binary.LittleEndian.Uint16(audioData[offset:])))
			case 4:
				value = int(int32(binary.LittleEndian.Uint32(audioData[offset:])))
			}

			// Assuming first channel is frequency and second channel is amplitude
			if ch == 0 {
				tuple := [2]int{value, 0}
				if len(tuples) > 0 {
					tuple[1] = tuples[len(tuples)-1][1]
				}
				tuples = append(tuples, tuple)
			} else if ch == 1 {
				tuples[len(tuples)-1][1] = value
			}
		}
	}

	return tuples
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go path/to/wavfile")
		return
	}

	filePath := os.Args[1]

	// Read the WAV file
	data, err := readWAVFile(filePath)
	if err != nil {
		fmt.Println("Error reading WAV file:", err)
		return
	}

	// Parse the WAV header and fmt chunk
	numChannels, bitsPerSample, sampleRate, dataSize, audioData, err := parseWAVHeader(data)
	if err != nil {
		fmt.Println("Error parsing WAV file:", err)
		return
	}

	fmt.Printf("Number of Channels: %d\n", numChannels)
	fmt.Printf("Bits per Sample: %d\n", bitsPerSample)
	fmt.Printf("Sample Rate: %d\n", sampleRate)
	fmt.Printf("Data Size: %d bytes\n", dataSize)

	// Convert audio data to list of tuples (frequency, amplitude)
	tuples := convertAudioToTuples(numChannels, bitsPerSample, audioData)

	// Calculate original size
	originalSize := len(tuples) * int(unsafe.Sizeof(tuples[0]))

	// Delta encode tuples
	encodedTuples := DeltaEncode(tuples)

	// Flatten the tuples for Huffman encoding
	var flattenedData []int
	for _, tuple := range encodedTuples {
		flattenedData = append(flattenedData, tuple[0], tuple[1])
	}

	// Calculate frequency map
	freqMap := make(map[int]int)
	for _, value := range flattenedData {
		freqMap[value]++
	}

	// Build Huffman tree
	huffmanRoot := buildHuffmanTree(freqMap)

	// Generate Huffman codes
	codeMap := make(map[int]string)
	generateHuffmanCodes(huffmanRoot, "", codeMap)

	// Encode data using Huffman codes
	encodedString := encodeData(flattenedData, codeMap)

	// Calculate packed size in bytes
	packedSize := (len(encodedString) + 7) / 8 // bits to bytes

	// Calculate compression ratio
	compressionRatio := float64(originalSize) / float64(packedSize)

	// Decode data back to original form
	decodedFlattenedData := decodeData(encodedString, huffmanRoot)

	// Reconstruct tuples from flattened data
	var unpackedEncodedData [][2]int
	for i := 0; i < len(decodedFlattenedData); i += 2 {
		unpackedEncodedData = append(unpackedEncodedData, [2]int{decodedFlattenedData[i], decodedFlattenedData[i+1]})
	}

	// Delta decode tuples
	unpackedData := DeltaDecode(unpackedEncodedData)
	_ = unpackedData

	// Print results
	// fmt.Println("Original:", tuples)
	fmt.Printf("Original size: %d bytes\n", originalSize)
	// fmt.Println("Encoded:", encodedTuples)
	fmt.Println("Huffman Encoded String Length:", len(encodedString))
	fmt.Printf("Packed size: %d bytes (Compression ratio: %.2f)\n", packedSize, compressionRatio)
	// fmt.Println("Unpacked:", unpackedData)
}
