package chunker

import (
	"regexp"
	"strings"
)

var (
	jsFuncPat     = regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:async\s+)?function\*?\s+([A-Za-z_$][\w$]*)`)
	jsClassPat    = regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:default\s+)?class\s+([A-Za-z_$][\w$]*)`)
	jsMethodPat   = regexp.MustCompile(`(?m)^\s*([A-Za-z_$][\w$]*)\s*\([^)]*\)\s*\{`)
	jsArrowPat    = regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=\s*(?:async\s*)?\(`)
	jsRustFnPat   = regexp.MustCompile(`(?m)^\s*(?:pub(?:\(\w+\)\s+)?)?(?:async\s+)?fn\s+([A-Za-z_$][\w$]*)`)
	jsRustImplPat = regexp.MustCompile(`(?m)^\s*impl(?:<[^>]*>)?\s+(?:[A-Za-z_$][\w$]*)\s*\{`)
	jsCommentRe   = regexp.MustCompile(`(?m)^\s*(//|/\*|\*|#)`)
)

type jsScanner struct {
	lines []string
	pos   int
}

func newJSScanner(content string) *jsScanner {
	return &jsScanner{lines: strings.Split(content, "\n")}
}

func (s *jsScanner) line() string {
	if s.pos >= len(s.lines) {
		return ""
	}
	return s.lines[s.pos]
}

func (s *jsScanner) eof() bool {
	return s.pos >= len(s.lines)
}

func (s *jsScanner) next() string {
	l := s.line()
	s.pos++
	return l
}

type jsDecl struct {
	line    int
	endLine int
	name    string
	kind    string
	col     int
}

func findMatchingBrace(lines []string, start int) int {
	depth := 0
	started := false
	for i := start; i < len(lines); i++ {
		line := lines[i]
		for _, r := range line {
			switch r {
			case '{':
				depth++
				started = true
			case '}':
				if depth > 0 {
					depth--
					if started && depth == 0 {
						return i
					}
				}
			}
		}
	}
	return len(lines) - 1
}

func findMatchingParen(lines []string, start int) int {
	depth := 0
	started := false
	for i := start; i < len(lines); i++ {
		line := lines[i]
		for _, r := range line {
			switch r {
			case '(':
				depth++
				started = true
			case ')':
				if depth > 0 {
					depth--
					if started && depth == 0 {
						return i
					}
				}
			}
		}
	}
	return len(lines) - 1
}

func isJSCommentOrDirective(line string) bool {
	return jsCommentRe.MatchString(line)
}

func findJSDecls(lines []string) []jsDecl {
	var decls []jsDecl
	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || isJSCommentOrDirective(line) {
			i++
			continue
		}

		var name, kind string
		isDecl := false

		if m := jsFuncPat.FindStringSubmatch(line); m != nil {
			name, kind = m[1], "function"
			isDecl = true
		} else if m := jsClassPat.FindStringSubmatch(line); m != nil {
			name, kind = m[1], "class"
			isDecl = true
		} else if m := jsArrowPat.FindStringSubmatch(line); m != nil {
			name, kind = m[1], "function"
			isDecl = true
		} else if m := jsRustFnPat.FindStringSubmatch(line); m != nil {
			name, kind = m[1], "function"
			isDecl = true
		} else if m := jsRustImplPat.FindStringSubmatch(line); m != nil {
			name, kind = m[1], "impl"
			isDecl = true
		}

		if !isDecl {
			if m := jsMethodPat.FindStringSubmatch(line); m != nil {
				name, kind = m[1], "function"
				isDecl = true
			}
		}

		if isDecl {
			endLine := i
			if strings.ContainsRune(line, '{') {
				endLine = findMatchingBrace(lines, i)
			} else {
				for j := i + 1; j < len(lines); j++ {
					if strings.ContainsRune(lines[j], '{') {
						endLine = findMatchingBrace(lines, j)
						break
					}
					if strings.TrimSpace(lines[j]) != "" {
						break
					}
				}
			}
			col := strings.Index(line, name)
			if col < 0 {
				col = 0
			}
			decls = append(decls, jsDecl{
				line:    i,
				endLine: endLine,
				name:    name,
				kind:    kind,
				col:     col,
			})
			i = endLine + 1
			continue
		}

		i++
	}
	return decls
}

func jsDeclsToChunks(lines []string, decls []jsDecl) []Chunk {
	if len(decls) == 0 {
		return nil
	}
	var chunks []Chunk
	for idx, d := range decls {
		end := d.endLine
		if end >= len(lines) {
			end = len(lines) - 1
		}
		if d.line > end {
			end = d.line
		}
		content := strings.Join(lines[d.line:end+1], "\n")
		chunks = append(chunks, Chunk{
			Content:   content,
			StartLine: d.line,
			EndLine:   end,
			Index:     idx,
			Symbols:   []Symbol{{Name: d.name, Kind: d.kind, Line: d.line, Col: d.col}},
		})
	}
	return chunks
}

func JSChunker(content string) []Chunk {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil
	}
	decls := findJSDecls(lines)
	chunks := jsDeclsToChunks(lines, decls)
	if chunks == nil {
		return lineBasedChunk(content, 50, 0)
	}
	return chunks
}

func RustChunker(content string) []Chunk {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil
	}
	decls := findJSDecls(lines)
	chunks := jsDeclsToChunks(lines, decls)
	if chunks == nil {
		return lineBasedChunk(content, 50, 0)
	}
	return chunks
}
