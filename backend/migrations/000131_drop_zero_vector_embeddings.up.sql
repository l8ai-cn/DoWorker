-- Drop zero-vector embeddings produced by the legacy ASCII-only tokenizer
-- (any non-ASCII text → empty tokens → zero vector → pgvector NaN distance
-- → MCP 500). Affected blocks re-embed on their next write via embedBlock's
-- hash-mismatch path.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'block_embeddings'
          AND column_name = 'vec'
    ) THEN
        EXECUTE 'DELETE FROM block_embeddings WHERE vector_norm(vec) = 0';
    END IF;
END
$$;
