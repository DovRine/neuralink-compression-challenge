package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

func readAudioFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	audioData := make([]byte, info.Size())
	_, err = file.Read(audioData)
	if err != nil {
		return nil, err
	}

	return audioData, nil
}

func indexedRunLengthEncode(data []byte, outputFilePath string) error {
	encodedData := []byte{}
	i := 0

	for i < len(data) {
		startIndex := i
		count := 0
		for i < len(data) && data[i] == data[startIndex] {
			count++
			i++
		}
		startIndexBytes := make([]byte, 4)
		countBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(startIndexBytes, uint32(startIndex))
		binary.BigEndian.PutUint32(countBytes, uint32(count))
		encodedData = append(encodedData, startIndexBytes...)
		encodedData = append(encodedData, countBytes...)
		encodedData = append(encodedData, data[startIndex])
	}

	return os.WriteFile(outputFilePath, encodedData, 0644)
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: encode <input_filename> <output_filename>")
		os.Exit(1)
	}

	inputFilename := os.Args[1]
	outputFilename := os.Args[2]

	audioData, err := readAudioFile(inputFilename)
	if err != nil {
		fmt.Println("Error reading audio file:", err)
		os.Exit(1)
	}

	err = indexedRunLengthEncode(audioData, outputFilename)
	if err != nil {
		fmt.Println("Error encoding audio:", err)
		os.Exit(1)
	}
}
