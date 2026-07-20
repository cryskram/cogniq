package indexer

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/cryskram/relith/internal/config"
	"github.com/cryskram/relith/internal/db"
)

type IndexResult struct {
	FilesIndexed int
	FilesSkipped int
	FilesError   int
	TotalChunks  int
	Elapsed      time.Duration
}

type Indexer struct {
	db     *sql.DB
	logger zerolog.Logger
	cfg    config.IndexerConfig
}

func New(database *sql.DB, logger zerolog.Logger, cfg config.IndexerConfig) *Indexer {
	return &Indexer{
		db:     database,
		logger: logger,
		cfg:    cfg,
	}
}

func (idx *Indexer) queries() *db.Queries {
	return db.New(idx.db)
}

func (idx *Indexer) IndexRepo(ctx context.Context, repoPath string, repoID int64) (IndexResult, error) {
	start := time.Now()
	idx.logger.Info().Str("path", repoPath).Int64("repo_id", repoID).Msg("indexing repo")

	q := idx.queries()
	if err := q.UpdateRepoStatus(ctx, db.UpdateRepoStatusParams{
		ID:     repoID,
		Status: "indexing",
	}); err != nil {
		return IndexResult{}, fmt.Errorf("set status indexing: %w", err)
	}

	files, err := WalkRepo(repoPath, idx.cfg.MaxFileSize)
	if err != nil {
		return IndexResult{}, fmt.Errorf("walk repo: %w", err)
	}

	idx.logger.Info().Int("files_found", len(files)).Msg("walk complete")

	existingDocs, err := q.ListDocuments(ctx, repoID)
	if err != nil {
		return IndexResult{}, fmt.Errorf("list existing docs: %w", err)
	}

	existingByPath := make(map[string]db.Document, len(existingDocs))
	for _, doc := range existingDocs {
		existingByPath[doc.Path] = doc
	}

	visited := sync.Map{}

	concurrency := idx.cfg.Concurrency
	if concurrency < 1 {
		concurrency = 1
	}

	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex
	result := IndexResult{}
	var wg sync.WaitGroup

	for _, fi := range files {
		fi := fi
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if err := idx.processFile(ctx, fi, repoID, existingByPath, &visited, &mu, &result); err != nil {
				idx.logger.Error().Err(err).Str("file", fi.RelPath).Msg("processing file")
				mu.Lock()
				result.FilesError++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	var toDelete []int64
	for _, doc := range existingDocs {
		if _, seen := visited.Load(doc.Path); !seen {
			toDelete = append(toDelete, doc.ID)
		}
	}
	for _, id := range toDelete {
		if err := q.DeleteDocument(ctx, id); err != nil {
			idx.logger.Error().Err(err).Int64("doc_id", id).Msg("delete stale document")
		}
	}

	now := time.Now()
	if err := q.UpdateRepoStatus(ctx, db.UpdateRepoStatusParams{
		ID:            repoID,
		Status:        "ready",
		LastIndexedAt: sql.NullTime{Time: now, Valid: true},
		FileCount:     int64(len(files) - result.FilesError),
	}); err != nil {
		return result, fmt.Errorf("set status ready: %w", err)
	}

	result.Elapsed = time.Since(start)
	idx.logger.Info().
		Int("indexed", result.FilesIndexed).
		Int("skipped", result.FilesSkipped).
		Int("errors", result.FilesError).
		Int("chunks", result.TotalChunks).
		Int("stale_removed", len(toDelete)).
		Dur("elapsed", result.Elapsed).
		Msg("indexing complete")

	return result, nil
}

func (idx *Indexer) processFile(
	ctx context.Context,
	fi FileInfo,
	repoID int64,
	existingByPath map[string]db.Document,
	visited *sync.Map,
	mu *sync.Mutex,
	result *IndexResult,
) error {
	visited.Store(fi.RelPath, struct{}{})

	content, err := ReadFileContent(fi.FullPath, idx.cfg.MaxFileSize)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	if content == "" {
		mu.Lock()
		result.FilesSkipped++
		mu.Unlock()
		return nil
	}

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))

	existing, exists := existingByPath[fi.RelPath]
	if exists && existing.Hash == hash {
		mu.Lock()
		result.FilesSkipped++
		mu.Unlock()
		return nil
	}

	lang := DetectLanguage(fi.RelPath)
	mimeStr := ""
	if lang != "" {
		mimeStr = "text/" + lang
	}
	langStr := lang

	chunks := ChunkContent(content, DefaultChunkSize, DefaultChunkOverlap)

	tx, err := idx.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	qtx := idx.queries().WithTx(tx)

	if exists {
		if err := qtx.UpdateDocument(ctx, db.UpdateDocumentParams{
			ID:       existing.ID,
			Size:     fi.Size,
			Hash:     hash,
			ModTime:  time.Unix(fi.ModTime, 0),
			MimeType: sql.NullString{String: mimeStr, Valid: mimeStr != ""},
			Language: sql.NullString{String: langStr, Valid: langStr != ""},
		}); err != nil {
			return fmt.Errorf("update doc: %w", err)
		}
		if err := qtx.DeleteChunksByDoc(ctx, existing.ID); err != nil {
			return fmt.Errorf("delete chunks: %w", err)
		}
		for _, c := range chunks {
			if _, err := qtx.CreateChunk(ctx, db.CreateChunkParams{
				DocID:      existing.ID,
				ChunkIndex: int64(c.Index),
				Content:    c.Content,
			}); err != nil {
				return fmt.Errorf("create chunk: %w", err)
			}
		}
	} else {
		doc, err := qtx.CreateDocument(ctx, db.CreateDocumentParams{
			RepoID:   repoID,
			Path:     fi.RelPath,
			Size:     fi.Size,
			Hash:     hash,
			ModTime:  time.Unix(fi.ModTime, 0),
			MimeType: sql.NullString{String: mimeStr, Valid: mimeStr != ""},
			Language: sql.NullString{String: langStr, Valid: langStr != ""},
		})
		if err != nil {
			return fmt.Errorf("create doc: %w", err)
		}
		for _, c := range chunks {
			if _, err := qtx.CreateChunk(ctx, db.CreateChunkParams{
				DocID:      doc.ID,
				ChunkIndex: int64(c.Index),
				Content:    c.Content,
			}); err != nil {
				return fmt.Errorf("create chunk: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	mu.Lock()
	result.FilesIndexed++
	result.TotalChunks += len(chunks)
	mu.Unlock()

	return nil
}

func IsGitRepo(path string) bool {
	info, err := os.Stat(path + "/.git")
	if err != nil {
		return false
	}
	return info.IsDir()
}
