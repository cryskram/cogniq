package search

import "strings"

var fts5Special = []string{`"`, `(`, `)`, `*`, `^`, `+`, `-`, `~`, `:`, "AND", "OR", "NOT", "NEAR"}

func hasFTS5Operators(input string) bool {
	upper := strings.ToUpper(input)
	for _, op := range fts5Special {
		if strings.Contains(upper, op) {
			return true
		}
	}
	return false
}

func escapeFTS5Term(term string) string {
	escaped := strings.NewReplacer(
		`"`, `""`,
		`(`, `(`,
		`)`, `)`,
	).Replace(term)
	return `"` + escaped + `"`
}

func buildMatchQuery(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	if hasFTS5Operators(input) {
		return input
	}

	terms := strings.Fields(input)
	if len(terms) == 0 {
		return ""
	}

	if len(terms) == 1 {
		return escapeFTS5Term(terms[0]) + "*"
	}

	var parts []string
	for _, t := range terms {
		parts = append(parts, escapeFTS5Term(t))
	}
	return strings.Join(parts, " ")
}
