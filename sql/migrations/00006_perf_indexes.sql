-- +goose Up

CREATE INDEX IF NOT EXISTS idx_symbols_name_doc_id ON symbols(name, doc_id);
CREATE INDEX IF NOT EXISTS idx_refs_name_doc_id ON refs(name, doc_id);

-- +goose Down

DROP INDEX IF EXISTS idx_symbols_name_doc_id;
DROP INDEX IF EXISTS idx_refs_name_doc_id;
