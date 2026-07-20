package chunker

import "strings"

type Symbol struct {
	Name string
	Kind string
	Line int
	Col  int
}

type Chunk struct {
	Content   string
	StartLine int
	EndLine   int
	Index     int
	Symbols   []Symbol
}

type Chunker func(content string) []Chunk

var registry = map[string]Chunker{
	"Go":          GoChunker,
	"Python":      PythonChunker,
	"JavaScript":  BraceChunker,
	"TypeScript":  BraceChunker,
	"Rust":        BraceChunker,
	"Java":        BraceChunker,
	"Kotlin":      BraceChunker,
	"C":           BraceChunker,
	"C++":         BraceChunker,
	"C#":          BraceChunker,
	"Objective-C": BraceChunker,
	"Zig":         BraceChunker,
	"Swift":       BraceChunker,
	"Scala":       BraceChunker,
	"Dart":        BraceChunker,
	"PHP":         BraceChunker,
	"Ruby":        BraceChunker,
	"Perl":        BraceChunker,
	"F#":          BraceChunker,
}

func ForLanguage(language string) Chunker {
	if fn, ok := registry[language]; ok {
		return fn
	}
	return FallbackChunker
}

func FallbackChunker(content string) []Chunk {
	return lineBasedChunk(content, 50, 10)
}

func lineBasedChunk(content string, chunkSize, overlap int) []Chunk {
	if chunkSize <= 0 {
		chunkSize = 50
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
		return []Chunk{{
			Content:   content,
			StartLine: 0,
			EndLine:   len(lines) - 1,
			Index:     0,
		}}
	}

	var chunks []Chunk
	start := 0
	index := 0
	for start < len(lines) {
		end := start + chunkSize
		if end > len(lines) {
			end = len(lines)
		}
		chunks = append(chunks, Chunk{
			Content:   strings.Join(lines[start:end], "\n"),
			StartLine: start,
			EndLine:   end - 1,
			Index:     index,
		})
		index++
		start += chunkSize - overlap
		if start >= len(lines) {
			break
		}
	}
	return chunks
}

type declInfo struct {
	line    int
	symbols []Symbol
	endLine int
}

func buildChunks(lines []string, decls []declInfo) []Chunk {
	if len(decls) == 0 {
		return nil
	}

	// Set end lines for all but the last decl
	for i := 0; i < len(decls); i++ {
		if decls[i].endLine < 0 {
			if i+1 < len(decls) {
				decls[i].endLine = decls[i+1].line
			} else {
				decls[i].endLine = len(lines)
			}
		}
	}

	var chunks []Chunk
	idx := 0
	pos := 0

	for _, d := range decls {
		if d.line > pos {
			chunks = append(chunks, Chunk{
				Content:   strings.Join(lines[pos:d.line], "\n"),
				StartLine: pos,
				EndLine:   d.line - 1,
				Index:     idx,
			})
			idx++
		}
		chunks = append(chunks, Chunk{
			Content:   strings.Join(lines[d.line:d.endLine], "\n"),
			StartLine: d.line,
			EndLine:   d.endLine - 1,
			Index:     idx,
			Symbols:   d.symbols,
		})
		idx++
		pos = d.endLine
	}

	if pos < len(lines) {
		chunks = append(chunks, Chunk{
			Content:   strings.Join(lines[pos:], "\n"),
			StartLine: pos,
			EndLine:   len(lines) - 1,
			Index:     idx,
		})
	}

	return chunks
}
