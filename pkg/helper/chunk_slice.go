package helper

func ChunkSlice[K interface{}](slice []K, chunkSize int) [][]K {
	var chunks [][]K
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}
