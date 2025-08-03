CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    http_monitor_id BIGINT NOT NULL,
    contact_id BIGINT NOT NULL,
    notification_type VARCHAR(50) NOT NULL, -- 'failure', 'recovery', 'maintenance'
    message TEXT NOT NULL,
    sent_at TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'sent', 'failed'
    error_message TEXT NULL,
    CONSTRAINT fk_notification_monitor FOREIGN KEY (http_monitor_id) REFERENCES http_monitors(id) ON DELETE CASCADE,
    CONSTRAINT fk_notification_contact FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_notifications_monitor ON notifications(http_monitor_id, sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_contact ON notifications(contact_id, sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status, sent_at);
