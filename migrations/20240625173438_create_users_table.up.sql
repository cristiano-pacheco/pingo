-- Create user status enum
CREATE TYPE user_status AS ENUM ('pending', 'active', 'inactive', 'suspended');

-- Create users table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    status user_status NOT NULL DEFAULT 'pending',
    confirmation_token BYTEA,
    confirmed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create magic tokens table for one-time login links
CREATE TABLE magic_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token BYTEA NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ,
    CONSTRAINT unused_token CHECK (used_at IS NULL OR created_at <= used_at)
);

-- Create indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_magic_tokens_user ON magic_tokens(user_id);
CREATE INDEX idx_magic_tokens_token ON magic_tokens(token);