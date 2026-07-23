-- name: GetGraphEdges :many
WITH symbol_freq AS (
    SELECT s.name
    FROM symbols s
    JOIN documents d ON d.id = s.doc_id
    WHERE d.repo_id = ?
    GROUP BY s.name
    HAVING COUNT(DISTINCT s.doc_id) <= 20
)
SELECT
  r.doc_id AS source_id,
  d.path AS source_path,
  s.doc_id AS target_id,
  d2.path AS target_path,
  COUNT(*) AS weight
FROM refs r
JOIN symbols s ON s.name = r.name
JOIN documents d ON d.id = r.doc_id
JOIN documents d2 ON d2.id = s.doc_id
WHERE d.id != d2.id
  AND d.repo_id = ?
  AND d2.repo_id = ?
  AND r.name IN (SELECT name FROM symbol_freq)
  AND s.name IN (SELECT name FROM symbol_freq)
GROUP BY r.doc_id, s.doc_id
ORDER BY weight DESC
LIMIT 1000;

-- name: GetGraphEdgesFromTable :many
SELECT
  ge.source_doc_id AS source_id,
  d.path AS source_path,
  ge.target_doc_id AS target_id,
  d2.path AS target_path,
  ge.weight
FROM graph_edges ge
JOIN documents d ON d.id = ge.source_doc_id
JOIN documents d2 ON d2.id = ge.target_doc_id
WHERE ge.repo_id = ?
  AND ge.kind = 'references'
ORDER BY ge.weight DESC
LIMIT 1000;

-- name: GetGraphNodes :many
SELECT id, path, repo_id FROM documents
WHERE repo_id = ?
ORDER BY id;
