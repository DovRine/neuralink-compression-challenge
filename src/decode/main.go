package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

func indexedRunLengthDecode(filename string, outputPath string) ([]byte, error) {
	encodedData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var decodedData []byte
	for i := 0; i < len(encodedData); i += 9 {
		startIndex := int(binary.BigEndian.Uint32(encodedData[i : i+4]))
		count := int(binary.BigEndian.Uint32(encodedData[i+4 : i+8]))
		value := encodedData[i+8]

		for j := 0; j < count; j++ {
			if startIndex+j >= len(decodedData) {
				tmp := make([]byte, startIndex+j+1)
				copy(tmp, decodedData)
				decodedData = tmp
			}
			decodedData[startIndex+j] = value
		}
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, err
	}
	defer outputFile.Close()

	_, err = outputFile.Write(decodedData)
	if err != nil {
		return nil, err
	}

	return decodedData, nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: decode <encoded_filename> <output_path>")
		os.Exit(1)
	}

	filename := os.Args[1]
	outputPath := os.Args[2]

	_, err := indexedRunLengthDecode(filename, outputPath)
	if err != nil {
		fmt.Println("Error decoding audio:", err)
		os.Exit(1)
	}

	fmt.Println("Decoded audio saved to:", outputPath)
}
