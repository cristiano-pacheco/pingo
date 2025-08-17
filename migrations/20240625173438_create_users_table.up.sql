-- Create user status enum
DROP TYPE IF EXISTS user_status;
CREATE TYPE user_status AS ENUM ('pending', 'active', 'inactive', 'suspended');

-- Create users table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    status user_status NOT NULL DEFAULT 'pending',
    password_hash BYTEA NOT NULL,
    confirmation_token BYTEA,
    confirmed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_confirmation_token ON users(confirmation_token);
CREATE INDEX idx_magic_tokens_user ON magic_tokens(user_id);
CREATE INDEX idx_magic_tokens_token ON magic_tokens(token);