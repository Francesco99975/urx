-- +goose Up

ALTER TABLE url_clicks
ADD COLUMN ip_hash BYTEA,
ADD COLUMN device TEXT,
ADD COLUMN os TEXT,
ADD COLUMN browser TEXT,
ADD COLUMN is_bot BOOLEAN DEFAULT FALSE;

-- Optional: normalize country to ISO-2 (if you're ready)
ALTER TABLE url_clicks
ALTER COLUMN country TYPE CHAR(2);

-- Optional indexes depending on analytics usage
CREATE INDEX idx_clicks_country ON url_clicks(country);
-- CREATE INDEX idx_clicks_device ON url_clicks(device);
-- CREATE INDEX idx_clicks_os ON url_clicks(os);
-- CREATE INDEX idx_clicks_browser ON url_clicks(browser);

CREATE INDEX idx_urls_user_id ON urls(user_id);

CREATE INDEX idx_url_clicks_url_id_created_at
ON url_clicks(url_id, created_at DESC);

CREATE INDEX idx_url_clicks_created_at
ON url_clicks(created_at DESC);
