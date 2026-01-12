-- 002_vectors.sql: pgvector extension and documents table for RAG

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS documents (
    id SERIAL PRIMARY KEY,
    source TEXT NOT NULL, -- 'boe', 'dgt', 'local_news', 'manual'
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    summary TEXT,
    url TEXT,
    embedding_dims INTEGER, -- Dimension of the embedding (768 for Ollama, 1536 for OpenAI)
    embedding vector(768), -- Default to 768 for nomic-embed-text / Gemini
    metadata JSONB,
    indexed_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP -- For time-sensitive info
);

CREATE INDEX IF NOT EXISTS idx_documents_source ON documents(source);
-- IVFFlat index for cosine similarity search (requires fixed dimensions)
CREATE INDEX IF NOT EXISTS idx_documents_embedding ON documents USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
