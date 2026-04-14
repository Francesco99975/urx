-- +goose Down

-- DROP INDEX IF EXISTS idx_clicks_browser;
-- DROP INDEX IF EXISTS idx_clicks_os;
-- DROP INDEX IF EXISTS idx_clicks_device;
DROP INDEX IF EXISTS idx_clicks_country;

DROP INDEX IF EXISTS idx_urls_user_id;
DROP INDEX IF EXISTS idx_url_clicks_url_id_created_at;
DROP INDEX IF EXISTS idx_url_clicks_created_at;

ALTER TABLE url_clicks
DROP COLUMN IF EXISTS is_bot,
DROP COLUMN IF EXISTS browser,
DROP COLUMN IF EXISTS os,
DROP COLUMN IF EXISTS device;

ALTER TABLE url_clicks
ALTER COLUMN country TYPE TEXT;
