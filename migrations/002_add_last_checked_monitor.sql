ALTER TABLE monitors ADD COLUMN last_checked_at timestamp;
CREATE INDEX monitor_last_checked_at_idx ON monitors(last_checked_at);
