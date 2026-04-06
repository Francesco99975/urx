-- name: CreateClick :one
INSERT INTO url_clicks (url_id, ip, user_agent, referer)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CountClicksByURL :one
SELECT COUNT(*) AS total
FROM url_clicks
WHERE url_id = $1;

-- name: GetClicksByURL :many
SELECT id, ip, user_agent, referer, created_at
FROM url_clicks
WHERE url_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountClicksPerIP :many
SELECT ip, COUNT(*) AS total
FROM url_clicks
WHERE url_id = $1
GROUP BY ip
ORDER BY total DESC
LIMIT $2;

-- name: DeleteClicksBefore :exec
DELETE FROM url_clicks
WHERE created_at < $1;
