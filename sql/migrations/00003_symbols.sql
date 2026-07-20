-- +goose Up

CREATE TABLE symbols (
    id       INTEGER PRIMARY KEY,
    doc_id   INTEGER NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    name     TEXT NOT NULL,
    kind     TEXT NOT NULL,
    line     INTEGER NOT NULL,
    col      INTEGER NOT NULL
);
CREATE INDEX idx_symbols_name ON symbols(name);
CREATE INDEX idx_symbols_doc_id ON symbols(doc_id);
CREATE INDEX idx_symbols_kind ON symbols(kind);

-- +goose Down

DROP INDEX IF EXISTS idx_symbols_kind;
DROP INDEX IF EXISTS idx_symbols_doc_id;
DROP INDEX IF EXISTS idx_symbols_name;
DROP TABLE IF EXISTS symbols;
