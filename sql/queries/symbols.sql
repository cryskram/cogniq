-- name: CreateSymbol :one
INSERT INTO symbols (doc_id, name, kind, line, col)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: DeleteSymbolsByDoc :exec
DELETE FROM symbols
WHERE doc_id = ?;

-- name: FindSymbolsByName :many
SELECT s.*, d.path, d.repo_id, r.name AS repo_name
FROM symbols s
JOIN documents d ON d.id = s.doc_id
JOIN repositories r ON r.id = d.repo_id
WHERE s.name LIKE ? || '%'
ORDER BY s.name, d.path;

-- name: FindSymbolsByNameAndKind :many
SELECT s.*, d.path, d.repo_id, r.name AS repo_name
FROM symbols s
JOIN documents d ON d.id = s.doc_id
JOIN repositories r ON r.id = d.repo_id
WHERE s.name LIKE ? || '%' AND s.kind = ?
ORDER BY s.name, d.path;

-- name: FindSymbolsByRepo :many
SELECT s.*, d.path, d.repo_id, r.name AS repo_name
FROM symbols s
JOIN documents d ON d.id = s.doc_id
JOIN repositories r ON r.id = d.repo_id
WHERE r.name = ? AND s.name LIKE ? || '%'
ORDER BY s.name, d.path;

-- name: FindSymbolsByRepoAndKind :many
SELECT s.*, d.path, d.repo_id, r.name AS repo_name
FROM symbols s
JOIN documents d ON d.id = s.doc_id
JOIN repositories r ON r.id = d.repo_id
WHERE r.name = ? AND s.name LIKE ? || '%' AND s.kind = ?
ORDER BY s.name, d.path;
