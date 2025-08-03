CREATE TABLE IF NOT EXISTS http_monitors (
    id BIGSERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    check_timeout INTEGER NOT NULL,
    fail_threshold SMALLINT NOT NULL,
    check_interval_seconds INTEGER NOT NULL DEFAULT 300,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    http_url VARCHAR(2048) NOT NULL,
    http_method VARCHAR(10) NOT NULL,
    request_headers JSONB NOT NULL DEFAULT '{}',
    valid_response_statuses INTEGER[] NOT NULL DEFAULT '{200}',
    last_checked_at TIMESTAMP NULL,
    last_status VARCHAR(100) NULL,
    consecutive_failures INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS http_monitor_contacts (
    http_monitor_id BIGINT NOT NULL,
    contact_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (http_monitor_id, contact_id),
    CONSTRAINT fk_monitor_contact_monitor FOREIGN KEY (http_monitor_id) REFERENCES http_monitors(id) ON DELETE CASCADE,
    CONSTRAINT fk_monitor_contact_contact FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS http_monitor_checks (
    id BIGSERIAL PRIMARY KEY,
    http_monitor_id BIGINT NOT NULL,
    checked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    response_time_ms INTEGER NULL,
    status_code INTEGER NULL,
    success BOOLEAN NOT NULL,
    error_message TEXT NULL,
    CONSTRAINT fk_monitor_check_monitor FOREIGN KEY (http_monitor_id) REFERENCES http_monitors(id) ON DELETE CASCADE
);