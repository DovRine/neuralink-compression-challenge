package main

import (
	"container/heap"
)

// HuffmanNode represents a node in the Huffman tree
type HuffmanNode struct {
	value    int
	freq     int
	left     *HuffmanNode
	right    *HuffmanNode
	code     uint64
	codeLen  uint8
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

// BuildHuffmanTree builds a Huffman tree from a frequency map
func BuildHuffmanTree(freqMap map[int]int) *HuffmanNode {
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

// GenerateHuffmanCodes generates Huffman codes for all values in the Huffman tree
func GenerateHuffmanCodes(node *HuffmanNode, code uint64, codeLen uint8, codeMap map[int]*HuffmanNode) {
	if node == nil {
		return
	}
	if node.left == nil && node.right == nil {
		node.code = code
		node.codeLen = codeLen
		codeMap[node.value] = node
	}
	GenerateHuffmanCodes(node.left, code<<1, codeLen+1, codeMap)
	GenerateHuffmanCodes(node.right, (code<<1)|1, codeLen+1, codeMap)
}

// EncodeData encodes the given data using Huffman codes
func EncodeData(data []int, codeMap map[int]*HuffmanNode) []byte {
	buf := make([]byte, (len(data)*8+7)/8)
	var bitPos uint8
	var bytePos int
	for _, value := range data {
		node := codeMap[value]
		for i := node.codeLen; i > 0; i-- {
			if (node.code>>(i-1))&1 == 1 {
				buf[bytePos] |= 1 << (7 - bitPos)
			}
			bitPos++
			if bitPos == 8 {
				bitPos = 0
				bytePos++
			}
		}
	}
	return buf
}

// DecodeData decodes the given encoded data using the Huffman tree
func DecodeData(encoded []byte, root *HuffmanNode, totalBits int) []int {
	node := root
	decoded := make([]int, 0, totalBits/8)
	bitCount := 0
	for _, b := range encoded {
		for i := 0; i < 8; i++ {
			if bitCount >= totalBits {
				return decoded
			}
			if (b>>(7-i))&1 == 0 {
				node = node.left
			} else {
				node = node.right
			}
			if node.left == nil && node.right == nil {
				decoded = append(decoded, node.value)
				node = root
			}
			bitCount++
		}
	}
	return decoded
}
