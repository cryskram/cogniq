package indexer

import "strings"

const DefaultChunkSize = 50
const DefaultChunkOverlap = 10

type Chunk struct {
	Index   int
	Content string
}

func ChunkContent(content string, chunkSize, overlap int) []Chunk {
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= chunkSize {
		overlap = chunkSize / 2
	}

	if content == "" {
		return nil
	}

	lines := strings.Split(content, "\n")
	if len(lines) <= chunkSize {
		return []Chunk{{Index: 0, Content: content}}
	}

	var chunks []Chunk
	start := 0
	index := 0
	for start < len(lines) {
		end := start + chunkSize
		if end > len(lines) {
			end = len(lines)
		}
		chunkText := strings.Join(lines[start:end], "\n")
		chunks = append(chunks, Chunk{Index: index, Content: chunkText})
		index++
		start += chunkSize - overlap
		if start >= len(lines) {
			break
		}
	}
	return chunks
}
