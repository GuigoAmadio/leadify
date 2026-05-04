-- Enable the pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create the core table
CREATE TABLE instagram_profiles (
    -- Core Identifiers
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    profile_url VARCHAR(512),
    
    -- Hard Relational Filters
    entity_type VARCHAR(100),
    monetization_platforms TEXT[],
    resolved_outbound_links TEXT[],
    vibe_and_controversy VARCHAR(100),
    
    -- The Raw Extracted Text
    full_semantic_json JSONB,
    
    -- The Multi-Vector Indexes (Truncated to 256 dimensions for MRL)
    vector_niche vector(256),
    vector_appearance vector(256),
    vector_aesthetic vector(256),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create HNSW indexes for lightning-fast cosine similarity search
-- These ensure your searches stay under 50ms even with tens of thousands of rows
CREATE INDEX ON instagram_profiles USING hnsw (vector_niche vector_cosine_ops);
CREATE INDEX ON instagram_profiles USING hnsw (vector_appearance vector_cosine_ops);
CREATE INDEX ON instagram_profiles USING hnsw (vector_aesthetic vector_cosine_ops);