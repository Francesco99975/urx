-- name: CreateClick :one
INSERT INTO url_clicks (url_id, ip, ip_hash, user_agent, device, os, browser, referer, country, is_bot)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: CountClicksByURL :one
SELECT COUNT(*) AS total
FROM url_clicks
WHERE url_id = $1;

-- name: CountUserLinks :one
SELECT COUNT(*)
FROM urls
WHERE user_id = $1;

-- name: SumUserClicks :one
SELECT COALESCE(SUM(clicks), 0)
FROM urls
WHERE user_id = $1;

-- name: SumUserClicksToday :one
SELECT COUNT(*)
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE u.user_id = $1
  AND uc.created_at >= date_trunc('day', now());

-- name: SumUserClicksLast7Days :one
SELECT COUNT(*)
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE u.user_id = $1
  AND uc.created_at >= now() - interval '7 days';

-- name: GetClicksByURL :many
SELECT id, ip, user_agent, device, os, browser, referer, country, created_at
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

-- name: GetTopUserURLs :many
WITH current_week AS (
    SELECT
        uc.url_id,
        COUNT(*) AS clicks
    FROM url_clicks uc
    JOIN urls u ON u.id = uc.url_id
    WHERE u.user_id = $1
      AND uc.created_at >= now() - interval '7 days'
    GROUP BY uc.url_id
),
previous_week AS (
    SELECT
        uc.url_id,
        COUNT(*) AS clicks
    FROM url_clicks uc
    JOIN urls u ON u.id = uc.url_id
    WHERE u.user_id = $1
      AND uc.created_at >= now() - interval '14 days'
      AND uc.created_at < now() - interval '7 days'
    GROUP BY uc.url_id
)
SELECT
    u.id,
    u.code,
    u.long_url,
    u.clicks AS total_clicks,
    COALESCE(cw.clicks, 0) AS week_clicks,
    COALESCE(pw.clicks, 0) AS prev_week_clicks
FROM urls u
LEFT JOIN current_week cw ON cw.url_id = u.id
LEFT JOIN previous_week pw ON pw.url_id = u.id
WHERE u.user_id = $1
ORDER BY u.clicks DESC
LIMIT 5;

-- name: GetTopReferrers :many
SELECT
    COALESCE(NULLIF(uc.referer, ''), 'Direct') AS referer,
    COUNT(*) AS clicks
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE u.user_id = $1
GROUP BY referer
ORDER BY clicks DESC
LIMIT 5;

-- name: GetDeviceBreakdown :many
SELECT
    device,
    COUNT(*) AS clicks
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE u.user_id = $1
GROUP BY device
ORDER BY clicks DESC;

-- name: GetBrowserBreakdown :many
SELECT
    browser,
    COUNT(*) AS clicks
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE u.user_id = $1
GROUP BY browser
ORDER BY clicks DESC;

-- name: GetRecentUserClicks :many
SELECT
    uc.created_at,
    u.code,
    uc.user_agent,
    COALESCE(NULLIF(uc.referer, ''), 'Direct') AS referer
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE u.user_id = $1
ORDER BY uc.created_at DESC
LIMIT 20;

-- name: DeleteClicksBefore :exec
DELETE FROM url_clicks
WHERE created_at < $1;
