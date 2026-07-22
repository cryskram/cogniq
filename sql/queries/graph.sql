-- name: GetGraphEdges :many
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
GROUP BY r.doc_id, s.doc_id
ORDER BY weight DESC
LIMIT 1000;

-- name: GetGraphNodes :many
SELECT id, path, repo_id FROM documents
WHERE repo_id = ?
ORDER BY id;
