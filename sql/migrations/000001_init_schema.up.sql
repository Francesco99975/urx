-- Create a trigger function to update the updated column
CREATE OR REPLACE FUNCTION update_updated()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a macro to apply the trigger to a table
CREATE OR REPLACE FUNCTION apply_update_trigger(table_name TEXT)
RETURNS VOID AS $$
BEGIN
 IF NOT EXISTS (
    SELECT 1
    FROM information_schema.triggers
    WHERE trigger_schema = 'public'
      AND trigger_name = format('trigger_update_updated_%I', table_name)
  ) THEN
    EXECUTE format('
        CREATE TRIGGER trigger_update_updated_%I
        BEFORE UPDATE ON %I
        FOR EACH ROW
        EXECUTE FUNCTION update_updated()
    ', table_name, table_name);
  END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE roles (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (id) VALUES
    ('DEVELOPER'),
    ('ADMIN'),
    ('USER')
ON CONFLICT (id) DO NOTHING;

SELECT apply_update_trigger('roles');

CREATE TABLE users (
    id UUID PRIMARY KEY,
    role TEXT NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    username VARCHAR(21) UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    twofa_secret TEXT,
    twofa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    last_login TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

SELECT apply_update_trigger('users');


CREATE TABLE password_resets (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_password_resets_user_id ON password_resets(user_id);
CREATE INDEX idx_password_resets_token ON password_resets(token);
CREATE INDEX idx_password_resets_expires_at ON password_resets(expires_at);

-- Email verifications
CREATE TABLE email_verifications (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_verifications_user_id ON email_verifications(user_id);
CREATE INDEX idx_email_verifications_token ON email_verifications(token);
CREATE INDEX idx_email_verifications_expires_at ON email_verifications(expires_at);

CREATE TABLE twofa_backup_codes (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash TEXT NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_twofa_backup_codes_user_id ON twofa_backup_codes(user_id);



CREATE TABLE urls (
    id            BIGSERIAL PRIMARY KEY,
    code          VARCHAR(7) NOT NULL,
    long_url      TEXT NOT NULL,
    long_url_hash BYTEA NOT NULL,
    user_id       UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at    TIMESTAMPTZ,

    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    is_private    BOOLEAN NOT NULL DEFAULT FALSE,

    clicks        BIGINT NOT NULL DEFAULT 0,

    CONSTRAINT urls_code_unique UNIQUE (code)
);

CREATE UNIQUE INDEX idx_urls_code ON urls(code);
CREATE UNIQUE INDEX idx_urls_long_url_hash ON urls(long_url_hash);

CREATE INDEX idx_urls_active_code
ON urls(code)
WHERE is_active = TRUE;

CREATE INDEX idx_urls_expires_at
ON urls(expires_at)
WHERE expires_at IS NOT NULL;

CREATE INDEX idx_urls_user_id_created_at
ON urls(user_id, created_at DESC);

CREATE TABLE url_clicks (
    id            BIGSERIAL PRIMARY KEY,
    url_id        BIGINT NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    ip            INET,
    user_agent    TEXT,
    referer       TEXT,
    country       TEXT
);

CREATE INDEX idx_clicks_url_id ON url_clicks(url_id);
CREATE INDEX idx_clicks_created_at ON url_clicks(created_at);
