package chunker

import (
	"regexp"
	"strings"
)

var (
	braceFuncPat = regexp.MustCompile(`(?i)^(async\s+)?(public\s+|private\s+|protected\s+|static\s+|export\s+|default\s+)*(function\s+|fn\s+|def\s+)?\s*([A-Za-z_]\w*)\s*\(`)
	braceClassPat = regexp.MustCompile(`(?i)^(public\s+|private\s+|protected\s+|abstract\s+|static\s+|export\s+|default\s+)*(class|struct|interface|trait|enum|impl)\s+([A-Za-z_]\w*)`)

	braceSkipPats = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^\s*(import|package|using|namespace|#include|#define|#if|#endif|#pragma|#region|#endregion)\s`),
		regexp.MustCompile(`^\s*@\w+`),
	}
)

func BraceChunker(content string) []Chunk {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil
	}

	var decls []declInfo
	braceDepth := 0
	inDecl := -1

	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			updateBrace(&braceDepth, lines[i])
			if inDecl >= 0 && braceDepth == 0 {
				decls[len(decls)-1].endLine = i + 1
				inDecl = -1
			}
			continue
		}

		isSkip := false
		for _, p := range braceSkipPats {
			if p.MatchString(trimmed) {
				isSkip = true
				break
			}
		}
		if isSkip {
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

		if m := braceClassPat.FindStringSubmatch(trimmed); m != nil {
			name = m[len(m)-1]
			kind = strings.ToLower(m[len(m)-2])
			if kind == "impl" {
				kind = "impl"
			}
			col = strings.Index(lines[i], name)
			if col < 0 {
				col = strings.Index(trimmed, name)
			}
			isDecl = true
		} else if m := braceFuncPat.FindStringSubmatch(trimmed); m != nil {
			name = m[len(m)-1]
			kind = "function"
			col = strings.Index(lines[i], name)
			if col < 0 {
				col = strings.Index(trimmed, name)
			}
			isDecl = true
		}

		if isDecl {
			if inDecl >= 0 {
				decls[len(decls)-1].endLine = i
			}

			bb := countBraces(lines[i])

			if strings.Contains(trimmed, "{") || strings.Count(trimmed, "(")+strings.Count(trimmed, ")") > 0 {
				if bb > 0 {
					braceDepth = bb - 1
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
					braceDepth = 0
					inDecl = -1
					decls = append(decls, declInfo{
						line:    i,
						endLine: i + 1,
						symbols: []Symbol{{Name: name, Kind: kind, Line: i, Col: col}},
					})
				}
			} else {
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

	if inDecl >= 0 && len(decls) > 0 {
		decls[len(decls)-1].endLine = len(lines)
	}

	chunks := buildChunks(lines, decls)
	if chunks == nil {
		return lineBasedChunk(content, 50, 10)
	}
	return chunks
}
