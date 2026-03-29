CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    channel VARCHAR(20) NOT NULL,
    type VARCHAR(40) NOT NULL,
    recipient TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    provider VARCHAR(40) NOT NULL,
    status VARCHAR(20) NOT NULL,
    reference_id TEXT NULL,
    provider_message_id TEXT NULL,
    provider_response TEXT NULL,
    error_message TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_notifications_recipient ON notifications(recipient);
CREATE INDEX idx_notifications_type_created_at ON notifications(type, created_at DESC);
CREATE INDEX idx_notifications_status_created_at ON notifications(status, created_at DESC);
