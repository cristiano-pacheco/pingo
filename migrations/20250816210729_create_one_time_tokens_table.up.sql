CREATE TABLE one_time_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL,
    token_type VARCHAR(50) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Primary lookup index: for token verification
-- This covers the main query: WHERE user_id = ? AND token_hash = ? AND token_type = ?
CREATE INDEX idx_tokens_verification ON one_time_tokens (user_id, token_hash, token_type);

-- Cleanup index: for deleting old tokens before creating new ones
-- This covers: DELETE FROM one_time_tokens WHERE user_id = ? AND token_type = ?
CREATE INDEX idx_tokens_cleanup ON one_time_tokens (user_id, token_type);

-- Expiration cleanup index: for background jobs removing expired tokens
-- This covers: DELETE FROM one_time_tokens WHERE expires_at < NOW()
CREATE INDEX idx_tokens_expires ON one_time_tokens (expires_at);