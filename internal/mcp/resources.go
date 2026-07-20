package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func (s *Server) handleResourceRepo(ctx context.Context, uri string) []ResourceContents {
	parts := strings.Split(strings.TrimPrefix(uri, "relith://repos/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		repos, err := s.queries.ListRepos(ctx)
		if err != nil {
			return []ResourceContents{{
				URI: uri, MimeType: "text/plain",
				Text: fmt.Sprintf("error: %v", err),
			}}
		}
		data, _ := json.MarshalIndent(repos, "", "  ")
		return []ResourceContents{{
			URI: uri, MimeType: "application/json",
			Text: string(data),
		}}
	}

	return []ResourceContents{{
		URI: uri, MimeType: "text/plain",
		Text: "resource not implemented",
	}}
}
