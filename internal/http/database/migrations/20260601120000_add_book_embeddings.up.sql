-- pgvector enables efficient semantic similarity search in PostgreSQL.
-- JSONB is not suitable for nearest-neighbor vector search (no <-> operator,
-- no vector indexes, slower distance computations).
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS book_embeddings (
    book_id UUID PRIMARY KEY,
    embedding VECTOR(1536) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (book_id)
        REFERENCES books(id)
        ON DELETE CASCADE
);
