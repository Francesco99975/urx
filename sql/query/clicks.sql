-- name: CreateClick :one
INSERT INTO url_clicks (url_id, ip, ip_hash, user_agent, device, os, browser, referer, country, is_bot)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: CountClicksByURL :one
SELECT COUNT(*) AS total
FROM url_clicks
WHERE url_id = $1 AND is_bot = false;

-- name: CountUserLinks :one
SELECT COUNT(*)
FROM urls
WHERE user_id = $1;

-- name: SumUserClicks :one
SELECT COALESCE(total, 0) AS total_clicks
FROM (
    SELECT SUM(clicks) AS total
    FROM urls
    WHERE user_id = $1
) t;

-- name: SumUserClicksToday :one
SELECT COUNT(*)
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE u.user_id = $1
  AND uc.created_at >= date_trunc('day', now()) AND uc.is_bot = false;

-- name: SumUserClicksLast7Days :one
SELECT COUNT(*)
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE u.user_id = $1
  AND uc.created_at >= now() - interval '7 days' AND uc.is_bot = false;

-- name: GetClicksByURL :many
SELECT id, ip, user_agent, device, os, browser, referer, country, created_at
FROM url_clicks
WHERE url_id = $1 AND is_bot = false
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountClicksPerIP :many
SELECT ip, COUNT(*) AS total
FROM url_clicks
WHERE url_id = $1 AND is_bot = false
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
    WHERE u.user_id = $1 AND uc.is_bot = false
      AND uc.created_at >= now() - interval '7 days'
    GROUP BY uc.url_id
),
previous_week AS (
    SELECT
        uc.url_id,
        COUNT(*) AS clicks
    FROM url_clicks uc
    JOIN urls u ON u.id = uc.url_id
    WHERE u.user_id = $1 AND uc.is_bot = false
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
    referer,
    clicks
FROM (
    SELECT
        COALESCE(NULLIF(uc.referer, ''), 'Direct')::TEXT AS referer,
        COUNT(*)::BIGINT AS clicks
    FROM url_clicks uc
    JOIN urls u ON u.id = uc.url_id
    WHERE u.user_id = $1 AND uc.is_bot = false
    GROUP BY 1
) t
ORDER BY clicks DESC
LIMIT 5;

-- name: GetDeviceBreakdown :many
SELECT
    device,
    COUNT(*) AS clicks
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE u.user_id = $1 AND uc.is_bot = false
GROUP BY device
ORDER BY clicks DESC;

-- name: GetBrowserBreakdown :many
SELECT
    browser,
    COUNT(*) AS clicks
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE u.user_id = $1 AND uc.is_bot = false
GROUP BY browser
ORDER BY clicks DESC;

-- name: GetRecentUserClicks :many
SELECT
    created_at,
    code,
    browser,
    os,
    referer
FROM (
    SELECT
        uc.created_at,
        u.code,
        uc.browser,
        uc.os,
        COALESCE(NULLIF(uc.referer, ''), 'Direct')::TEXT AS referer
    FROM url_clicks uc
    JOIN urls u ON u.id = uc.url_id
    WHERE u.user_id = $1 AND uc.is_bot = false
) t
ORDER BY created_at DESC
LIMIT 20;

-- name: DeleteClicksBefore :exec
DELETE FROM url_clicks
WHERE created_at < $1;


-- name: CountPublicClicksByURL :one
SELECT COUNT(*) AS total
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE uc.url_id = $1
  AND u.user_id IS NULL AND uc.is_bot = false;

-- name: CountPublicLinks :one
SELECT COUNT(*)
FROM urls
WHERE user_id IS NULL;

-- name: SumPublicClicks :one
SELECT COALESCE(total, 0) AS total_clicks
FROM (
    SELECT SUM(clicks) AS total
    FROM urls
    WHERE user_id IS NULL
) t;


-- name: SumPublicClicksToday :one
SELECT COUNT(*)
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE u.user_id IS NULL
  AND uc.created_at >= date_trunc('day', now()) AND uc.is_bot = false;

-- name: SumPublicClicksLast7Days :one
SELECT COUNT(*)
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE u.user_id IS NULL
  AND uc.created_at >= now() - interval '7 days' AND uc.is_bot = false;

-- name: GetPublicClicksByURL :many
SELECT
    uc.id,
    uc.ip,
    uc.user_agent,
    uc.device,
    uc.os,
    uc.browser,
    uc.referer,
    uc.country,
    uc.created_at
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE uc.url_id = $1
  AND u.user_id IS NULL AND uc.is_bot = false
ORDER BY uc.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPublicClicksPerIP :many
SELECT uc.ip, COUNT(*) AS total
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE uc.url_id = $1
  AND u.user_id IS NULL AND uc.is_bot = false
GROUP BY uc.ip
ORDER BY total DESC
LIMIT $2;

-- name: GetTopPublicURLs :many
WITH current_week AS (
    SELECT
        uc.url_id,
        COUNT(*) AS clicks
    FROM url_clicks uc
    JOIN urls u ON u.id = uc.url_id
    WHERE u.user_id IS NULL
      AND uc.created_at >= now() - interval '7 days' AND uc.is_bot = false
    GROUP BY uc.url_id
),
previous_week AS (
    SELECT
        uc.url_id,
        COUNT(*) AS clicks
    FROM url_clicks uc
    JOIN urls u ON u.id = uc.url_id
    WHERE u.user_id IS NULL
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
WHERE u.user_id IS NULL
ORDER BY u.clicks DESC
LIMIT 5;

-- name: GetTopPublicReferrers :many
SELECT
    referer,
    clicks
FROM (
    SELECT
        COALESCE(NULLIF(uc.referer, ''), 'Direct')::TEXT AS referer,
        COUNT(*)::BIGINT AS clicks
    FROM url_clicks uc
    JOIN urls u ON u.id = uc.url_id
    WHERE u.user_id IS NULL AND uc.is_bot = false
    GROUP BY 1
) t
ORDER BY clicks DESC
LIMIT 5;

-- name: GetPublicDeviceBreakdown :many
SELECT
    uc.device,
    COUNT(*) AS clicks
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE u.user_id IS NULL AND uc.is_bot = false
GROUP BY uc.device
ORDER BY clicks DESC;

-- name: GetPublicBrowserBreakdown :many
SELECT
    uc.browser,
    COUNT(*) AS clicks
FROM url_clicks uc
JOIN urls u ON u.id = uc.url_id
WHERE u.user_id IS NULL AND uc.is_bot = false
GROUP BY uc.browser
ORDER BY clicks DESC;

-- name: GetRecentPublicClicks :many
SELECT
    created_at,
    code,
    browser,
    os,
    referer
FROM (
    SELECT
        uc.created_at,
        u.code,
        uc.browser,
        uc.os,
        COALESCE(NULLIF(uc.referer, ''), 'Direct')::TEXT AS referer
    FROM url_clicks uc
    JOIN urls u ON u.id = uc.url_id
    WHERE u.user_id IS NULL AND uc.is_bot = false
) t
ORDER BY created_at DESC
LIMIT 20;
