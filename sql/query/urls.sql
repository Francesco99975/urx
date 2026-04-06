-- name: CreateURL :one
INSERT INTO urls (code, long_url, long_url_hash, user_id, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCodeByHash :one
SELECT code
FROM urls
WHERE long_url_hash = $1
  AND is_active = TRUE
  AND (expires_at IS NULL OR expires_at > NOW())
LIMIT 1;

-- name: GetLongURLByHash :one
SELECT long_url
FROM urls
WHERE long_url_hash = $1
  AND is_active = TRUE
  AND (expires_at IS NULL OR expires_at > NOW())
LIMIT 1;

-- name: GetAndIncrementURL :one
UPDATE urls
SET clicks = clicks + 1
WHERE code = $1
  AND is_active = true
  AND (expires_at IS NULL OR expires_at > NOW())
RETURNING id, long_url, long_url_hash;

-- name: GetURLByCode :one
SELECT id, long_url, is_active, expires_at
FROM urls
WHERE code = $1
LIMIT 1;

-- name: GetActiveURLByCode :one
SELECT id, long_url
FROM urls
WHERE code = $1
  AND is_active = true
  AND (expires_at IS NULL OR expires_at > NOW())
LIMIT 1;

-- name: GetURLByID :one
SELECT *
FROM urls
WHERE id = $1
LIMIT 1;

-- name: GetURLsByUser :many
SELECT id, code, long_url, long_url_hash, clicks, created_at, expires_at, is_active
FROM urls
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: IncrementClicks :exec
UPDATE urls
SET clicks = clicks + 1
WHERE id = $1;

-- name: DeactivateURL :exec
UPDATE urls
SET is_active = false
WHERE id = $1;


