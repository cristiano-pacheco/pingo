CREATE TABLE verification_codes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code CHAR(6) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ,
    CONSTRAINT unused_code CHECK (used_at IS NULL OR created_at <= used_at),
    CONSTRAINT valid_code_format CHECK (code ~ '^[0-9]{6}$')
);

CREATE UNIQUE INDEX idx_verification_lookup ON verification_codes (user_id, code);