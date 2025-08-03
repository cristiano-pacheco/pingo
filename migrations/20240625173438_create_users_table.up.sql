-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    email VARCHAR(320) NOT NULL,
    password_hash BYTEA NOT NULL,
    "status" VARCHAR(50) NOT NULL,
    reset_password_token BYTEA NULL,
    account_confirmation_token BYTEA NULL,
    last_login_at TIMESTAMP NULL,
    email_verified_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS user_email_idx ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_reset_token ON users(reset_password_token) WHERE reset_password_token IS NOT NULL;