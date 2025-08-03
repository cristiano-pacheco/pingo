DROP TABLE magic_tokens;
DROP TABLE users;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_magic_tokens_user;
DROP INDEX IF EXISTS idx_magic_tokens_token;