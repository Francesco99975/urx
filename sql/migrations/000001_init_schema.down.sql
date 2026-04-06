-- Drop indexes (optional, since DROP TABLE CASCADE handles them, but explicit is cleaner)

DROP INDEX IF EXISTS idx_clicks_created_at;
DROP INDEX IF EXISTS idx_clicks_url_id;

DROP INDEX IF EXISTS idx_urls_user_id_created_at;
DROP INDEX IF EXISTS idx_urls_expires_at;
DROP INDEX IF EXISTS idx_urls_active_code;
DROP INDEX IF EXISTS idx_urls_code;
DROP INDEX IF EXISTS idx_urls_long_url_hash;

DROP INDEX IF EXISTS idx_twofa_backup_codes_user_id;

DROP INDEX IF EXISTS idx_email_verifications_expires_at;
DROP INDEX IF EXISTS idx_email_verifications_token;
DROP INDEX IF EXISTS idx_email_verifications_user_id;

DROP INDEX IF EXISTS idx_password_resets_expires_at;
DROP INDEX IF EXISTS idx_password_resets_token;
DROP INDEX IF EXISTS idx_password_resets_user_id;

DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables (reverse dependency order)

DROP TABLE IF EXISTS url_clicks;
DROP TABLE IF EXISTS urls;

DROP TABLE IF EXISTS twofa_backup_codes;
DROP TABLE IF EXISTS email_verifications;
DROP TABLE IF EXISTS password_resets;

DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;

-- Drop trigger functions (this implicitly removes triggers tied to them)

DROP FUNCTION IF EXISTS apply_update_trigger(TEXT);
DROP FUNCTION IF EXISTS update_updated();
