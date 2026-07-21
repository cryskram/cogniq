package chunker

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func GoChunkerAST(content string) []Chunk {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil
	}

	lines := strings.Split(content, "\n")
	var decls []declInfo

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			name := d.Name.Name
			kind := "function"
			if d.Recv != nil {
				kind = "method"
			}
			startPos := d.Pos()
			if d.Doc != nil {
				startPos = d.Doc.Pos()
			}
			startLine := fset.Position(startPos).Line - 1
			col := fset.Position(d.Name.NamePos).Column - 1
			startOffset := fset.Position(startPos).Offset
			endOffset := fset.Position(d.End()).Offset
			extracted := content[startOffset:endOffset]
			actualEndLine := startLine + strings.Count(extracted, "\n")
			decls = append(decls, declInfo{
				line:    startLine,
				endLine: actualEndLine + 1,
				symbols: []Symbol{{Name: name, Kind: kind, Line: startLine, Col: col}},
			})
		case *ast.GenDecl:
			if d.Tok == token.IMPORT {
				continue
			}
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					name := s.Name.Name
					kind := "type"
					if _, ok := s.Type.(*ast.StructType); ok {
						kind = "struct"
					} else if _, ok := s.Type.(*ast.InterfaceType); ok {
						kind = "interface"
					}
					startPos := s.Pos()
					if d.Doc != nil {
						startPos = d.Doc.Pos()
					}
					startLine := fset.Position(startPos).Line - 1
					col := fset.Position(s.Name.NamePos).Column - 1
					startOffset := fset.Position(startPos).Offset
					endOffset := fset.Position(s.End()).Offset
					extracted := content[startOffset:endOffset]
					actualEndLine := startLine + strings.Count(extracted, "\n")
					decls = append(decls, declInfo{
						line:    startLine,
						endLine: actualEndLine + 1,
						symbols: []Symbol{{Name: name, Kind: kind, Line: startLine, Col: col}},
					})
				case *ast.ValueSpec:
					if len(s.Names) == 0 {
						continue
					}
					name := s.Names[0].Name
					kind := "variable"
					if d.Tok == token.CONST {
						kind = "constant"
					}
					startPos := s.Pos()
					if d.Doc != nil {
						startPos = d.Doc.Pos()
					}
					startLine := fset.Position(startPos).Line - 1
					col := fset.Position(s.Names[0].NamePos).Column - 1
					startOffset := fset.Position(startPos).Offset
					endOffset := fset.Position(s.End()).Offset
					extracted := content[startOffset:endOffset]
					actualEndLine := startLine + strings.Count(extracted, "\n")
					decls = append(decls, declInfo{
						line:    startLine,
						endLine: actualEndLine + 1,
						symbols: []Symbol{{Name: name, Kind: kind, Line: startLine, Col: col}},
					})
				}
			}
		}
	}

	if len(decls) == 0 {
		return nil
	}

	chunks := buildChunks(lines, decls)
	if chunks == nil {
		return lineBasedChunk(content, 50, 0)
	}
	return chunks
}
