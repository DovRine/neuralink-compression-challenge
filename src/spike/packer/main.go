package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run main.go [encode|decode] path/to/wavfile path/to/outputfile")
		return
	}

	mode := os.Args[1]
	inputFilePath := os.Args[2]
	outputFilePath := os.Args[3]

	switch mode {
	case "encode":
		encode(inputFilePath, outputFilePath)
	case "decode":
		decode(inputFilePath, outputFilePath)
	default:
		fmt.Println("Invalid mode. Use 'encode' or 'decode'.")
	}
}

func encode(inputFilePath, outputFilePath string) {
	// Read the WAV file
	data, err := readWAVFile(inputFilePath)
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

	// Delta encode tuples
	encodedTuples := DeltaEncode(tuples)

	// Flatten the tuples for Huffman encoding
	flattenedData := make([]int, len(encodedTuples)*2)
	for i, tuple := range encodedTuples {
		flattenedData[i*2] = tuple[0]
		flattenedData[i*2+1] = tuple[1]
	}

	// Calculate frequency map
	freqMap := make(map[int]int)
	for _, value := range flattenedData {
		freqMap[value]++
	}

	// Build Huffman tree
	huffmanRoot := BuildHuffmanTree(freqMap)

	// Generate Huffman codes
	codeMap := make(map[int]*HuffmanNode)
	GenerateHuffmanCodes(huffmanRoot, 0, 0, codeMap)

	// Measure encoding time
	startEncoding := time.Now()
	// Encode data using Huffman codes
	encodedData := EncodeData(flattenedData, codeMap)
	encodingTime := time.Since(startEncoding)

	// Calculate packed size in bytes
	packedSize := len(encodedData)

	// Create intermediate compressed file
	compressedFilePath := inputFilePath[:len(inputFilePath)-len(filepath.Ext(inputFilePath))] + ".compressed"
	compressedFile, err := os.Create(compressedFilePath)
	if err != nil {
		fmt.Println("Error creating compressed file:", err)
		return
	}
	defer compressedFile.Close()

	// Write headers and compressed data to intermediate file
	if err := writeCompressedFile(compressedFile, numChannels, bitsPerSample, sampleRate, uint32(len(encodedData)), encodedData); err != nil {
		fmt.Println("Error writing compressed file:", err)
		return
	}

	// Calculate compression ratio
	originalSize := len(audioData)
	compressionRatio := float64(originalSize) / float64(packedSize)

	// Print results
	fmt.Printf("Original size: %d bytes\n", originalSize)
	fmt.Printf("Packed size: %d bytes (Compression ratio: %.2f)\n", packedSize, compressionRatio)
	fmt.Printf("Encoding time: %s\n", encodingTime)
}

func decode(inputFilePath, outputFilePath string) {
	// Read the compressed file
	numChannels, bitsPerSample, sampleRate, dataSize, compressedData, err := readCompressedFile(inputFilePath)
	if err != nil {
		fmt.Println("Error reading compressed file:", err)
		return
	}

	// Calculate frequency map
	freqMap := make(map[int]int)
	for _, value := range compressedData {
		freqMap[int(value)]++
	}

	// Build Huffman tree
	huffmanRoot := BuildHuffmanTree(freqMap)

	// Measure decoding time
	startDecoding := time.Now()
	// Decode data back to original form
	decodedFlattenedData := DecodeData(compressedData, huffmanRoot, int(dataSize)*8)
	decodingTime := time.Since(startDecoding)

	// Reconstruct tuples from flattened data
	unpackedEncodedData := make([][2]int, len(decodedFlattenedData)/2)
	for i := range unpackedEncodedData {
		unpackedEncodedData[i][0] = decodedFlattenedData[i*2]
		unpackedEncodedData[i][1] = decodedFlattenedData[i*2+1]
	}

	// Delta decode tuples
	unpackedData := DeltaDecode(unpackedEncodedData)

	// Convert unpacked tuples back to byte array
	reconstructedAudioData := make([]byte, len(unpackedData)*int(bitsPerSample/8)*int(numChannels))
	sampleSize := int(bitsPerSample / 8)
	for i, tuple := range unpackedData {
		for ch := 0; ch < int(numChannels); ch++ {
			offset := (i*int(numChannels) + ch) * sampleSize
			switch sampleSize {
			case 1:
				reconstructedAudioData[offset] = byte(tuple[ch])
			case 2:
				binary.LittleEndian.PutUint16(reconstructedAudioData[offset:], uint16(tuple[ch]))
			case 4:
				binary.LittleEndian.PutUint32(reconstructedAudioData[offset:], uint32(tuple[ch]))
			}
		}
	}

	// Write unpacked data to the output file
	if err := writeWAVFile(outputFilePath, numChannels, bitsPerSample, sampleRate, reconstructedAudioData); err != nil {
		fmt.Println("Error writing unpacked data:", err)
		return
	}

	// Print results
	fmt.Printf("Decoding time: %s\n", decodingTime)
}
