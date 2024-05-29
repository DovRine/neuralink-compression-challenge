package main

import (
	"container/heap"
	"fmt"
	"unsafe"
)

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

func main() {
	// Example tuples (frequency, amplitude)
	tuples := [][2]int{{440, 100}, {445, 105}, {450, 110}, {460, 120}, {470, 130}}

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

	// Print results
	fmt.Println("Original:", tuples)
	fmt.Printf("Original size: %d bytes\n", originalSize)
	fmt.Println("Encoded:", encodedTuples)
	fmt.Println("Huffman Encoded String Length:", len(encodedString))
	fmt.Printf("Packed size: %d bytes (Compression ratio: %.2f)\n", packedSize, compressionRatio)
	fmt.Println("Unpacked:", unpackedData)
}
