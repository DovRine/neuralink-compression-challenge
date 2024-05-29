package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

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
	tuples := make([][2]int, len(audioData)/(int(numChannels)*(int(bitsPerSample)/8)))

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
				tuples[i][0] = value
			} else if ch == 1 {
				tuples[i][1] = value
			}
		}
	}

	return tuples
}

func writeWAVFile(filePath string, numChannels uint16, bitsPerSample uint16, sampleRate uint32, audioData []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write RIFF header
	if err := binary.Write(writer, binary.LittleEndian, [4]byte{'R', 'I', 'F', 'F'}); err != nil {
		return err
	}
	chunkSize := uint32(36 + len(audioData))
	if err := binary.Write(writer, binary.LittleEndian, chunkSize); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.LittleEndian, [4]byte{'W', 'A', 'V', 'E'}); err != nil {
		return err
	}

	// Write fmt subchunk
	if err := binary.Write(writer, binary.LittleEndian, [4]byte{'f', 'm', 't', ' '}); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.LittleEndian, uint32(16)); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.LittleEndian, uint16(1)); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.LittleEndian, numChannels); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.LittleEndian, sampleRate); err != nil {
		return err
	}
	byteRate := sampleRate * uint32(numChannels) * uint32(bitsPerSample) / 8
	if err := binary.Write(writer, binary.LittleEndian, byteRate); err != nil {
		return err
	}
	blockAlign := numChannels * bitsPerSample / 8
	if err := binary.Write(writer, binary.LittleEndian, blockAlign); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.LittleEndian, bitsPerSample); err != nil {
		return err
	}

	// Write data subchunk
	if err := binary.Write(writer, binary.LittleEndian, [4]byte{'d', 'a', 't', 'a'}); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.LittleEndian, uint32(len(audioData))); err != nil {
		return err
	}
	if _, err := writer.Write(audioData); err != nil {
		return err
	}

	return writer.Flush()
}

func writeCompressedFile(file *os.File, numChannels uint16, bitsPerSample uint16, sampleRate uint32, dataSize uint32, encodedData []byte) error {
	writer := bufio.NewWriter(file)
	if err := binary.Write(writer, binary.LittleEndian, numChannels); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.LittleEndian, bitsPerSample); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.LittleEndian, sampleRate); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.LittleEndian, dataSize); err != nil {
		return err
	}
	if _, err := writer.Write(encodedData); err != nil {
		return err
	}
	return writer.Flush()
}

func readCompressedFile(filePath string) (uint16, uint16, uint32, uint32, []byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, 0, 0, nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	var numChannels uint16
	var bitsPerSample uint16
	var sampleRate uint32
	var dataSize uint32

	if err := binary.Read(reader, binary.LittleEndian, &numChannels); err != nil {
		return 0, 0, 0, 0, nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &bitsPerSample); err != nil {
		return 0, 0, 0, 0, nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &sampleRate); err != nil {
		return 0, 0, 0, 0, nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &dataSize); err != nil {
		return 0, 0, 0, 0, nil, err
	}

	compressedData := make([]byte, dataSize)
	if _, err := io.ReadFull(reader, compressedData); err != nil {
		return 0, 0, 0, 0, nil, err
	}

	return numChannels, bitsPerSample, sampleRate, dataSize, compressedData, nil
}
