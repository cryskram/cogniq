package chunker

import (
	"regexp"
	"strings"
)

var (
	goFuncPat   = regexp.MustCompile(`^func\s+(\([^)]*\)\s+)?([A-Za-z_]\w*)`)
	goTypePat   = regexp.MustCompile(`^type\s+([A-Za-z_]\w*)`)
	goStructPat = regexp.MustCompile(`^type\s+([A-Za-z_]\w*)\s+struct(\s*\{)?$`)
	goIfacePat  = regexp.MustCompile(`^type\s+([A-Za-z_]\w*)\s+interface(\s*\{)?$`)
)

func GoChunker(content string) []Chunk {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil
	}

	var decls []declInfo
	braceDepth := 0
	inDecl := -1

	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || skipLine(trimmed) {
			updateBrace(&braceDepth, lines[i])
			if inDecl >= 0 && braceDepth == 0 {
				decls[len(decls)-1].endLine = i + 1
				inDecl = -1
			}
			continue
		}

		var name, kind string
		col := 0
		isDecl := false

		if m := goFuncPat.FindStringSubmatch(trimmed); m != nil {
			name = m[2]
			kind = "function"
			if strings.HasPrefix(trimmed, "func (") {
				kind = "method"
			}
			col = strings.Index(lines[i], name)
			if col < 0 {
				col = strings.Index(trimmed, name)
			}
			isDecl = true
		} else if m := goTypePat.FindStringSubmatch(trimmed); m != nil {
			name = m[1]
			col = strings.Index(lines[i], name)
			if col < 0 {
				col = strings.Index(trimmed, name)
			}
			if goStructPat.MatchString(trimmed) {
				kind = "struct"
			} else if goIfacePat.MatchString(trimmed) {
				kind = "interface"
			} else {
				kind = "type"
			}
			isDecl = true
		}

		if isDecl {
			// Close previous open declaration
			if inDecl >= 0 {
				decls[len(decls)-1].endLine = i
			}

			bb := countBraces(lines[i])

			if bb > 0 {
				// Has opening brace(s) on same line — track scope
				braceDepth = bb - 1 // count the opening brace
				inDecl = i
				decls = append(decls, declInfo{
					line:    i,
					endLine: -1,
					symbols: []Symbol{{Name: name, Kind: kind, Line: i, Col: col}},
				})
				if braceDepth == 0 {
					decls[len(decls)-1].endLine = i + 1
					inDecl = -1
				}
			} else {
				// No brace — single line or simple type alias
				braceDepth = 0
				inDecl = -1
				decls = append(decls, declInfo{
					line:    i,
					endLine: i + 1,
					symbols: []Symbol{{Name: name, Kind: kind, Line: i, Col: col}},
				})
			}
			continue
		}

		updateBrace(&braceDepth, lines[i])
		if inDecl >= 0 && braceDepth == 0 {
			decls[len(decls)-1].endLine = i + 1
			inDecl = -1
		}
	}

	// Close any remaining open declaration
	if inDecl >= 0 && len(decls) > 0 {
		decls[len(decls)-1].endLine = len(lines)
	}

	chunks := buildChunks(lines, decls)
	if chunks == nil {
		return lineBasedChunk(content, 50, 10)
	}
	return chunks
}

func countBraces(s string) int {
	return strings.Count(s, "{") - strings.Count(s, "}")
}

func updateBrace(depth *int, line string) {
	*depth += countBraces(line)
	if *depth < 0 {
		*depth = 0
	}
}

func skipLine(trimmed string) bool {
	if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "package ") || strings.HasPrefix(trimmed, "import") {
		return true
	}
	return false
}
