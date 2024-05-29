package main

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
