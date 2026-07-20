package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cryskram/relith/internal/db"
	"github.com/cryskram/relith/internal/indexer"
	"github.com/cryskram/relith/internal/search"
)

type handlers struct {
	queries  *db.Queries
	indexer  *indexer.Indexer
	searcher *search.Searcher
}

func (h *handlers) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handlers) listRepos(w http.ResponseWriter, r *http.Request) {
	repos, err := h.queries.ListRepos(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list repos: "+err.Error())
		return
	}
	if repos == nil {
		repos = []db.Repository{}
	}
	writeJSON(w, http.StatusOK, repos)
}

func (h *handlers) createRepo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path      string `json:"path"`
		Name      string `json:"name"`
		RemoteURL string `json:"remote_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	if req.Path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	repo, err := h.queries.CreateRepo(r.Context(), db.CreateRepoParams{
		Path:      req.Path,
		Name:      req.Name,
		RemoteUrl: sql.NullString{String: req.RemoteURL, Valid: req.RemoteURL != ""},
	})
	if err != nil {
		writeError(w, http.StatusConflict, "create repo: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, repo)
}

func (h *handlers) getRepo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid repo id")
		return
	}
	repo, err := h.queries.GetRepo(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "repo not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "get repo: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, repo)
}

func (h *handlers) deleteRepo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid repo id")
		return
	}
	if err := h.queries.DeleteRepo(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "delete repo: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handlers) indexRepo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid repo id")
		return
	}

	repo, err := h.queries.GetRepo(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "repo not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "get repo: "+err.Error())
		return
	}

	result, err := h.indexer.IndexRepo(r.Context(), repo.Path, repo.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "index: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"files_indexed": result.FilesIndexed,
		"files_skipped": result.FilesSkipped,
		"files_error":   result.FilesError,
		"total_chunks":  result.TotalChunks,
		"elapsed":       result.Elapsed.String(),
	})
}

func (h *handlers) content(w http.ResponseWriter, r *http.Request) {
	repoName := r.URL.Query().Get("repo")
	filePath := r.URL.Query().Get("path")
	if repoName == "" || filePath == "" {
		writeError(w, http.StatusBadRequest, "repo and path query params are required")
		return
	}

	repos, err := h.queries.ListRepos(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list repos: "+err.Error())
		return
	}

	var repo db.Repository
	found := false
	for _, rp := range repos {
		if rp.Name == repoName {
			repo = rp
			found = true
			break
		}
	}
	if !found {
		writeError(w, http.StatusNotFound, "repo not found: "+repoName)
		return
	}

	doc, err := h.queries.GetDocumentByPath(r.Context(), db.GetDocumentByPathParams{
		RepoID: repo.ID,
		Path:   filePath,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "file not found: "+filePath)
			return
		}
		writeError(w, http.StatusInternalServerError, "get document: "+err.Error())
		return
	}

	fullPath := filepath.Join(repo.Path, doc.Path)
	content, err := indexer.ReadFileContent(fullPath, 10*1024*1024)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "read file: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}

func (h *handlers) search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 0
	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	results, err := h.searcher.Search(ctx, query, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, results)
}


